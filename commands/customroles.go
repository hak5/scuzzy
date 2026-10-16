package commands

import (
	"errors"
	discordgo "github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
	"strings"
)

func (c *Commands) handleListCustomRoles(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	msgC := "You can choose from the following roles:\n\n"
	for _, v := range c.Config.CustomRoles {
		msgC += "<@&" + v.ID + "> (" + v.ShortName + ")\n"
	}
	msgC += "\n\n Use `/joinrole <role_name>` to join a role.\n"
	msgC += "Example: `/joinrole pineapple`.\n"

	msg := c.CreateDefinedEmbed("Joinable Roles", msgC, "", i.Member.User)

	return c.replyEmbed(s, i, msg)
}

func (c *Commands) customRoleID(shortName string) (string, error) {
	shortName = strings.ToLower(shortName)
	for _, role := range c.Config.CustomRoles {
		if role.ShortName == shortName {
			return role.ID, nil
		}
	}
	return "", errors.New("Unknown role. Use `/listroles` to see the available roles.")
}

func (c *Commands) handleJoinCustomRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	rUserID := i.Member.User.ID

	desiredRoleID, err := c.customRoleID(options(i)["role"].StringValue())
	if err != nil {
		return err
	}

	err = s.GuildMemberRoleAdd(i.GuildID, rUserID, desiredRoleID)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Join Role", "<@"+rUserID+">: You have joined <@&"+desiredRoleID+">!", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleLeaveCustomRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	rUserID := i.Member.User.ID

	desiredRoleID, err := c.customRoleID(options(i)["role"].StringValue())
	if err != nil {
		return err
	}

	err = s.GuildMemberRoleRemove(i.GuildID, rUserID, desiredRoleID)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Leave Role", "<@"+rUserID+">: You have left <@&"+desiredRoleID+">!", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleAddCustomRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)

	customRole := models.CustomRole{
		Name:      "",
		ShortName: strings.ToLower(opts["short_name"].StringValue()),
		ID:        opts["role"].RoleValue(s, i.GuildID).ID,
	}

	c.Config.CustomRoles = append(c.Config.CustomRoles, customRole)

	err := c.saveConfig()
	if err != nil {
		return err
	}

	// Refresh the role choices on /joinrole and /leaverole
	c.registerCommands()
	err = c.SyncCommands(s)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Add Role", "<@&"+customRole.ID+"> can now be joined as `"+customRole.ShortName+"`.", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}
