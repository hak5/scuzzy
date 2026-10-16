package commands

import (
	"errors"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func (c *Commands) RegisterCommand(cmd ScuzzyCommand) {
	log.Printf("[*] Registering Command '%s'\n", cmd.Name)
	if cmd.Type == 0 {
		cmd.Type = discordgo.ChatApplicationCommand
	}
	cmd.Index = len(c.ScuzzyCommands) + 1
	c.ScuzzyCommands[cmd.Name] = cmd
	c.ScuzzyCommandsByIndex[cmd.Index] = cmd
}

func (c *Commands) RegisterHandlers() {
	c.startAutoMod()
	c.registerCommands()
}

// registerCommands builds the command list. Choices come from the config, so this
// is re-run (followed by SyncCommands) whenever the config changes.
func (c *Commands) registerCommands() {
	c.ScuzzyCommands = make(map[string]ScuzzyCommand)
	c.ScuzzyCommandsByIndex = make(map[int]ScuzzyCommand)

	var colors []string
	for _, r := range c.Config.ColorRoles {
		colors = append(colors, r.Name)
	}
	var customRoles []string
	for _, r := range c.Config.CustomRoles {
		customRoles = append(customRoles, r.ShortName)
	}

	// Misc Commands
	c.RegisterCommand(ScuzzyCommand{Name: "help", Description: "Show Help Text", Ephemeral: true, Handler: c.handleHelp})
	c.RegisterCommand(ScuzzyCommand{Name: "info", Description: "Show Bot Info", Handler: c.handleInfo})
	c.RegisterCommand(ScuzzyCommand{Name: "md", Description: "Show common Discord MarkDown formatting", Ephemeral: true, Handler: c.handleMarkdownInfo,
		Options: []*discordgo.ApplicationCommandOption{optBool("public", "Show to everyone in the channel (staff only)")}})
	c.RegisterCommand(ScuzzyCommand{Name: "userinfo", Description: "Display a users information", Handler: c.handleUserInfo,
		Options: []*discordgo.ApplicationCommandOption{optUser("user", "User to look up (defaults to you)", false)}})
	c.RegisterCommand(ScuzzyCommand{Name: "serverinfo", Description: "Display the current servers information", Handler: c.handleServerInfo})
	c.RegisterCommand(ScuzzyCommand{Name: "no", Description: "No.", Handler: c.handleCat})

	// User Settings
	c.RegisterCommand(ScuzzyCommand{Name: "colors", Description: "Display available color roles", Handler: c.handleUserColors})
	c.RegisterCommand(ScuzzyCommand{Name: "color", Description: "Set a color role for yourself", Handler: c.handleUserColor,
		Options: []*discordgo.ApplicationCommandOption{optString("color", "Color to use", true, colors...)}})
	c.RegisterCommand(ScuzzyCommand{Name: "listroles", Description: "List user joinable roles", Handler: c.handleListCustomRoles})
	c.RegisterCommand(ScuzzyCommand{Name: "joinrole", Description: "Join an available role for yourself", Handler: c.handleJoinCustomRole,
		Options: []*discordgo.ApplicationCommandOption{optString("role", "Role to join", true, customRoles...)}})
	c.RegisterCommand(ScuzzyCommand{Name: "leaverole", Description: "Leave an available role", Handler: c.handleLeaveCustomRole,
		Options: []*discordgo.ApplicationCommandOption{optString("role", "Role to leave", true, customRoles...)}})

	// Conversion Helpers
	c.RegisterCommand(ScuzzyCommand{Name: "ctof", Description: "Convert Celsius to Farenheit", Handler: c.handleCtoF,
		Options: []*discordgo.ApplicationCommandOption{optNumber("celsius", "Temperature in °C")}})
	c.RegisterCommand(ScuzzyCommand{Name: "ftoc", Description: "Convert Farenheit to Celsius", Handler: c.handleFtoC,
		Options: []*discordgo.ApplicationCommandOption{optNumber("farenheit", "Temperature in °F")}})
	c.RegisterCommand(ScuzzyCommand{Name: "metofe", Description: "Convert Meters to Feet", Handler: c.handleMetersToFeet,
		Options: []*discordgo.ApplicationCommandOption{optNumber("meters", "Distance in meters")}})
	c.RegisterCommand(ScuzzyCommand{Name: "fetome", Description: "Convert Feet to Meters", Handler: c.handleFeetToMeters,
		Options: []*discordgo.ApplicationCommandOption{optNumber("feet", "Distance in feet")}})
	c.RegisterCommand(ScuzzyCommand{Name: "cmtoin", Description: "Convert Centimeters to Inches", Handler: c.handleCentimeterToInch,
		Options: []*discordgo.ApplicationCommandOption{optNumber("centimeters", "Distance in centimeters")}})
	c.RegisterCommand(ScuzzyCommand{Name: "intocm", Description: "Convert Inches to Centimeters", Handler: c.handleInchToCentimeter,
		Options: []*discordgo.ApplicationCommandOption{optNumber("inches", "Distance in inches")}})
	c.RegisterCommand(ScuzzyCommand{Name: "google4u", Description: "Displays a letmegooglethat link", Handler: c.handleGoogle4U,
		Options: []*discordgo.ApplicationCommandOption{optString("query", "What to search for", true)}})

	//Product Helpers
	c.RegisterCommand(ScuzzyCommand{Name: "askdocs", Description: "Ask docs.hak5.org a question", Handler: c.handleAskDocs,
		Options: []*discordgo.ApplicationCommandOption{optString("question", "Your question", true)}})
	c.RegisterCommand(ScuzzyCommand{Name: "docs", Description: "Link Docs", Handler: c.handleDocs})
	c.RegisterCommand(ScuzzyCommand{Name: "github", Description: "Link Github", Handler: c.handleGithub})
	c.RegisterCommand(ScuzzyCommand{Name: "payloads", Description: "Link Github + PayloadHub", Handler: c.handlePayloads})
	c.RegisterCommand(ScuzzyCommand{Name: "payloadhub", Description: "Link PayloadHub", Handler: c.handlePayloadHub})
	c.RegisterCommand(ScuzzyCommand{Name: "support", Description: "Link Support", Handler: c.handleSupport})
	c.RegisterCommand(ScuzzyCommand{Name: "shop", Description: "Link Shop", Handler: c.handleShop})
	c.RegisterCommand(ScuzzyCommand{Name: "invite", Description: "Link direct channel invite", Handler: c.handleInvite})
	// TODO payloadstudio view

	// Admin Commands
	c.RegisterCommand(ScuzzyCommand{Name: "ping", Description: "Ping Scuzzy", AdminOnly: true, Ephemeral: true, Handler: c.handlePing})
	c.RegisterCommand(ScuzzyCommand{Name: "rules", Description: "Display the Server Rules", AdminOnly: true, Handler: c.handleRules})
	c.RegisterCommand(ScuzzyCommand{Name: "status", Description: "Set Bot Status", AdminOnly: true, Ephemeral: true, Handler: c.handleSetStatus,
		Options: []*discordgo.ApplicationCommandOption{optString("status", "Status text", true)}})
	c.RegisterCommand(ScuzzyCommand{Name: "purge", Description: "Purge Channel Messages", AdminOnly: true, Ephemeral: true, Handler: c.handlePurgeChannel,
		Options: []*discordgo.ApplicationCommandOption{optInt("count", "Number of messages to delete", 1, 100)}})
	c.RegisterCommand(ScuzzyCommand{Name: "kick", Description: "Kick a User", AdminOnly: true, Handler: c.handleKickUser,
		Options: []*discordgo.ApplicationCommandOption{optUser("user", "User to kick", true), optString("reason", "Reason", false)}})
	c.RegisterCommand(ScuzzyCommand{Name: "ban", Description: "Ban a User", AdminOnly: true, Handler: c.handleBanUser,
		Options: []*discordgo.ApplicationCommandOption{optUser("user", "User to ban (mention or ID)", true), optString("reason", "Reason", false)}})
	c.RegisterCommand(ScuzzyCommand{Name: "slow", Description: "Set Channel Slow Mode", AdminOnly: true, Handler: c.handleSetSlowmode,
		Options: []*discordgo.ApplicationCommandOption{optInt("seconds", "Seconds between messages", 0, 21600), optBool("all", "Apply to every channel")}})
	c.RegisterCommand(ScuzzyCommand{Name: "unslow", Description: "Unset Channel Slow Mode", AdminOnly: true, Handler: c.handleUnsetSlowmode,
		Options: []*discordgo.ApplicationCommandOption{optBool("all", "Apply to every channel")}})
	c.RegisterCommand(ScuzzyCommand{Name: "ignore", Description: "Add a user to Scuzzy's ignore list", AdminOnly: true, Ephemeral: true, Handler: c.handleIgnoreUser,
		Options: []*discordgo.ApplicationCommandOption{optUser("user", "User to ignore", true)}})
	c.RegisterCommand(ScuzzyCommand{Name: "unignore", Description: "Remove a user from Scuzzy's ignore list", AdminOnly: true, Ephemeral: true, Handler: c.handleUnIgnoreUser,
		Options: []*discordgo.ApplicationCommandOption{optUser("user", "User to unignore", true)}})
	c.RegisterCommand(ScuzzyCommand{Name: "setconfig", Description: "Set Configuration", AdminOnly: true, Ephemeral: true, Handler: c.handleSetConfig,
		Options: []*discordgo.ApplicationCommandOption{optString("key", "Config key", true), optString("value", "New value", true)}})
	c.RegisterCommand(ScuzzyCommand{Name: "getconfig", Description: "Print Configuration", AdminOnly: true, Ephemeral: true, Handler: c.handleGetConfig,
		Options: []*discordgo.ApplicationCommandOption{optString("key", "Config key (defaults to all)", false)}})
	c.RegisterCommand(ScuzzyCommand{Name: "saveconfig", Description: "Save Configuration to Disk", AdminOnly: true, Ephemeral: true, Handler: c.handleSaveConfig})
	c.RegisterCommand(ScuzzyCommand{Name: "reloadconfig", Description: "Reload Configuration", AdminOnly: true, Ephemeral: true, Handler: c.handleReloadConfig})
	c.RegisterCommand(ScuzzyCommand{Name: "addrole", Description: "Add a joinable role", AdminOnly: true, Ephemeral: true, Handler: c.handleAddCustomRole,
		Options: []*discordgo.ApplicationCommandOption{optString("short_name", "Name members will pick", true), optRole("role", "Role to make joinable")}})
	c.RegisterCommand(ScuzzyCommand{Name: "Move message", Type: discordgo.MessageApplicationCommand, AdminOnly: true, Handler: c.handleMoveto})
	c.RegisterCommand(ScuzzyCommand{Name: "Copy message", Type: discordgo.MessageApplicationCommand, AdminOnly: true, Handler: c.handleBringto})
}

func (c *Commands) ProcessMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) error {
	msgChannelID := m.ChannelID

	if m.BeforeDelete == nil {
		return errors.New("Couldn't get deleted message data.")
	}

	msgContent := m.BeforeDelete.Content
	msgAuthor := m.BeforeDelete.Author

	msg := "`Username` - " + msgAuthor.Username + "#" + msgAuthor.Discriminator + "\n"
	msg += "`User ID` - " + msgAuthor.ID + "\n"
	msg += "`Channel` - <#" + msgChannelID + ">\n"
	msg += "`Message` - " + msgContent + "\n"

	embed := c.CreateDefinedEmbed("Deleted Message", msg, "", msgAuthor)
	_, err := s.ChannelMessageSendEmbed(c.Config.LoggingChannel, embed)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) ProcessMessageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) error {
	msgChannelID := m.ChannelID

	msg := "`Channel` - <#" + msgChannelID + ">\n"
	msg += "Message IDs: \n"
	for k, v := range m.Messages {
		msg += strconv.Itoa(k) + ": `" + v + "`\n"
	}

	embed := c.CreateDefinedEmbed("Deleted Bulk Messages", msg, "", nil)
	_, err := s.ChannelMessageSendEmbed(c.Config.LoggingChannel, embed)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) ProcessUserJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) error {
	userChannel, err := s.UserChannelCreate(m.User.ID)
	if err != nil {
		log.Print("[!] Error (User Join): " + err.Error())
		return err
	}

	_, err = s.ChannelMessageSend(userChannel.ID, c.Config.WelcomeText)
	if err != nil {
		log.Print("[!] Error (User Join): " + err.Error())
		return err
	}

	for _, roleID := range c.Config.JoinRoleIDs {
		err = s.GuildMemberRoleAdd(c.Config.GuildID, m.User.ID, roleID)
		if err != nil {
			log.Print("[!] Error (User Join)" + err.Error())
			return err
		}
	}

	return nil
}

func (c *Commands) ProcessMessage(s *discordgo.Session, m interface{}) {
	switch m.(type) {
	case *discordgo.MessageCreate:
		// Auto moderation (honeypot channel, spam detection)
		c.ProcessAutoMod(s, m.(*discordgo.MessageCreate).Message)
		break
	case *discordgo.InteractionCreate:
		// Slash commands, message commands and their components
		c.ProcessInteraction(s, m.(*discordgo.InteractionCreate))
		break
	case *discordgo.MessageEdit:
		//TODO logging
		break
	case *discordgo.MessageDelete:
		// Log deleted messages to the logging channel.
		err := c.ProcessMessageDelete(s, m.(*discordgo.MessageDelete))
		if err != nil {
			eMsg := c.CreateDefinedEmbed("Error (Message Deleted)", err.Error(), "error", nil)
			_, err = s.ChannelMessageSendEmbed(c.Config.LoggingChannel, eMsg)
			if err != nil {
				log.Println("[!] Error " + err.Error())
			}
		}
		break
	case *discordgo.MessageDeleteBulk:
		err := c.ProcessMessageDeleteBulk(s, m.(*discordgo.MessageDeleteBulk))
		if err != nil {
			eMsg := c.CreateDefinedEmbed("Error (Message Bulk Deleted)", err.Error(), "error", nil)
			_, err = s.ChannelMessageSendEmbed(c.Config.LoggingChannel, eMsg)
			if err != nil {
				log.Println("[!] Error " + err.Error())
			}
		}
		break
	case *discordgo.GuildMemberAdd:
		// Handle new member (Welcome message, etc)
		err := c.ProcessUserJoin(s, m.(*discordgo.GuildMemberAdd))
		if err != nil {
			log.Println("[!] Error (Guild Member Joined): " + err.Error())
		}
		break
	}
}
