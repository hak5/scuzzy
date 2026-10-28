package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
	"log"
	"net/http"
	"strings"
)

// TODO for use in command arguments
/* Match on one two or three word arguments*/
func aliasMatch(product models.Product, args []string) bool {
	for n := 1; n <= 3 && n < len(args); n++ {
		phrase := strings.ToLower(strings.Join(args[1:n+1], " "))
		for _, alias := range product.Aliases {
			if strings.ToLower(alias) == phrase {
				return true
			}
		}
	}
	return false
}

func (c *Commands) matchAlias(args []string) models.Product {
	switch true {
	case aliasMatch(models.Ducky, args):
		return models.Ducky
	case aliasMatch(models.PayloadStudio, args):
		return models.PayloadStudio
	case aliasMatch(models.Bunny, args):
		return models.Bunny
	case aliasMatch(models.Croc, args):
		return models.Croc
	case aliasMatch(models.Shark, args):
		return models.Shark
	case aliasMatch(models.CloudC2, args):
		return models.CloudC2
	case aliasMatch(models.Crab, args):
		return models.Crab
	case aliasMatch(models.Pineapple, args):
		return models.Pineapple
	case aliasMatch(models.Coconut, args):
		return models.Coconut
	case aliasMatch(models.Squirrel, args):
		return models.Squirrel
	case aliasMatch(models.Turtle, args):
		return models.Turtle
	case aliasMatch(models.OMG, args):
		return models.OMG
	default:
		return models.Default
	}
}

// why i didnt just use a map? because im dumb
func (c *Commands) getProductFromChannelID(cid string) models.Product {
	switch cid {
	case models.Ducky.ChannelID:
		return models.Ducky
	case models.PayloadStudio.ChannelID:
		return models.PayloadStudio
	case models.Bunny.ChannelID:
		return models.Bunny
	case models.Croc.ChannelID:
		return models.Croc
	case models.Shark.ChannelID:
		return models.Shark
	case models.CloudC2.ChannelID:
		return models.CloudC2
	case models.Crab.ChannelID:
		return models.Crab
	case models.Pineapple.ChannelID:
		return models.Pineapple
	case models.Coconut.ChannelID:
		return models.Coconut
	case models.Squirrel.ChannelID:
		return models.Squirrel
	case models.Turtle.ChannelID:
		return models.Turtle
	case models.OMG.ChannelID:
		return models.OMG
	default:
		return models.Default
	}
}

func (c *Commands) handleAskDocs(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	input := options(i)["question"].StringValue()

	model := c.getProductFromChannelID(i.ChannelID)

	// Default guard
	if model.SpaceID == models.Default.SpaceID {
		return c.reply(s, i, "AI Search not enabled for this product, but you can still look for yourself here: "+model.Docslink)
	}

	pre := "<:thonk:666272807524761620> Asking the"
	post := "docs..."
	askEndpoint := models.GITBOOKAPI + model.SpaceID + models.SEARCHENDPOINT
	searchingIn := pre + " " + model.Emoji + " " + model.ProductName + " " + post
	err := c.reply(s, i, searchingIn)
	if err != nil {
		return err
	}

	j, err := json.Marshal(&models.AISearchRequest{Query: input})
	if err != nil {
		return errors.New("error marshalling input")
	}

	r, err := http.NewRequest("POST", askEndpoint, bytes.NewBuffer(j))
	if err != nil {
		return errors.New("error creating request")
	}

	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return errors.New("error creating request client")
	}

	defer res.Body.Close()

	resp := &models.AISearchResult{}
	derr := json.NewDecoder(res.Body).Decode(resp)
	if derr != nil {
		return errors.New("error getting response from docs")
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("error getting response from docs: " + res.Status)
	}

	prepend := "<:9999iq:1050238565671514172> AI Generated Response:\n\n"
	err = c.reply(s, i, "> "+input+"\n\n"+prepend+resp.Answer.Text+"\n\n"+model.Docslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}

	return nil
}

func (c *Commands) handleDocs(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, model.Emoji+" "+model.ProductName+" Docs: "+model.Docslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleGithub(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, model.Emoji+" "+model.ProductName+" Github: "+model.Githublink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handlePayloads(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, model.Emoji+" "+model.ProductName+" Github: "+model.Githublink+"\n"+
		model.Emoji+" "+model.ProductName+" PayloadHub: "+model.Payloadslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handlePayloadHub(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, model.Emoji+" "+model.ProductName+" PayloadHub: "+model.Payloadslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleSupport(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, model.Emoji+" "+model.ProductName+" Support: "+model.SupportLink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleInvite(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	if model.InviteLink == models.Default.InviteLink {
		err := c.reply(s, i, model.Emoji+"Invite Link: "+model.InviteLink)
		if err != nil {
			log.Print("[!] Error writing response: " + err.Error())
			return err
		}
	} else {
		err := c.reply(s, i, model.Emoji+" "+model.ProductName+" Channel Invite Link: "+model.InviteLink)
		if err != nil {
			log.Print("[!] Error writing response: " + err.Error())
			return err
		}
	}
	return nil
}

func (c *Commands) handleShop(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	model := c.getProductFromChannelID(i.ChannelID)
	err := c.reply(s, i, "Shop "+model.Emoji+" "+model.ProductName+" "+model.Shoplink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}
