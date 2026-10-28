package commands

import (
	"strings"
	"testing"
)

func TestBatchEmbedsCount(t *testing.T) {
	var list []copiedMessage
	for n := 0; n < 23; n++ {
		list = append(list, copiedMessage{AuthorID: "1", AuthorName: "a", Text: "hi"})
	}
	b := batchEmbeds(list)
	if len(b) != 3 || len(b[0]) != 10 || len(b[1]) != 10 || len(b[2]) != 3 {
		t.Fatalf("got batch sizes %d", len(b))
	}
}

func TestBatchEmbedsLength(t *testing.T) {
	long := strings.Repeat("x", 2000)
	var list []copiedMessage
	for n := 0; n < 5; n++ {
		list = append(list, copiedMessage{AuthorID: "1", AuthorName: "a", Text: long})
	}
	for _, b := range batchEmbeds(list) {
		total := 0
		for _, e := range b {
			total += len(e.Description) + len(e.Author.Name) + len(e.Footer.Text)
		}
		if total > 6000 {
			t.Fatalf("batch has %d embed characters", total)
		}
	}
}

func TestEmbedPlaceholderWithoutText(t *testing.T) {
	if e := (copiedMessage{AuthorName: "a"}).embed(); !strings.Contains(e.Description, "no message text") {
		t.Fatalf("got %q", e.Description)
	}
}
