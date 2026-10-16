package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
	"github.com/foxtrot/scuzzy/permissions"
)

type ScuzzyHandler func(session *discordgo.Session, i *discordgo.InteractionCreate) error

type ScuzzyCommand struct {
	Index       int
	Name        string
	Description string
	AdminOnly   bool
	// Ephemeral replies are only visible to the user who ran the command.
	Ephemeral bool
	// Type defaults to a slash command. Message commands appear under Apps when right-clicking a message.
	Type    discordgo.ApplicationCommandType
	Options []*discordgo.ApplicationCommandOption
	Handler ScuzzyHandler
}

type Commands struct {
	Token                 string
	Permissions           *permissions.Permissions
	Config                *models.Configuration
	ScuzzyCommands        map[string]ScuzzyCommand
	ScuzzyCommandsByIndex map[int]ScuzzyCommand

	spamTracker *SpamTracker
}
