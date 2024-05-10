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
	return strings.Contains(strings.Join(product.Aliases[:], " "), args[1]) ||
		strings.Contains(strings.Join(product.Aliases, " "), strings.Join(args[1:3], " ")) ||
		strings.Contains(strings.Join(product.Aliases, " "), strings.Join(args[1:4], " "))
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

func (c *Commands) handleAskDocs(s *discordgo.Session, m *discordgo.MessageCreate) error {
	args := strings.Split(m.Content, " ")

	if len(args) < 2 {
		return errors.New("missing question")
	}

	input := m.Content[strings.Index(m.Content, " "):len(m.Content)]

	model := c.getProductFromChannelID(m.ChannelID)

	// Default guard
	if model.SpaceID == models.Default.SpaceID {
		_, err := s.ChannelMessageSend(m.ChannelID, "AI Search not enabled for this product, but you can still look for yourself here: "+model.Docslink)
		if err != nil {
			return err
		}
		return nil
	}

	pre := "<:thonk:666272807524761620> Asking the"
	post := "docs..."
	askEndpoint := models.GITBOOKAPI + model.SpaceID + models.SEARCHENDPOINT
	searchingIn := pre + " " + model.Emoji + " " + model.ProductName + " " + post
	_, err := s.ChannelMessageSend(m.ChannelID, searchingIn)
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
	_, err = s.ChannelMessageSend(m.ChannelID, prepend+resp.Answer.Text+"\n\n"+model.Docslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}

	return nil
}

func (c *Commands) handleDocs(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" Docs: "+model.Docslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleGithub(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" Github: "+model.Githublink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handlePayloads(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" Github: "+model.Githublink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	_, err = s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" PayloadHub: "+model.Payloadslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handlePayloadHub(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" PayloadHub: "+model.Payloadslink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleSupport(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" Support: "+model.SupportLink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}

func (c *Commands) handleInvite(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	if model.InviteLink == models.Default.InviteLink {
		_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+"Invite Link: "+model.InviteLink)
		if err != nil {
			log.Print("[!] Error writing response: " + err.Error())
			return err
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, model.Emoji+" "+model.ProductName+" Channel Invite Link: "+model.InviteLink)
		if err != nil {
			log.Print("[!] Error writing response: " + err.Error())
			return err
		}
	}
	return nil
}

func (c *Commands) handleShop(s *discordgo.Session, m *discordgo.MessageCreate) error {
	model := c.getProductFromChannelID(m.ChannelID)
	_, err := s.ChannelMessageSend(m.ChannelID, "Shop "+model.Emoji+" "+model.ProductName+" "+model.Shoplink)
	if err != nil {
		log.Print("[!] Error writing response: " + err.Error())
		return err
	}
	return nil
}
