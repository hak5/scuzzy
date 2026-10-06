package commands

import (
	"strconv"
	"testing"
	"time"

	"github.com/foxtrot/scuzzy/models"
)

const testWindow = 30 * time.Second

func record(t *SpamTracker, author, content, channel string, at time.Time) (SpamVerdict, []models.TrackedInstance) {
	inst := models.TrackedInstance{ChannelID: channel, MessageID: strconv.FormatInt(at.UnixNano(), 10), SeenAt: at}
	verdict, _, hits := t.Record(author, content, inst, testWindow, 4, 3)
	return verdict, hits
}

func TestSameChannelRepeats(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	for i := 0; i < 3; i++ {
		if v, _ := record(tr, "u1", "buy nitro", "c1", now.Add(time.Duration(i)*time.Second)); v != SpamClean {
			t.Fatalf("flagged after %d messages", i+1)
		}
	}
	v, hits := record(tr, "u1", "buy nitro", "c1", now.Add(3*time.Second))
	if v != SpamFlagged || len(hits) != 4 {
		t.Fatalf("expected flag with 4 hits, got %v %d", v, len(hits))
	}
}

func TestCrossChannel(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	record(tr, "u1", "free steam", "c1", now)
	record(tr, "u1", "free steam", "c2", now.Add(time.Second))
	v, hits := record(tr, "u1", "free steam", "c3", now.Add(2*time.Second))
	if v != SpamFlagged || len(hits) != 3 {
		t.Fatalf("expected cross-channel flag, got %v %d", v, len(hits))
	}
}

// A raid bot blasting 10 channels should be flagged once, with every later copy cleaned up.
func TestStragglersAfterFlag(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	flagged, stragglers := 0, 0
	for i := 0; i < 10; i++ {
		v, _ := record(tr, "u1", "raid", "c"+strconv.Itoa(i), now.Add(time.Duration(i)*time.Millisecond))
		switch v {
		case SpamFlagged:
			flagged++
		case SpamStraggler:
			stragglers++
		}
	}
	if flagged != 1 || stragglers != 7 {
		t.Fatalf("expected 1 flag + 7 stragglers, got %d + %d", flagged, stragglers)
	}

	// Once the window passes with no copies, the user starts fresh.
	if v, _ := record(tr, "u1", "raid", "c0", now.Add(time.Minute)); v != SpamClean {
		t.Fatalf("expected clean after window, got %v", v)
	}
}

func TestDifferentAuthorsNotCombined(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	for i := 0; i < 10; i++ {
		if v, _ := record(tr, "u"+strconv.Itoa(i), "lol", "c1", now); v != SpamClean {
			t.Fatalf("different authors flagged")
		}
	}
}

func TestOutsideWindowNotFlagged(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	for i := 0; i < 10; i++ {
		if v, _ := record(tr, "u1", "hi", "c1", now.Add(time.Duration(i)*20*time.Second)); v != SpamClean {
			t.Fatalf("slow repeats flagged at %d", i)
		}
	}
}

func TestMarkBanned(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	if !tr.MarkBanned("u1", now, time.Minute) {
		t.Fatal("first post not reported")
	}
	if tr.MarkBanned("u1", now.Add(time.Second), time.Minute) {
		t.Fatal("duplicate post reported")
	}
	if !tr.MarkBanned("u1", now.Add(2*time.Minute), time.Minute) {
		t.Fatal("not reported after window")
	}
}

func TestSweep(t *testing.T) {
	tr := NewSpamTracker()
	now := time.Now()
	record(tr, "u1", "hi", "c1", now)
	for i := 0; i < 4; i++ {
		record(tr, "u2", "spam", "c1", now)
	}
	tr.MarkBanned("u3", now, time.Second)
	// A straggler extends u2's cleanup period to now+40s.
	record(tr, "u2", "spam", "c1", now.Add(10*time.Second))

	tr.Sweep(now.Add(testWindow), testWindow)
	if len(tr.messages) != 1 || len(tr.banned) != 0 {
		t.Fatalf("expected only the flagged entry to remain, got %d messages, %d bans", len(tr.messages), len(tr.banned))
	}
	tr.Sweep(now.Add(40*time.Second), testWindow)
	if len(tr.messages) != 0 {
		t.Fatalf("sweep left %d messages", len(tr.messages))
	}
}
