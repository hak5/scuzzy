package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
)

func (c *Commands) handleSetConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := options(i)
	configKey := opts["key"].StringValue()
	configVal := opts["value"].StringValue()

	rt := reflect.TypeOf(*c.Config)
	for f := 0; f < rt.NumField(); f++ {
		x := rt.Field(f)
		tagVal := strings.Split(x.Tag.Get("json"), ",")[0]
		tagName := x.Name

		if tagVal == configKey {
			prop := reflect.ValueOf(c.Config).Elem().FieldByName(tagName)

			switch prop.Interface().(type) {
			case string:
				prop.SetString(configVal)
				break
			case int:
				intVal, err := strconv.ParseInt(configVal, 10, 64)
				if err != nil {
					return err
				}
				prop.SetInt(intVal)
				break
			case float64:
				floatVal, err := strconv.ParseFloat(configVal, 64)
				if err != nil {
					return err
				}
				prop.SetFloat(floatVal)
				break
			case bool:
				boolVal, err := strconv.ParseBool(configVal)
				if err != nil {
					return err
				}
				prop.SetBool(boolVal)
				break
			default:
				return errors.New("Unsupported key value type")
			}

			msgE := c.CreateDefinedEmbed("Set Configuration", "Successfully set property '"+configKey+"'!", "success", i.Member.User)
			return c.replyEmbed(s, i, msgE)
		}
	}

	return errors.New("Unknown key specified")
}

func (c *Commands) handleGetConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	//TODO: Handle printing of slices (check the Type, loop accordingly)

	configKey := "all"
	if o, ok := options(i)["key"]; ok {
		configKey = o.StringValue()
	}

	msg := ""

	rt := reflect.TypeOf(*c.Config)
	for f := 0; f < rt.NumField(); f++ {
		x := rt.Field(f)
		tagVal := strings.Split(x.Tag.Get("json"), ",")[0]
		tagName := x.Name
		prop := reflect.ValueOf(c.Config).Elem().FieldByName(tagName)

		if configKey == "all" {
			switch prop.Interface().(type) {
			case string:
				if len(prop.String()) > 256 {
					// Truncate large values.
					msg += "`" + tagName + "` - " + "Truncated...\n"
				} else {
					msg += "`" + tagName + "` - `" + prop.String() + "`\n"
				}
				break
			default:
				// Ignore non strings for now...
				msg += "`" + tagName + "` - Skipped Value\n"
				continue
			}
		} else {
			if tagVal == configKey {
				switch prop.Interface().(type) {
				case string:
					msg += "`" + tagName + "` - `" + prop.String() + "`\n"
				default:
					// Ignore non strings for now...
					msg += "`" + tagName + "` - Skipped Value\n"
				}

				eMsg := c.CreateDefinedEmbed("Get Configuration", msg, "success", i.Member.User)
				return c.replyEmbed(s, i, eMsg)
			}
		}
	}

	if msg == "" {
		return errors.New("Unknown key specified")
	}

	eMsg := c.CreateDefinedEmbed("Get Configuration", msg, "success", i.Member.User)
	return c.replyEmbed(s, i, eMsg)
}

func (c *Commands) handleReloadConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	fBuf, err := ioutil.ReadFile(c.Config.ConfigPath)
	if err != nil {
		return err
	}

	conf := &models.Configuration{}

	err = json.Unmarshal(fBuf, &conf)
	if err != nil {
		return err
	}

	conf.ConfigPath = c.Config.ConfigPath
	conf.GuildName = c.Config.GuildName
	c.Config = conf
	c.Permissions.Config = conf

	// Refresh command choices (colors, joinable roles) from the new config
	c.registerCommands()
	err = c.SyncCommands(s)
	if err != nil {
		return err
	}

	eMsg := c.CreateDefinedEmbed("Reload Configuration", "Successfully reloaded configuration from disk", "success", i.Member.User)
	return c.replyEmbed(s, i, eMsg)
}

func (c *Commands) saveConfig() error {
	j, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.Config.ConfigPath, j, os.ModePerm)
}

func (c *Commands) handleSaveConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := c.saveConfig()
	if err != nil {
		return err
	}

	eMsg := c.CreateDefinedEmbed("Save Configuration", "Saved runtime configuration successfully", "success", i.Member.User)
	return c.replyEmbed(s, i, eMsg)
}

func (c *Commands) handleCat(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.reply(s, i, "https://giphy.com/gifs/cat-cute-no-rCxogJBzaeZuU")
}

func (c *Commands) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	msg := c.CreateDefinedEmbed("Ping", "Pong", "success", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	desc := "**Source**:   https://github.com/foxtrot/scuzzy\n"
	desc += "**Language**: Go\n"
	desc += "**Commands**: See `/help`\n\n\n"

	gm, err := s.GuildMember(c.Config.GuildID, s.State.User.ID)
	if err != nil {
		return err
	}

	d := models.CustomEmbed{
		Title:          "Scuzzy Information",
		Desc:           desc,
		ImageURL:       "",
		ImageH:         100,
		ImageW:         100,
		Color:          0xFFA500,
		URL:            "",
		Type:           "",
		Timestamp:      "",
		FooterText:     "Made with  ❤  by Foxtrot",
		FooterImageURL: "https://cdn.discordapp.com/avatars/514163441548656641/a_ac5e022e77e62e7793711ebde8cdf4a1.gif",
		ThumbnailURL:   gm.User.AvatarURL(""),
		ThumbnailH:     150,
		ThumbnailW:     150,
		ProviderURL:    "",
		ProviderText:   "",
		AuthorText:     "",
		AuthorURL:      "",
		AuthorImageURL: "",
	}

	msg := c.CreateCustomEmbed(&d)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	keys := make([]int, 0, len(c.ScuzzyCommands))
	for _, cmd := range c.ScuzzyCommands {
		keys = append(keys, cmd.Index)
	}
	sort.Ints(keys)

	desc := "**Available Commands**\n"
	for _, k := range keys {
		command := c.ScuzzyCommandsByIndex[k]

		if !command.AdminOnly && command.Type == discordgo.ChatApplicationCommand {
			desc += "`/" + command.Name + "` - " + command.Description + "\n"
		}
	}

	if c.Permissions.CheckAdminRole(i.Member) {
		desc += "\n"
		desc += "**Admin Commands**\n"
		for _, k := range keys {
			command := c.ScuzzyCommandsByIndex[k]

			if !command.AdminOnly {
				continue
			}
			if command.Type == discordgo.MessageApplicationCommand {
				desc += "`" + command.Name + "` - Right-click a message > Apps\n"
			} else {
				desc += "`/" + command.Name + "` - " + command.Description + "\n"
			}
		}
	}

	msg := c.CreateDefinedEmbed("Help", desc, "", i.Member.User)
	return c.replyEmbed(s, i, msg)
}

func (c *Commands) handleRules(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	msg := c.Config.RulesText
	embedTitle := "Rules (" + c.Config.GuildName + ")"
	embed := c.CreateDefinedEmbed(embedTitle, msg, "success", i.Member.User)
	return c.replyEmbed(s, i, embed)
}

func (c *Commands) handleMarkdownInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	public := false
	if o, ok := options(i)["public"]; ok {
		public = o.BoolValue() && c.Permissions.CheckAdminRole(i.Member)
	}

	desc := "*Italic* text goes between `*single asterisks*`\n"
	desc += "**Bold** text goes between `**double asterisks**`\n"
	desc += "***Bold and Italic*** text goes between `***triple asterisks***`\n"
	desc += "__Underlined__ text goes between `__double underscore__`\n"
	desc += "~~Strikethrough~~ text goes between `~~double tilde~~`\n"
	desc += "||Spoilers|| go between `|| double pipe ||`\n\n"
	desc += "You can combine the above styles.\n\n"
	desc += "Inline Code Blocks start and end with a single ``​`​``\n"
	desc += "Multi line Code Blocks start and end with ``​```​``\n"
	desc += "Multi line Code Blocks can also specify a language with ``​```​language`` at the start\n\n"
	desc += "Single line quotes start with `>`\n"
	desc += "Multi line quotes start with `>>>`\n"

	msg := c.CreateDefinedEmbed("Discord Markdown", desc, "", i.Member.User)
	if !public {
		// Only the requester sees it, replacing the old auto-delete after 15 seconds.
		return c.replyEmbed(s, i, msg)
	}

	_, err := s.ChannelMessageSendEmbed(i.ChannelID, msg)
	if err != nil {
		return err
	}
	return c.reply(s, i, "Posted.")
}

func (c *Commands) handleCtoF(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["celsius"].FloatValue()

	cels := (inF * 9.0 / 5.0) + 32.0
	celsF := float64(cels)

	msg := fmt.Sprintf("`%.1f°c` is `%.1f°f`", inF, celsF)

	e := c.CreateDefinedEmbed("Celsius to Farenheit", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleFtoC(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["farenheit"].FloatValue()

	faren := (inF - 32) * 5 / 9
	farenF := float64(faren)

	msg := fmt.Sprintf("`%.1f°f` is `%.1f°c`", inF, farenF)

	e := c.CreateDefinedEmbed("Farenheit to Celsius", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleMetersToFeet(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["meters"].FloatValue()

	meters := inF * 3.28
	metersF := float64(meters)

	msg := fmt.Sprintf("`%.1fm` is `%.1fft`", inF, metersF)

	e := c.CreateDefinedEmbed("Meters to Feet", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleFeetToMeters(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["feet"].FloatValue()

	feet := inF / 3.28
	feetF := float64(feet)

	msg := fmt.Sprintf("`%.1fft` is `%.1fm`", inF, feetF)

	e := c.CreateDefinedEmbed("Feet to Meters", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleCentimeterToInch(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["centimeters"].FloatValue()

	inch := inF / 2.54
	inchF := float64(inch)

	msg := fmt.Sprintf("`%.1fcm` is `%.1fin`", inF, inchF)

	e := c.CreateDefinedEmbed("Centimeter To Inch", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleInchToCentimeter(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	inF := options(i)["inches"].FloatValue()

	cm := inF * 2.54
	cmF := float64(cm)

	msg := fmt.Sprintf("`%.1fin` is `%.1fcm`", inF, cmF)

	e := c.CreateDefinedEmbed("Inch to Centimeter", msg, "", i.Member.User)
	return c.replyEmbed(s, i, e)
}

func (c *Commands) handleUserInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var (
		mHandle   *discordgo.Member
		requester *discordgo.Member
		err       error
	)

	requester = i.Member
	mHandle = i.Member
	if o, ok := options(i)["user"]; ok {
		mHandle, err = s.GuildMember(c.Config.GuildID, o.UserValue(nil).ID)
		if err != nil {
			return err
		}
	}

	rUserID := mHandle.User.ID
	rUserNick := mHandle.Nick
	rUsername := mHandle.User.Username
	rUserDiscrim := mHandle.User.Discriminator
	rUserAvatar := mHandle.User.AvatarURL("4096")
	rJoinTime := mHandle.JoinedAt
	rRoles := mHandle.Roles

	if len(rUserNick) == 0 {
		rUserNick = "No Nickname"
	}

	rRolesTidy := ""
	if len(rRoles) == 0 {
		rRolesTidy = "No Roles"
	} else {
		for _, role := range rRoles {
			rRolesTidy += "<@&" + role + "> "
		}
	}

	msg := "**User ID**: `" + rUserID + "`\n"
	msg += "**User Name**: `" + rUsername + "`\n"
	msg += "**User Nick**: `" + rUserNick + "`\n"
	msg += "**User Discrim**: `#" + rUserDiscrim + "`\n"
	msg += "**User Join**:  `" + rJoinTime.String() + "`\n"
	msg += "**User Roles**: " + rRolesTidy + "\n"

	embedData := models.CustomEmbed{
		URL:            "",
		Title:          "User Info (" + rUsername + ")",
		Desc:           msg,
		Type:           "",
		Timestamp:      time.Now().Format(time.RFC3339),
		Color:          0xFFA500,
		FooterText:     "Requested by " + requester.User.Username + "#" + requester.User.Discriminator,
		FooterImageURL: "",
		ImageURL:       "",
		ImageH:         0,
		ImageW:         0,
		ThumbnailURL:   rUserAvatar,
		ThumbnailH:     512,
		ThumbnailW:     512,
		ProviderURL:    "",
		ProviderText:   "",
		AuthorText:     "",
		AuthorURL:      "",
		AuthorImageURL: "",
	}

	embed := c.CreateCustomEmbed(&embedData)
	return c.replyEmbed(s, i, embed)
}

func (c *Commands) handleServerInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	g, err := s.Guild(c.Config.GuildID)
	if err != nil {
		return err
	}

	sID := c.Config.GuildID
	sName := c.Config.GuildName

	chans, _ := s.GuildChannels(c.Config.GuildID)
	sChannels := strconv.Itoa(len(chans))
	sEmojis := strconv.Itoa(len(g.Emojis))
	sRoles := strconv.Itoa(len(g.Roles))
	sRegion := g.Region

	iID, _ := strconv.Atoi(c.Config.GuildID)
	createdMSecs := ((iID / 4194304) + 1420070400000) / 1000
	sCreatedAt := time.Unix(int64(createdMSecs), 0).Format(time.RFC1123)

	sIconURL := g.IconURL("256")

	user := i.Member.User

	desc := "**Server ID**: `" + sID + "`\n"
	desc += "**Server Name**: `" + sName + "`\n"
	desc += "**Server Channels**: `" + sChannels + "`\n"
	desc += "**Server Emojis**: `" + sEmojis + "`\n"
	desc += "**Server Roles**: `" + sRoles + "`\n"
	desc += "**Server Region**: `" + sRegion + "`\n"
	desc += "**Server Creation**: `" + sCreatedAt + "`\n"

	embedData := models.CustomEmbed{
		URL:            "",
		Title:          "Server Info (" + sName + ")",
		Desc:           desc,
		Type:           "",
		Timestamp:      time.Now().Format(time.RFC3339),
		Color:          0xFFA500,
		FooterText:     "Requested by " + user.Username + "#" + user.Discriminator,
		FooterImageURL: "",
		ImageURL:       "",
		ImageH:         0,
		ImageW:         0,
		ThumbnailURL:   sIconURL,
		ThumbnailH:     256,
		ThumbnailW:     256,
		ProviderURL:    "",
		ProviderText:   "",
		AuthorText:     "",
		AuthorURL:      "",
		AuthorImageURL: "",
	}

	msg := c.CreateCustomEmbed(&embedData)
	return c.replyEmbed(s, i, msg)
}
func (c *Commands) handleGoogle4U(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	input := options(i)["query"].StringValue()

	desc := "https://letmegooglethat.com/?q=" + url.QueryEscape(input)

	msg := c.CreateDefinedEmbed("Google", desc, "", i.Member.User)
	return c.replyEmbed(s, i, msg)
}
