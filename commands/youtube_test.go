package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// fakeYoutube serves a feed, the shorts redirect check and the videos API.
type fakeYoutube struct {
	feed   []youtubeEntry // newest first, like the real feed
	shorts map[string]bool
	status map[string]string // liveBroadcastContent per video; missing = not returned
	stream map[string]bool
}

func (f *fakeYoutube) server(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/feed", func(w http.ResponseWriter, r *http.Request) {
		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0"?><feed xmlns:yt="http://www.youtube.com/xml/schemas/2015" xmlns="http://www.w3.org/2005/Atom">`)
		for _, e := range f.feed {
			sb.WriteString("<entry><yt:videoId>" + e.ID + "</yt:videoId><title>" + e.Title + "</title></entry>")
		}
		sb.WriteString("</feed>")
		w.Write([]byte(sb.String()))
	})
	mux.HandleFunc("/shorts/", func(w http.ResponseWriter, r *http.Request) {
		if f.shorts[strings.TrimPrefix(r.URL.Path, "/shorts/")] {
			return
		}
		http.Redirect(w, r, "/watch", http.StatusSeeOther)
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "k" {
			t.Errorf("API called without key")
		}
		var items []map[string]interface{}
		for _, id := range strings.Split(r.URL.Query().Get("id"), ",") {
			st, ok := f.status[id]
			if !ok {
				continue
			}
			it := map[string]interface{}{"id": id, "snippet": map[string]string{"title": "api " + id, "liveBroadcastContent": st}}
			if f.stream[id] {
				it["liveStreamingDetails"] = map[string]string{}
			}
			items = append(items, it)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
	})
	return httptest.NewServer(mux)
}

type postedVideo struct {
	kind youtubeKind
	id   string
}

func newTestWatcher(t *testing.T, srv *httptest.Server, apiKey, statePath string, posted *[]postedVideo) *youtubeWatcher {
	w := newYoutubeWatcher(statePath, apiKey, func(k youtubeKind, e youtubeEntry) error {
		*posted = append(*posted, postedVideo{k, e.ID})
		return nil
	})
	w.feedURL = srv.URL + "/feed"
	w.shortsURL = srv.URL + "/shorts/"
	w.apiURL = srv.URL + "/api"
	return w
}

func TestYoutubeFirstRunSkipsExisting(t *testing.T) {
	f := &fakeYoutube{feed: []youtubeEntry{{"b", "B"}, {"a", "A"}}}
	srv := f.server(t)
	defer srv.Close()

	var posted []postedVideo
	w := newTestWatcher(t, srv, "", filepath.Join(t.TempDir(), "state.json"), &posted)
	if err := w.poll("UC1"); err != nil {
		t.Fatal(err)
	}
	if len(posted) != 0 {
		t.Fatalf("first run posted %v", posted)
	}
}

func TestYoutubePostsNewVideosAndShorts(t *testing.T) {
	f := &fakeYoutube{feed: []youtubeEntry{{"a", "A"}}, shorts: map[string]bool{"s": true}}
	srv := f.server(t)
	defer srv.Close()

	state := filepath.Join(t.TempDir(), "state.json")
	var posted []postedVideo
	w := newTestWatcher(t, srv, "", state, &posted)
	w.poll("UC1")

	f.feed = []youtubeEntry{{"s", "S"}, {"v", "V"}, {"a", "A"}}
	if err := w.poll("UC1"); err != nil {
		t.Fatal(err)
	}
	want := []postedVideo{{youtubeVideo, "v"}, {youtubeShort, "s"}}
	if len(posted) != 2 || posted[0] != want[0] || posted[1] != want[1] {
		t.Fatalf("posted %v, want %v (oldest first)", posted, want)
	}

	// A restarted bot reads the saved state and doesn't repost
	posted = nil
	w = newTestWatcher(t, srv, "", state, &posted)
	if err := w.poll("UC1"); err != nil {
		t.Fatal(err)
	}
	if len(posted) != 0 {
		t.Fatalf("reposted after restart: %v", posted)
	}
}

func TestYoutubeChangedChannelReseeds(t *testing.T) {
	f := &fakeYoutube{feed: []youtubeEntry{{"a", "A"}}}
	srv := f.server(t)
	defer srv.Close()

	var posted []postedVideo
	w := newTestWatcher(t, srv, "", filepath.Join(t.TempDir(), "state.json"), &posted)
	w.poll("UC1")
	f.feed = []youtubeEntry{{"x", "X"}, {"a", "A"}}
	w.poll("UC2")
	if len(posted) != 0 {
		t.Fatalf("switching channels posted %v", posted)
	}
}

func TestYoutubeLiveStreams(t *testing.T) {
	f := &fakeYoutube{
		feed:   []youtubeEntry{{"a", "A"}},
		status: map[string]string{},
		stream: map[string]bool{"up": true, "done": true},
	}
	srv := f.server(t)
	defer srv.Close()

	var posted []postedVideo
	w := newTestWatcher(t, srv, "k", filepath.Join(t.TempDir(), "state.json"), &posted)
	w.poll("UC1")

	// A scheduled stream waits; a stream that already ended posts as a replay; a removed video is skipped
	f.feed = []youtubeEntry{{"gone", "G"}, {"done", "D"}, {"up", "U"}, {"a", "A"}}
	f.status["up"] = "upcoming"
	f.status["done"] = "none"
	if err := w.poll("UC1"); err != nil {
		t.Fatal(err)
	}
	if len(posted) != 1 || posted[0] != (postedVideo{youtubeReplay, "done"}) {
		t.Fatalf("posted %v", posted)
	}

	// Still scheduled: nothing new
	w.poll("UC1")
	if len(posted) != 1 {
		t.Fatalf("posted %v while stream was upcoming", posted)
	}

	// Goes live, even after dropping out of the feed
	f.feed = []youtubeEntry{{"a", "A"}}
	f.status["up"] = "live"
	w.poll("UC1")
	if len(posted) != 2 || posted[1] != (postedVideo{youtubeLive, "up"}) {
		t.Fatalf("posted %v", posted)
	}

	// Ends: not posted again
	f.status["up"] = "none"
	w.poll("UC1")
	if len(posted) != 2 {
		t.Fatalf("posted %v after stream ended", posted)
	}
}
