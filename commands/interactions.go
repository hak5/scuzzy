package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// Admin commands are hidden from members without this permission by default.
// Server admins can change who sees each command under Server Settings > Integrations.
var adminDefaultPermissions int64 = discordgo.PermissionManageMessages

// SyncCommands replaces the guild's slash commands with the registered set.
// Guild commands update immediately, unlike global ones.
func (c *Commands) SyncCommands(s *discordgo.Session) error {
	guildOnly := []discordgo.InteractionContextType{discordgo.InteractionContextGuild}

	var cmds []*discordgo.ApplicationCommand
	for idx := 1; idx <= len(c.ScuzzyCommandsByIndex); idx++ {
		sc := c.ScuzzyCommandsByIndex[idx]
		ac := &discordgo.ApplicationCommand{
			Name:     sc.Name,
			Type:     sc.Type,
			Contexts: &guildOnly,
		}
		if sc.Type == discordgo.ChatApplicationCommand {
			ac.Description = sc.Description
			ac.Options = sc.Options
		}
		if sc.AdminOnly {
			ac.DefaultMemberPermissions = &adminDefaultPermissions
		}
		cmds = append(cmds, ac)
	}

	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, c.Config.GuildID, cmds)
	if err != nil {
		return err
	}

	log.Printf("[*] Synced %d commands with Discord\n", len(cmds))
	return nil
}

func (c *Commands) ProcessInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Ignore other guilds and DMs
	if i.GuildID != c.Config.GuildID || i.Member == nil {
		return
	}

	if c.Permissions.CheckIgnoredUser(i.Member.User) {
		log.Printf("[*] Ignoring interaction from ignored user.")
		c.respondError(s, i, "You are not able to use Scuzzy.")
		return
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		c.processCommand(s, i)
	case discordgo.InteractionMessageComponent:
		c.processComponent(s, i)
	}
}

func (c *Commands) processCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cName := i.ApplicationCommandData().Name
	user := i.Member.User

	cmd, ok := c.ScuzzyCommands[cName]
	if !ok {
		return
	}

	if cmd.AdminOnly && !c.Permissions.CheckAdminRole(i.Member) {
		log.Printf("[*] User %s tried to run admin command %s\n", user.Username, cName)
		c.respondError(s, i, "This command is only available to staff.")
		return
	}

	if !c.Permissions.CheckCommandRestrictions(cName, i.ChannelID) {
		c.respondError(s, i, "This command is not allowed in this channel.")
		return
	}

	log.Printf("[*] Running command %s (Requested by %s)\n", cName, user.Username)

	// Message commands respond on their own (with a channel picker).
	if cmd.Type == discordgo.MessageApplicationCommand {
		if err := cmd.Handler(s, i); err != nil {
			log.Printf("[!] Command %s (Requested by %s) had error: '%s'\n", cName, user.Username, err.Error())
			c.respondError(s, i, err.Error())
		}
		return
	}

	// Defer straight away so slow commands (Discord API calls, docs search) don't hit the 3 second limit.
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}
	if cmd.Ephemeral {
		resp.Data.Flags = discordgo.MessageFlagsEphemeral
	}
	if err := s.InteractionRespond(i.Interaction, resp); err != nil {
		log.Println("[!] Error (Defer Interaction): " + err.Error())
		return
	}

	err := cmd.Handler(s, i)
	if err != nil {
		log.Printf("[!] Command %s (Requested by %s) had error: '%s'\n", cName, user.Username, err.Error())

		eMsg := c.CreateDefinedEmbed("Error ("+cName+")", err.Error(), "error", user)
		if err = c.replyEmbed(s, i, eMsg); err != nil {
			log.Println("[!] Error: " + err.Error())
		}
	}
}

func (c *Commands) processComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !c.Permissions.CheckAdminRole(i.Member) {
		c.respondError(s, i, "This action is only available to staff.")
		return
	}

	// Acknowledge first; the picker message is then edited with the result.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate})
	if err != nil {
		log.Println("[!] Error (Defer Interaction): " + err.Error())
		return
	}

	err = c.handleMoveSelect(s, i)
	if err != nil {
		log.Println("[!] Error (Message Move): " + err.Error())
		eMsg := c.CreateDefinedEmbed("Error", err.Error(), "error", i.Member.User)
		embeds := []*discordgo.MessageEmbed{eMsg}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds, Components: &[]discordgo.MessageComponent{}})
		if err != nil {
			log.Println("[!] Error: " + err.Error())
		}
	}
}

// reply fills in a deferred response.
func (c *Commands) reply(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embeds ...*discordgo.MessageEmbed) error {
	if embeds == nil {
		embeds = []*discordgo.MessageEmbed{}
	}
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &embeds,
	})
	return err
}

func (c *Commands) replyEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) error {
	return c.reply(s, i, "", embed)
}

// respondError sends an immediate private error, for interactions that haven't been responded to yet.
func (c *Commands) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{c.CreateDefinedEmbed("Error", msg, "error", nil)},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("[!] Error (Interaction Response): " + err.Error())
	}
}

func options(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	opts := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, o := range i.ApplicationCommandData().Options {
		opts[o.Name] = o
	}
	return opts
}

func optString(name, desc string, required bool, choices ...string) *discordgo.ApplicationCommandOption {
	o := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: desc,
		Required:    required,
	}
	// Discord allows at most 25 choices; past that the option stays free text.
	if len(choices) <= 25 {
		for _, ch := range choices {
			o.Choices = append(o.Choices, &discordgo.ApplicationCommandOptionChoice{Name: ch, Value: ch})
		}
	}
	return o
}

func optNumber(name, desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionNumber,
		Name:        name,
		Description: desc,
		Required:    true,
	}
}

func optInt(name, desc string, min, max float64) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        name,
		Description: desc,
		Required:    true,
		MinValue:    &min,
		MaxValue:    max,
	}
}

func optUser(name, desc string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        name,
		Description: desc,
		Required:    required,
	}
}

func optRole(name, desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionRole,
		Name:        name,
		Description: desc,
		Required:    true,
	}
}

func optBool(name, desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        name,
		Description: desc,
	}
}
