package commands

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
)

const (
	defaultSpamWindow                = 30 * time.Second
	defaultSpamDuplicateThreshold    = 4
	defaultSpamCrossChannelThreshold = 3
	defaultSpamTimeout               = 60 * time.Minute
	defaultBurstWindow               = 10 * time.Second
	defaultBurstChannelThreshold     = 3
	spamSweepInterval                = time.Minute
	maxReportContentLength           = 1000
	freeBansDeleteMessageDays        = 7 // Discord's maximum
	freeBansDedupeWindow             = 5 * time.Minute
	maxEmbedFieldLength              = 1024
)

// SpamVerdict is the outcome of recording a message.
type SpamVerdict int

const (
	SpamClean     SpamVerdict = iota
	SpamFlagged               // crossed a threshold; act on it and report
	SpamStraggler             // another copy from a burst already flagged; just clean it up
)

// SpamTracker records recent messages per author+fingerprint so repeats can be detected.
// Only keyed hashes of message content are kept, never the content itself.
type SpamTracker struct {
	mu       sync.Mutex
	key      []byte // random per process, so fingerprints can't be reversed by guessing or outlive a restart
	messages map[string]*models.TrackedMessage
	banned   map[string]time.Time // users already handled by the free-bans channel
}

func NewSpamTracker() *SpamTracker {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal("[!] Error (Spam Tracker Key): " + err.Error())
	}
	return &SpamTracker{
		key:      key,
		messages: make(map[string]*models.TrackedMessage),
		banned:   make(map[string]time.Time),
	}
}

// Fingerprint returns an HMAC-SHA256 of the message text, used in place of the text itself.
func (t *SpamTracker) Fingerprint(text string) string {
	mac := hmac.New(sha256.New, t.key)
	mac.Write([]byte(text))
	return string(mac.Sum(nil))
}

// sighting adds inst to the entry for key and returns it, or returns nil when the
// key was already actioned within the window (the sighting is a straggler).
// The caller must hold t.mu.
func (t *SpamTracker) sighting(key, authorID, fingerprint string, inst models.TrackedInstance, window time.Duration) *models.TrackedMessage {
	tm, ok := t.messages[key]
	if !ok {
		tm = &models.TrackedMessage{AuthorID: authorID, Fingerprint: fingerprint}
		t.messages[key] = tm
	}

	if inst.SeenAt.Before(tm.ActionedUntil) {
		tm.ActionedUntil = inst.SeenAt.Add(window)
		return nil
	}

	tm.Instances = pruneInstances(tm.Instances, inst.SeenAt.Add(-window))
	tm.Instances = append(tm.Instances, inst)
	return tm
}

// flag hands back the sightings to act on and starts the cleanup period for stragglers.
func flag(tm *models.TrackedMessage, inst models.TrackedInstance, window time.Duration) []models.TrackedInstance {
	hits := tm.Instances
	tm.Instances = nil
	tm.ActionedUntil = inst.SeenAt.Add(window)
	return hits
}

func countChannels(instances []models.TrackedInstance) map[string]int {
	channels := map[string]int{}
	for _, i := range instances {
		channels[i.ChannelID]++
	}
	return channels
}

// Record adds a sighting and reports whether it crosses either spam threshold.
// Once flagged, further copies within the window are returned as stragglers so
// they can be deleted without another report.
func (t *SpamTracker) Record(authorID, fingerprint string, inst models.TrackedInstance, window time.Duration, dupThreshold, crossThreshold int) (SpamVerdict, string, []models.TrackedInstance) {
	t.mu.Lock()
	defer t.mu.Unlock()

	tm := t.sighting(authorID+"\x00"+fingerprint, authorID, fingerprint, inst, window)
	if tm == nil {
		return SpamStraggler, "", []models.TrackedInstance{inst}
	}

	channels := countChannels(tm.Instances)
	sameChannel := channels[inst.ChannelID]

	var reason string
	switch {
	case sameChannel >= dupThreshold:
		reason = fmt.Sprintf("Same message sent %d times in <#%s> within %s", sameChannel, inst.ChannelID, window)
	case len(channels) >= crossThreshold:
		reason = fmt.Sprintf("Same message sent in %d channels within %s", len(channels), window)
	default:
		return SpamClean, "", nil
	}

	return SpamFlagged, reason, flag(tm, inst, window)
}

// RecordBurst tracks every message from an author, whatever it says, and flags them once
// they span channelThreshold channels within the window. It only uses channel IDs and
// timestamps, so it works without the Message Content intent.
func (t *SpamTracker) RecordBurst(authorID string, inst models.TrackedInstance, window time.Duration, channelThreshold int) (SpamVerdict, string, []models.TrackedInstance) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Content entries are keyed "<author id>\x00<hash>"; author IDs are numeric so this can't collide.
	tm := t.sighting("burst\x00"+authorID, authorID, "", inst, window)
	if tm == nil {
		return SpamStraggler, "", []models.TrackedInstance{inst}
	}

	channels := countChannels(tm.Instances)
	if len(channels) < channelThreshold {
		return SpamClean, "", nil
	}

	reason := fmt.Sprintf("Posted in %d channels within %s", len(channels), window)
	return SpamFlagged, reason, flag(tm, inst, window)
}

// MarkBanned returns true the first time a user is seen within the window.
func (t *SpamTracker) MarkBanned(userID string, now time.Time, window time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if until, ok := t.banned[userID]; ok && now.Before(until) {
		return false
	}
	t.banned[userID] = now.Add(window)
	return true
}

// Sweep drops entries with no recent sightings and no active cleanup period.
func (t *SpamTracker) Sweep(now time.Time, window time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for key, tm := range t.messages {
		tm.Instances = pruneInstances(tm.Instances, now.Add(-window))
		if len(tm.Instances) == 0 && !now.Before(tm.ActionedUntil) {
			delete(t.messages, key)
		}
	}
	for userID, until := range t.banned {
		if !now.Before(until) {
			delete(t.banned, userID)
		}
	}
}

func pruneInstances(instances []models.TrackedInstance, cutoff time.Time) []models.TrackedInstance {
	kept := instances[:0]
	for _, i := range instances {
		if i.SeenAt.After(cutoff) {
			kept = append(kept, i)
		}
	}
	return kept
}

// messageText identifies "the same message", including attachment-only posts.
// Returns an empty string for messages that shouldn't be tracked.
func messageText(m *discordgo.Message) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(m.Content))
	for _, a := range m.Attachments {
		sb.WriteString("\n[attachment] " + a.Filename + " " + strconv.Itoa(a.Size))
	}
	return sb.String()
}

func (c *Commands) startAutoMod() {
	c.spamTracker = NewSpamTracker()
	go func() {
		for range time.Tick(spamSweepInterval) {
			// Sweep with the longer window so neither detector loses sightings it still needs.
			window := c.spamWindow()
			if bw := c.burstWindow(); bw > window {
				window = bw
			}
			c.spamTracker.Sweep(time.Now(), window)
		}
	}()
}

func (c *Commands) spamWindow() time.Duration {
	if c.Config.SpamWindowSeconds > 0 {
		return time.Duration(c.Config.SpamWindowSeconds) * time.Second
	}
	return defaultSpamWindow
}

func (c *Commands) spamDuplicateThreshold() int {
	if c.Config.SpamDuplicateThreshold > 0 {
		return c.Config.SpamDuplicateThreshold
	}
	return defaultSpamDuplicateThreshold
}

func (c *Commands) spamCrossChannelThreshold() int {
	if c.Config.SpamCrossChannelThreshold > 0 {
		return c.Config.SpamCrossChannelThreshold
	}
	return defaultSpamCrossChannelThreshold
}

func (c *Commands) burstWindow() time.Duration {
	if c.Config.BurstWindowSeconds > 0 {
		return time.Duration(c.Config.BurstWindowSeconds) * time.Second
	}
	return defaultBurstWindow
}

func (c *Commands) burstChannelThreshold() int {
	if c.Config.BurstChannelThreshold > 0 {
		return c.Config.BurstChannelThreshold
	}
	return defaultBurstChannelThreshold
}

func (c *Commands) spamTimeout() time.Duration {
	if c.Config.SpamTimeoutMinutes > 0 {
		return time.Duration(c.Config.SpamTimeoutMinutes) * time.Minute
	}
	return defaultSpamTimeout
}

func (c *Commands) moderatorChannel() string {
	if c.Config.ModeratorChannel != "" {
		return c.Config.ModeratorChannel
	}
	return c.Config.LoggingChannel
}

// moderatorRoleIDs returns the role(s) to ping: moderator_role_id if set, otherwise the admin roles.
func (c *Commands) moderatorRoleIDs() []string {
	if c.Config.ModeratorRoleID != "" {
		return []string{c.Config.ModeratorRoleID}
	}
	var ids []string
	for _, r := range c.Permissions.AdminRoles {
		ids = append(ids, r.ID)
	}
	return ids
}

// isAutoModExempt skips bots, webhooks, DMs, other guilds and staff.
func (c *Commands) isAutoModExempt(s *discordgo.Session, m *discordgo.Message) bool {
	if m.Author == nil || m.Author.Bot || m.WebhookID != "" {
		return true
	}
	if s.State.User != nil && m.Author.ID == s.State.User.ID {
		return true
	}
	if m.GuildID == "" || m.GuildID != c.Config.GuildID {
		return true
	}
	if m.Member != nil && c.Permissions.CheckAdminRole(m.Member) {
		return true
	}
	return false
}

// ProcessAutoMod runs the honeypot and spam checks. Returns true if the
// message was actioned and shouldn't be processed further.
func (c *Commands) ProcessAutoMod(s *discordgo.Session, m *discordgo.Message) bool {
	if c.isAutoModExempt(s, m) {
		return false
	}
	if c.processFreeBans(s, m) {
		return true
	}
	if c.processSpam(s, m) {
		return true
	}
	return c.processBurst(s, m)
}

func (c *Commands) processFreeBans(s *discordgo.Session, m *discordgo.Message) bool {
	if c.Config.FreeBansChannel == "" || m.ChannelID != c.Config.FreeBansChannel {
		return false
	}

	// Extra messages sent before the ban lands are removed by the ban itself.
	if !c.spamTracker.MarkBanned(m.Author.ID, time.Now(), freeBansDedupeWindow) {
		return true
	}

	reason := "[AUTO] Posted in honeypot channel"
	action := fmt.Sprintf("Ban + delete last %d days of messages", freeBansDeleteMessageDays)
	result := ""
	if c.Config.FreeBansEnforce {
		err := s.GuildBanCreateWithReason(c.Config.GuildID, m.Author.ID, reason, freeBansDeleteMessageDays)
		if err != nil {
			result = "Ban failed: " + err.Error()
		}
	}

	c.reportAutoMod(s, "Free Bans", reason, action, c.Config.FreeBansEnforce, false, result, m.Author, []string{m.ChannelID}, messageText(m))
	return true
}

func (c *Commands) processSpam(s *discordgo.Session, m *discordgo.Message) bool {
	if !c.Config.SpamDetection || c.spamTracker == nil {
		return false
	}

	text := messageText(m)
	if text == "" {
		return false
	}
	fingerprint := c.spamTracker.Fingerprint(text)

	inst := models.TrackedInstance{ChannelID: m.ChannelID, MessageID: m.ID, SeenAt: time.Now()}
	verdict, reason, hits := c.spamTracker.Record(m.Author.ID, fingerprint, inst, c.spamWindow(), c.spamDuplicateThreshold(), c.spamCrossChannelThreshold())
	return c.actionSpam(s, m, "Spam Detected", verdict, reason, hits, c.Config.SpamEnforce, text)
}

// processBurst catches accounts posting across many channels in a short time, without reading message text.
func (c *Commands) processBurst(s *discordgo.Session, m *discordgo.Message) bool {
	if !c.Config.BurstDetection || c.spamTracker == nil {
		return false
	}

	inst := models.TrackedInstance{ChannelID: m.ChannelID, MessageID: m.ID, SeenAt: time.Now()}
	verdict, reason, hits := c.spamTracker.RecordBurst(m.Author.ID, inst, c.burstWindow(), c.burstChannelThreshold())
	return c.actionSpam(s, m, "Channel Burst Detected", verdict, reason, hits, c.Config.BurstEnforce, messageText(m))
}

// actionSpam deletes flagged messages, times the author out and reports it (or, in dry-run
// mode, only reports what it would have done). Returns true if the message was handled.
func (c *Commands) actionSpam(s *discordgo.Session, m *discordgo.Message, title string, verdict SpamVerdict, reason string, hits []models.TrackedInstance, enforce bool, text string) bool {
	switch verdict {
	case SpamClean:
		return false
	case SpamStraggler:
		if enforce {
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				log.Println("[!] Error (AutoMod Straggler Delete): " + err.Error())
			}
		}
		return true
	}

	byChannel := map[string][]string{}
	var channels []string
	for _, h := range hits {
		if _, ok := byChannel[h.ChannelID]; !ok {
			channels = append(channels, h.ChannelID)
		}
		byChannel[h.ChannelID] = append(byChannel[h.ChannelID], h.MessageID)
	}

	timeout := c.spamTimeout()
	action := fmt.Sprintf("Delete %d message(s) + timeout for %s", len(hits), timeout)
	var errs []string
	if enforce {
		for channelID, ids := range byChannel {
			if err := deleteMessages(s, channelID, ids); err != nil {
				errs = append(errs, "Delete in <#"+channelID+"> failed: "+err.Error())
			}
		}
		until := time.Now().Add(timeout)
		if err := s.GuildMemberTimeout(c.Config.GuildID, m.Author.ID, &until, discordgo.WithAuditLogReason("[AUTO] Spam: "+reason)); err != nil {
			errs = append(errs, "Timeout failed: "+err.Error())
		}
	}

	// Ping moderators once the user is actually timed out so they can decide on a ban.
	c.reportAutoMod(s, title, reason, action, enforce, enforce, strings.Join(errs, "\n"), m.Author, channels, text)
	return true
}

func deleteMessages(s *discordgo.Session, channelID string, ids []string) error {
	if len(ids) == 1 {
		return s.ChannelMessageDelete(channelID, ids[0])
	}
	for start := 0; start < len(ids); start += 100 {
		end := start + 100
		if end > len(ids) {
			end = len(ids)
		}
		if err := s.ChannelMessagesBulkDelete(channelID, ids[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// reportAutoMod posts what was done (or, in dry-run mode, what would have been done) to the moderator channel.
func (c *Commands) reportAutoMod(s *discordgo.Session, title, reason, action string, enforced, pingModerators bool, errText string, user *discordgo.User, channels []string, content string) {
	status := "success"
	if !enforced {
		title = "[DRY RUN] " + title
		action = "Would have: " + action
		status = "warning"
	} else if errText != "" {
		status = "error"
	}

	var chanMentions []string
	for _, ch := range channels {
		chanMentions = append(chanMentions, "<#"+ch+">")
	}

	if r := []rune(content); len(r) > maxReportContentLength {
		content = string(r[:maxReportContentLength]) + "…"
	}
	content = strings.ReplaceAll(content, "```", "'''")
	if r := []rune(errText); len(r) > maxEmbedFieldLength {
		errText = string(r[:maxEmbedFieldLength-1]) + "…"
	}

	embed := c.CreateDefinedEmbed(title, reason, status, nil)
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "User", Value: fmt.Sprintf("<@%s> (%s, `%s`)", user.ID, user.Username, user.ID)},
		{Name: "Channels", Value: strings.Join(chanMentions, " ")},
		{Name: "Action", Value: action},
	}
	if content != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Triggering Message", Value: "```\n" + content + "\n```"})
	}
	if errText != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Errors", Value: errText})
	}

	log.Printf("[*] AutoMod: %s - %s (%s) - %s", title, user.ID, reason, action)
	msg := &discordgo.MessageSend{
		Embeds:          []*discordgo.MessageEmbed{embed},
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}
	if pingModerators {
		roles := c.moderatorRoleIDs()
		var mentions []string
		for _, id := range roles {
			mentions = append(mentions, "<@&"+id+">")
		}
		msg.Content = strings.Join(mentions, " ") + " user was auto-timed out for spam. Review and `/ban user:" + user.ID + "` if warranted."
		msg.AllowedMentions.Roles = roles
	}

	if _, err := s.ChannelMessageSendComplex(c.moderatorChannel(), msg); err != nil {
		log.Println("[!] Error (AutoMod Report): " + err.Error())
	}
}
