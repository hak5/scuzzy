package commands

import (
	"github.com/bwmarrin/discordgo"
)

func (c *Commands) handleSetStatus(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	st := options(i)["status"].StringValue()

	err := s.UpdateGameStatus(0, st)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Set Status", "Operation completed successfully.", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}
