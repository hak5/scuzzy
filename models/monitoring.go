package models

import (
	"time"
)

// TrackedInstance is a single sighting of a message fingerprint.
type TrackedInstance struct {
	ChannelID string
	MessageID string
	SeenAt    time.Time
}

// TrackedMessage holds recent sightings of the same content from the same author.
// Fingerprint is a keyed hash of the content; the content itself is never stored.
type TrackedMessage struct {
	AuthorID    string
	Fingerprint string
	Instances   []TrackedInstance
	// Further copies seen before this time are leftovers from a burst that was
	// already actioned; they're cleaned up without another report.
	ActionedUntil time.Time
}
