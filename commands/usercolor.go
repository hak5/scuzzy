package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

func (c *Commands) handleUserColors(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	msgC := "You can choose from the following colors:\n\n"
	for _, v := range c.Config.ColorRoles {
		msgC += "<@&" + v.ID + ">\n"
	}
	msgC += "\n\nUse `/color <color>` to set.\n"
	msgC += "Example: `/color red`.\n"

	msg := c.CreateDefinedEmbed("User Colors", msgC, "", i.Member.User)

	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleUserColor(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var err error

	rUserID := i.Member.User.ID
	roleColorName := options(i)["color"].StringValue()

	roleColorID := ""
	for _, role := range c.Config.ColorRoles {
		if role.Name == roleColorName {
			roleColorID = role.ID
			break
		}
	}
	if len(roleColorID) == 0 {
		return errors.New("Unknown color. Use `/colors` to see the available colors.")
	}

	for _, role := range c.Config.ColorRoles {
		// Attempt to remove all color roles regardless of if they have them or not.
		// Slow because of the REST requests...
		_ = s.GuildMemberRoleRemove(i.GuildID, rUserID, role.ID)
	}

	err = s.GuildMemberRoleAdd(i.GuildID, rUserID, roleColorID)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("User Color", "<@"+rUserID+">: Your color has been set to <@&"+roleColorID+">!", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}
