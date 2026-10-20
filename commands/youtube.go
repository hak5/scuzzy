package commands

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// The YouTube Data API key is read from the environment rather than the config file,
// so /getconfig can't print it.
const youtubeAPIKeyEnv = "SCUZZY_YOUTUBE_API_KEY"

// How long to remember posted videos after they drop out of the channel's feed.
const youtubeSeenRetention = 60 * 24 * time.Hour

type youtubeKind int

const (
	youtubeVideo youtubeKind = iota
	youtubeShort
	youtubeLive
	youtubeReplay
)

type youtubeEntry struct {
	ID    string
	Title string
}

// youtubeState is saved to disk so restarts don't repost videos.
type youtubeState struct {
	ChannelID string `json:"channel_id"`
	// Posted (or deliberately skipped) video IDs, with when they were last in the feed.
	Seen map[string]time.Time `json:"seen"`
	// Scheduled streams waiting to go live, by video ID, with their titles.
	Pending map[string]string `json:"pending"`
}

type youtubeWatcher struct {
	client    *http.Client
	feedURL   string
	shortsURL string
	apiURL    string
	apiKey    string
	statePath string
	state     *youtubeState
	post      func(kind youtubeKind, e youtubeEntry) error
}

func newYoutubeWatcher(statePath, apiKey string, post func(youtubeKind, youtubeEntry) error) *youtubeWatcher {
	return &youtubeWatcher{
		client: &http.Client{
			Timeout: 20 * time.Second,
			// /shorts/<id> redirects to /watch for normal videos; don't follow it
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		feedURL:   "https://www.youtube.com/feeds/videos.xml",
		shortsURL: "https://www.youtube.com/shorts/",
		apiURL:    "https://www.googleapis.com/youtube/v3/videos",
		apiKey:    apiKey,
		statePath: statePath,
		post:      post,
	}
}

func (c *Commands) youtubePollInterval() time.Duration {
	if c.Config.YoutubePollMinutes > 0 {
		return time.Duration(c.Config.YoutubePollMinutes) * time.Minute
	}
	return 5 * time.Minute
}

func (c *Commands) youtubeStatePath() string {
	if c.Config.YoutubeStateFile != "" {
		return c.Config.YoutubeStateFile
	}
	return filepath.Join(filepath.Dir(c.Config.ConfigPath), "youtube_seen.json")
}

// StartYoutubeFeed posts new uploads, shorts and live streams from a YouTube channel to the feed channel.
// It checks the config on every poll, so /reloadconfig can turn it on or off.
func (c *Commands) StartYoutubeFeed(s *discordgo.Session) {
	apiKey := os.Getenv(youtubeAPIKeyEnv)

	go func() {
		var w *youtubeWatcher
		for {
			if c.Config.YoutubeChannelID != "" && c.Config.YoutubeFeedChannel != "" {
				if w == nil || w.statePath != c.youtubeStatePath() {
					if apiKey == "" {
						log.Printf("[*] YouTube feed: %s is not set, so live streams will be posted as videos.\n", youtubeAPIKeyEnv)
					}
					w = newYoutubeWatcher(c.youtubeStatePath(), apiKey, func(kind youtubeKind, e youtubeEntry) error {
						return c.postYoutube(s, c.Config.YoutubeFeedChannel, kind, e)
					})
				}
				if err := w.poll(c.Config.YoutubeChannelID); err != nil {
					log.Println("[!] Error (YouTube Feed): " + err.Error())
				}
			}
			time.Sleep(c.youtubePollInterval())
		}
	}()
}

func (c *Commands) postYoutube(s *discordgo.Session, channelID string, kind youtubeKind, e youtubeEntry) error {
	link := "https://www.youtube.com/watch?v=" + e.ID
	var heading string
	switch kind {
	case youtubeShort:
		heading = "New Short"
		link = "https://www.youtube.com/shorts/" + e.ID
	case youtubeLive:
		heading = "🔴 Live now"
	case youtubeReplay:
		heading = "Live stream replay"
	default:
		heading = "New video"
	}

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "**" + heading + ":** " + e.Title + "\n" + link,
		// Never ping anyone, whatever the title says
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
	if err != nil {
		return err
	}
	log.Printf("[*] YouTube feed: posted %s %s\n", heading, e.ID)

	// Publish to servers following the channel if it's an Announcement channel
	if ch, err := s.State.Channel(channelID); err == nil && ch.Type == discordgo.ChannelTypeGuildNews {
		if _, err := s.ChannelMessageCrosspost(channelID, msg.ID); err != nil {
			log.Println("[!] Error (YouTube Feed Crosspost): " + err.Error())
		}
	}
	return nil
}

// poll checks the feed once and posts anything new.
func (w *youtubeWatcher) poll(channelID string) error {
	if err := w.loadState(); err != nil {
		return err
	}

	entries, err := w.fetchFeed(channelID)
	if err != nil {
		return err
	}

	now := time.Now()
	inFeed := map[string]bool{}
	var fresh []youtubeEntry
	for _, e := range entries {
		inFeed[e.ID] = true
		if _, ok := w.state.Seen[e.ID]; ok {
			w.state.Seen[e.ID] = now
			continue
		}
		if _, ok := w.state.Pending[e.ID]; ok {
			continue
		}
		fresh = append(fresh, e)
	}

	// First run for this channel: remember what's already there instead of posting it all.
	if w.state.ChannelID != channelID {
		w.state = &youtubeState{ChannelID: channelID, Seen: map[string]time.Time{}, Pending: map[string]string{}}
		for _, e := range entries {
			w.state.Seen[e.ID] = now
		}
		log.Printf("[*] YouTube feed: started watching %s (%d existing videos skipped)\n", channelID, len(entries))
		return w.saveState()
	}

	// The feed is newest first; post oldest first
	for l, r := 0, len(fresh)-1; l < r; l, r = l+1, r-1 {
		fresh[l], fresh[r] = fresh[r], fresh[l]
	}

	if w.apiKey != "" {
		err = w.classifyWithAPI(fresh, now)
	} else {
		for _, e := range fresh {
			w.postVideo(e, now)
		}
	}

	for id, t := range w.state.Seen {
		if !inFeed[id] && now.Sub(t) > youtubeSeenRetention {
			delete(w.state.Seen, id)
		}
	}

	if serr := w.saveState(); serr != nil {
		return serr
	}
	return err
}

// postVideo posts a finished upload as a video or short. Failures are left unmarked so the next poll retries.
func (w *youtubeWatcher) postVideo(e youtubeEntry, now time.Time) {
	kind := youtubeVideo
	short, err := w.isShort(e.ID)
	if err != nil {
		log.Println("[!] Error (YouTube Shorts Check): " + err.Error())
		return
	}
	if short {
		kind = youtubeShort
	}
	w.postAndMark(kind, e, now)
}

func (w *youtubeWatcher) postAndMark(kind youtubeKind, e youtubeEntry, now time.Time) {
	if err := w.post(kind, e); err != nil {
		log.Println("[!] Error (YouTube Feed Post): " + err.Error())
		return
	}
	delete(w.state.Pending, e.ID)
	w.state.Seen[e.ID] = now
}

// classifyWithAPI uses the YouTube Data API to tell live streams apart from uploads.
// Scheduled streams wait in Pending until they go live.
func (w *youtubeWatcher) classifyWithAPI(fresh []youtubeEntry, now time.Time) error {
	check := append([]youtubeEntry{}, fresh...)
	for id, title := range w.state.Pending {
		check = append(check, youtubeEntry{ID: id, Title: title})
	}
	if len(check) == 0 {
		return nil
	}

	details, err := w.fetchDetails(check)
	if err != nil {
		return err
	}

	for _, e := range check {
		d, ok := details[e.ID]
		if !ok {
			// Private, deleted or cancelled; don't post it
			delete(w.state.Pending, e.ID)
			w.state.Seen[e.ID] = now
			continue
		}
		if d.Title != "" {
			e.Title = d.Title
		}

		switch d.Status {
		case "upcoming":
			w.state.Pending[e.ID] = e.Title
		case "live":
			w.postAndMark(youtubeLive, e, now)
		default:
			if d.Stream {
				// A stream that finished between polls
				w.postAndMark(youtubeReplay, e, now)
			} else {
				w.postVideo(e, now)
			}
		}
	}
	return nil
}

type youtubeFeed struct {
	Entries []struct {
		VideoID string `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
		Title   string `xml:"title"`
	} `xml:"entry"`
}

func (w *youtubeWatcher) fetchFeed(channelID string) ([]youtubeEntry, error) {
	resp, err := w.client.Get(w.feedURL + "?channel_id=" + url.QueryEscape(channelID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("feed returned " + resp.Status)
	}

	var feed youtubeFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	var entries []youtubeEntry
	for _, e := range feed.Entries {
		if e.VideoID != "" {
			entries = append(entries, youtubeEntry{ID: e.VideoID, Title: e.Title})
		}
	}
	return entries, nil
}

func (w *youtubeWatcher) isShort(id string) (bool, error) {
	resp, err := w.client.Get(w.shortsURL + url.PathEscape(id))
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		return true, nil
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		return false, nil
	}
	return false, errors.New("shorts check returned " + resp.Status)
}

type youtubeDetails struct {
	Title  string
	Status string // none, upcoming or live
	Stream bool
}

func (w *youtubeWatcher) fetchDetails(entries []youtubeEntry) (map[string]youtubeDetails, error) {
	var ids []string
	for _, e := range entries {
		ids = append(ids, e.ID)
	}

	out := map[string]youtubeDetails{}
	// The API takes up to 50 IDs per request
	for len(ids) > 0 {
		n := len(ids)
		if n > 50 {
			n = 50
		}
		q := url.Values{
			"part": {"snippet,liveStreamingDetails"},
			"id":   {strings.Join(ids[:n], ",")},
			"key":  {w.apiKey},
		}
		ids = ids[n:]

		resp, err := w.client.Get(w.apiURL + "?" + q.Encode())
		if err != nil {
			// Don't log the URL, it contains the key
			return nil, errors.New("YouTube API request failed")
		}
		var body struct {
			Items []struct {
				ID      string `json:"id"`
				Snippet struct {
					Title                string `json:"title"`
					LiveBroadcastContent string `json:"liveBroadcastContent"`
				} `json:"snippet"`
				LiveStreamingDetails *struct{} `json:"liveStreamingDetails"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("YouTube API returned " + resp.Status)
		}
		if err != nil {
			return nil, err
		}

		for _, it := range body.Items {
			out[it.ID] = youtubeDetails{
				Title:  it.Snippet.Title,
				Status: it.Snippet.LiveBroadcastContent,
				Stream: it.LiveStreamingDetails != nil,
			}
		}
	}
	return out, nil
}

func (w *youtubeWatcher) loadState() error {
	if w.state != nil {
		return nil
	}
	w.state = &youtubeState{}

	buf, err := ioutil.ReadFile(w.statePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, w.state); err != nil {
		return err
	}
	if w.state.Seen == nil {
		w.state.Seen = map[string]time.Time{}
	}
	if w.state.Pending == nil {
		w.state.Pending = map[string]string{}
	}
	return nil
}

func (w *youtubeWatcher) saveState() error {
	buf, err := json.MarshalIndent(w.state, "", "  ")
	if err != nil {
		return err
	}
	// Write then rename so a crash can't leave a half-written file
	tmp := w.statePath + ".tmp"
	if err := ioutil.WriteFile(tmp, buf, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, w.statePath)
}
