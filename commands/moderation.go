package commands

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/actions"
	"strconv"
	"strings"
)

func (c *Commands) setSlowmode(s *discordgo.Session, channelID string, all bool, secs int) error {
	if all {
		channels, err := s.GuildChannels(c.Config.GuildID)
		if err != nil {
			return err
		}

		for _, channel := range channels {
			currPos := channel.Position
			s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
				Position:         &currPos,
				RateLimitPerUser: &secs,
			})
		}
		return nil
	}

	currChan, err := s.Channel(channelID)
	if err != nil {
		return err
	}

	currPos := currChan.Position
	_, err = s.ChannelEditComplex(channelID, &discordgo.ChannelEdit{
		Position:         &currPos,
		RateLimitPerUser: &secs,
	})
	return err
}

func (c *Commands) handleSetSlowmode(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)
	slowModeTime := int(opts["seconds"].IntValue())
	all := opts["all"] != nil && opts["all"].BoolValue()

	err := c.setSlowmode(s, i.ChannelID, all, slowModeTime)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Slow Mode", "Successfully set Slow Mode to `"+strconv.Itoa(slowModeTime)+"`.", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleUnsetSlowmode(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)
	all := opts["all"] != nil && opts["all"].BoolValue()

	err := c.setSlowmode(s, i.ChannelID, all, 0)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Slow Mode", "Successfully unset Slow Mode", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handlePurgeChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	msgCount := int(options(i)["count"].IntValue())

	chanMsgs, err := s.ChannelMessages(i.ChannelID, msgCount, "", "", "")
	if err != nil {
		return err
	}

	var delMsgs []string
	for _, v := range chanMsgs {
		delMsgs = append(delMsgs, v.ID)
	}

	err = deleteMessages(s, i.ChannelID, delMsgs)
	if err != nil {
		return err
	}

	msg := c.CreateDefinedEmbed("Purge Channel", "Purged `"+strconv.Itoa(len(delMsgs))+"` messages!", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleKickUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)

	kickReason := ""
	if o, ok := opts["reason"]; ok {
		kickReason = o.StringValue()
	}

	mHandle, err := s.GuildMember(c.Config.GuildID, opts["user"].UserValue(nil).ID)
	if err != nil {
		return err
	}

	err = actions.KickUser(s, c.Config.GuildID, mHandle.User.ID, kickReason)
	if err != nil {
		return err
	}

	msg := "User `" + mHandle.User.Username + "#" + mHandle.User.Discriminator + "` was kicked.\n"
	if len(kickReason) > 0 {
		msg += "Reason: `" + kickReason + "`\n"
	}

	embed := c.CreateDefinedEmbed("Kick User", msg, "success", i.Member.User)
	return c.replyEmbed(s, i, embed)
}

func (c *Commands) handleBanUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)

	banReason := ""
	if o, ok := opts["reason"]; ok {
		banReason = o.StringValue()
	}

	// Members who already left can still be banned by ID
	mHandle := opts["user"].UserValue(s)
	if mHandle == nil {
		return errors.New("Unknown user.")
	}

	err := actions.BanUser(s, c.Config.GuildID, mHandle.ID, banReason)
	if err != nil {
		return err
	}

	msg := "User `" + mHandle.Username + "#" + mHandle.Discriminator + "` was banned.\n"
	if len(banReason) > 0 {
		msg += "Reason: `" + banReason + "`\n"
	}

	embed := c.CreateDefinedEmbed("Ban User", msg, "success", i.Member.User)
	return c.replyEmbed(s, i, embed)
}

func (c *Commands) handleIgnoreUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	idStr := options(i)["user"].UserValue(nil).ID

	c.Config.IgnoredUsers = append(c.Config.IgnoredUsers, idStr)

	err := c.saveConfig()
	if err != nil {
		return err
	}

	eMsg := c.CreateDefinedEmbed("Ignore User", "<@!"+idStr+"> is now being ignored.", "success", i.Member.User)
	return c.replyEmbed(s, i, eMsg)
}

func (c *Commands) handleUnIgnoreUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	idStr := options(i)["user"].UserValue(nil).ID

	for k, v := range c.Config.IgnoredUsers {
		if v == idStr {
			c.Config.IgnoredUsers[k] = c.Config.IgnoredUsers[len(c.Config.IgnoredUsers)-1]
			c.Config.IgnoredUsers = c.Config.IgnoredUsers[:len(c.Config.IgnoredUsers)-1]
			break
		}
	}

	err := c.saveConfig()
	if err != nil {
		return err
	}

	eMsg := c.CreateDefinedEmbed("Unignore User", "<@!"+idStr+"> is not being ignored.", "success", i.Member.User)
	return c.replyEmbed(s, i, eMsg)
}

func (c *Commands) handleMoveto(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.promptMoveChannel(s, i, "move")
}

func (c *Commands) handleBringto(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.promptMoveChannel(s, i, "copy")
}

// promptMoveChannel answers a "Move message"/"Copy message" right-click with a private channel picker.
// Discord includes the target message's text in message command interactions, so this works without
// the Message Content intent. The text is carried in the picker's own embed rather than stored by the bot.
func (c *Commands) promptMoveChannel(s *discordgo.Session, i *discordgo.InteractionCreate, action string) error {
	data := i.ApplicationCommandData()
	target, ok := data.Resolved.Messages[data.TargetID]
	if !ok {
		return errors.New("error getting target message")
	}

	text := target.Content
	for _, a := range target.Attachments {
		text += "\n" + a.URL
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("This message has no text to " + action + ".")
	}

	preview := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    target.Author.Username,
			IconURL: target.Author.AvatarURL("64"),
		},
		Description: text,
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Where should this message be " + map[string]string{"move": "moved", "copy": "copied"}[action] + " to?",
			Embeds:  []*discordgo.MessageEmbed{preview},
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						MenuType:     discordgo.ChannelSelectMenu,
						CustomID:     action + ":" + target.ID + ":" + target.Author.ID,
						Placeholder:  "Pick a channel",
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews},
					},
				}},
			},
		},
	})
}

// handleMoveSelect finishes a move/copy once a moderator picks the destination channel.
func (c *Commands) handleMoveSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()

	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 3 || (parts[0] != "move" && parts[0] != "copy") || len(data.Values) != 1 {
		return errors.New("unknown action")
	}
	action, messageID, authorID := parts[0], parts[1], parts[2]
	moveToId := data.Values[0]
	modID := i.Member.User.ID

	if len(i.Message.Embeds) == 0 {
		return errors.New("error getting target message")
	}
	original := i.Message.Embeds[0]

	var announce, header string
	if action == "move" {
		announce = "Moderator <@" + modID + "> moved <@" + authorID + ">'s message to <#" + moveToId + "> to continue the conversation there"
		header = "Moderator <@" + modID + "> moved this message from <#" + i.ChannelID + ">\n\nOriginal message from <@" + authorID + ">:"
	} else {
		announce = "<@" + modID + "> copied <@" + authorID + ">'s message to <#" + moveToId + "> to continue the conversation there."
		header = "<@" + modID + "> copied this message from <#" + i.ChannelID + ">\n\nOriginal message from <@" + authorID + ">:"
	}

	_, err := s.ChannelMessageSend(i.ChannelID, announce)
	if err != nil {
		return errors.New("error announcing move")
	}

	_, err = s.ChannelMessageSendComplex(moveToId, &discordgo.MessageSend{
		Content: header,
		Embeds: []*discordgo.MessageEmbed{{
			Author:      original.Author,
			Description: original.Description,
		}},
	})
	if err != nil {
		return errors.New("error moving message")
	}

	done := "Copied to <#" + moveToId + ">."
	if action == "move" {
		err = s.ChannelMessageDelete(i.ChannelID, messageID)
		if err != nil {
			return errors.New("error removing original message")
		}
		done = "Moved to <#" + moveToId + ">."
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &done,
		Embeds:     &[]*discordgo.MessageEmbed{},
		Components: &[]discordgo.MessageComponent{},
	})
	return err
}
