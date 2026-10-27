package commands

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	maxEmbedDescriptionLength = 4096
	// Discord caps a message at 10 embeds and 6000 embed characters in total
	maxEmbedsPerMessage = 10
	maxEmbedCharsPerMsg = 5800
)

// copiedMessage is the part of a message that's reposted when copying or moving it.
type copiedMessage struct {
	ChannelID  string
	AuthorID   string
	AuthorName string
	AuthorIcon string
	Text       string
	Timestamp  time.Time
}

func newCopiedMessage(m *discordgo.Message) copiedMessage {
	cm := copiedMessage{ChannelID: m.ChannelID, Text: copyText(m), Timestamp: m.Timestamp}
	if m.Author != nil {
		cm.AuthorID = m.Author.ID
		cm.AuthorName = m.Author.Username
		cm.AuthorIcon = m.Author.AvatarURL("64")
	}
	return cm
}

// copyText is the text reposted for a message: its content plus attachment links.
func copyText(m *discordgo.Message) string {
	text := m.Content
	for _, a := range m.Attachments {
		text += "\n" + a.URL
	}
	return strings.TrimSpace(text)
}

func (cm copiedMessage) embed() *discordgo.MessageEmbed {
	text := cm.Text
	if text == "" {
		// Without the Message Content intent Discord sends other users' messages without text
		text = "*(no message text available)*"
	}
	if r := []rune(text); len(r) > maxEmbedDescriptionLength {
		text = string(r[:maxEmbedDescriptionLength-1]) + "…"
	}
	e := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: cm.AuthorName, IconURL: cm.AuthorIcon},
		Description: text,
	}
	if !cm.Timestamp.IsZero() {
		e.Timestamp = cm.Timestamp.Format(time.RFC3339)
	}
	return e
}

// lookupMessage finds a message in the state cache, falling back to the API.
func lookupMessage(s *discordgo.Session, channelID, messageID string) *discordgo.Message {
	if m, err := s.State.Message(channelID, messageID); err == nil {
		return m
	}
	m, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		return nil
	}
	return m
}

// cachedMessagesBy returns every cached message from a user, for deletions (like bans) that
// remove messages the bot hasn't been handed directly.
func cachedMessagesBy(s *discordgo.Session, guildID, userID string) []*discordgo.Message {
	g, err := s.State.Guild(guildID)
	if err != nil {
		return nil
	}

	s.State.RLock()
	defer s.State.RUnlock()

	var out []*discordgo.Message
	for _, list := range [][]*discordgo.Channel{g.Channels, g.Threads} {
		for _, ch := range list {
			for _, m := range ch.Messages {
				if m.Author != nil && m.Author.ID == userID {
					out = append(out, m)
				}
			}
		}
	}
	return out
}

// copyToModerators reposts messages to the moderator channel before Scuzzy deletes them,
// in the same format as the "Copy message" command. Errors are logged; the caller still deletes.
func (c *Commands) copyToModerators(s *discordgo.Session, reason string, msgs []copiedMessage) {
	if len(msgs) == 0 {
		return
	}

	byChannel := map[string][]copiedMessage{}
	var channels []string
	for _, m := range msgs {
		if _, ok := byChannel[m.ChannelID]; !ok {
			channels = append(channels, m.ChannelID)
		}
		byChannel[m.ChannelID] = append(byChannel[m.ChannelID], m)
	}

	for _, channelID := range channels {
		list := byChannel[channelID]
		sort.SliceStable(list, func(a, b int) bool { return list[a].Timestamp.Before(list[b].Timestamp) })

		if len(list) == 1 {
			m := list[0]
			c.sendModeratorCopy(s, "Scuzzy deleted this message from <#"+channelID+"> ("+reason+")\n\nOriginal message from <@"+m.AuthorID+">:", []*discordgo.MessageEmbed{m.embed()})
			continue
		}

		// Many messages are packed several to a post so a purge doesn't flood the channel
		batches := batchEmbeds(list)

		header := "Scuzzy deleted " + strconv.Itoa(len(list)) + " messages from <#" + channelID + "> (" + reason + "):"
		for idx, b := range batches {
			h := header
			if len(batches) > 1 {
				h += " part " + strconv.Itoa(idx+1) + "/" + strconv.Itoa(len(batches))
			}
			c.sendModeratorCopy(s, h, b)
		}
	}
}

func (c *Commands) sendModeratorCopy(s *discordgo.Session, header string, embeds []*discordgo.MessageEmbed) {
	_, err := s.ChannelMessageSendComplex(c.moderatorChannel(), &discordgo.MessageSend{
		Content: header,
		Embeds:  embeds,
		// Mentions show who was involved without pinging anyone
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
	if err != nil {
		log.Println("[!] Error (Copy Deleted Message): " + err.Error())
	}
}

// batchEmbeds packs messages into as few posts as Discord's embed limits allow.
func batchEmbeds(list []copiedMessage) [][]*discordgo.MessageEmbed {
	var batches [][]*discordgo.MessageEmbed
	var batch []*discordgo.MessageEmbed
	chars := 0
	for _, m := range list {
		e := m.embed()
		e.Footer = &discordgo.MessageEmbedFooter{Text: "User ID: " + m.AuthorID}
		n := len([]rune(e.Description)) + len([]rune(e.Author.Name)) + len(e.Footer.Text)
		if len(batch) == maxEmbedsPerMessage || (len(batch) > 0 && chars+n > maxEmbedCharsPerMsg) {
			batches = append(batches, batch)
			batch, chars = nil, 0
		}
		batch = append(batch, e)
		chars += n
	}
	return append(batches, batch)
}
