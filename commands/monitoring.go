package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxtrot/scuzzy/models"
	"hash/fnv"
	"time"
)

var Tracking map[uint32]models.TrackedMessage
var windowToConsiderIrrelevant = time.Minute * 5
var windowToConsiderDuplicate = time.Second * 30
var crossChannelSpamCountThreshold = 1
var spamCountThreshold = 3

func hashContent(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return h.Sum32()
}

func track(m *discordgo.Message) bool {
	//todo ignore self bot
	hash := hashContent(m.Content)
	record, exists := Tracking[hash]
	if !exists {
		newMessage := models.TrackedMessage{
			Author:         m.Author,
			Instances:      map[string]models.CachedMessage{},
			MessageContent: m.Content,
		}
		newMessage.Instances[m.ChannelID] = models.CachedMessage{
			ID:        m.ID,
			CreatedAt: time.Now(),
			Count:     1,
		}
		Tracking[hash] = newMessage
	} else {
		//ModerationReason := "[DELETED + TIMEOUT] Multiple of the same exact message sent repeatedly in the same channel within a short period of time is considered spam and was removed"
		chanMessage, existsInChannel := record.Instances[m.ChannelID]
		if existsInChannel {
			chanMessage.Count++
			record.Instances[m.ChannelID] = chanMessage
			Tracking[hash] = record

			if time.Since(chanMessage.CreatedAt) > windowToConsiderIrrelevant {
				// this tracking has "expired"
				delete(record.Instances, m.ChannelID)
				Tracking[hash] = record
				return false
			}
			countViolation := chanMessage.Count > spamCountThreshold
			frequencyViolation := time.Since(chanMessage.CreatedAt) < windowToConsiderDuplicate
			switch true {
			case countViolation && frequencyViolation:
				// remove obviously spam qualified by both frequency and quantity
				cleanupDuplicates(record.Instances)
				timeoutUser()
				return true
			case countViolation:
				cleanupDuplicates(record.Instances)
				return true
			case frequencyViolation:
				cleanupDuplicates(record.Instances)
				return true
			default:
				warnUser()
				return false
			}
		} else {
			channelCount := 0
			smallestTimeSince := time.Duration(10000000)
			for _, _ = range record.Instances {
				channelCount++
				since := time.Since(chanMessage.CreatedAt)
				if smallestTimeSince < since {
					smallestTimeSince = since
				}
			}
			countViolation := channelCount > crossChannelSpamCountThreshold
			frequencyViolation := smallestTimeSince < windowToConsiderDuplicate
			switch true {
			case countViolation && frequencyViolation:
				// remove obviously spam qualified by both frequency and quantity
				cleanupDuplicates(record.Instances)
				timeoutUser()
				return true
			case countViolation:
				cleanupDuplicates(record.Instances)
				return true
			case frequencyViolation:
				cleanupDuplicates(record.Instances)
				return true
			default:
				warnUser()
				return false
			}
		}
	}
	return false
}

func cleanupDuplicates(instances map[string]models.CachedMessage) {

}

func warnUser() {

}

func timeoutUser() {

}
