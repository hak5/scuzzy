package models

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

type CachedMessage struct {
	ID        string
	CreatedAt time.Time
	Count     int
}

type TrackedMessage struct {
	Author         *discordgo.User `json:"author"`
	Instances      map[string]CachedMessage
	MessageContent string
}
