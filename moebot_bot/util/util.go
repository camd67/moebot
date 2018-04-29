package util

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

const (
	CaseInsensitive = iota
	CaseSensitive
)

type SyncUIDByChannelMap struct {
	sync.RWMutex
	M map[string][]string
}

type SyncCooldownMap struct {
	sync.RWMutex
	M map[string]int64
}

func IntContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func StrContains(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.EqualFold(e, a) {
				return true
			}
		} else {
			if a == e {
				return true
			}
		}
	}
	return false
}

func StrContainsPrefix(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.HasPrefix(strings.ToUpper(a), strings.ToUpper(e)) {
				return true
			}
		} else {
			if strings.HasPrefix(a, e) {
				return true
			}
		}
	}
	return false
}

func MakeAlphaOnly(s string) string {
	reg := regexp.MustCompile("[^A-Za-z ]+")
	return reg.ReplaceAllString(s, "")
}

func NormalizeNewlines(s string) string {
	reg := regexp.MustCompile("(\r\n|\r|\n)")
	return reg.ReplaceAllString(s, "\n")
}

/*
Converts a user's ID into a mention.
This is useful when you don't have a User object, but want to mention them
*/
func UserIdToMention(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
}

func ExtractChannelIdFromString(message string) (id string, valid bool) {
	// channelIds go with the format of <#1234567>
	if len(message) < 2 || len(message) > 23 {
		return "", false
	}
	id = message[2 : len(message)-1]
	_, err := strconv.Atoi(id)
	return id, err == nil
}

func MakeStringBold(s string) string {
	return "**" + s + "**"
}

func MakeStringItalic(s string) string {
	return "_" + s + "_"
}

func MakeStringStrikethrough(s string) string {
	return "~~" + s + "~~"
}

func FindRoleByName(roles []*discordgo.Role, toFind string) *discordgo.Role {
	toFind = strings.ToUpper(toFind)
	for _, r := range roles {
		if strings.ToUpper(r.Name) == toFind {
			return r
		}
	}
	return nil
}

func FindRoleById(roles []*discordgo.Role, toFind string) *discordgo.Role {
	// for some reason roleIds have spaces in them...
	toFind = strings.TrimSpace(toFind)
	for _, r := range roles {
		if r.ID == toFind {
			return r
		}
	}
	return nil
}

func GetStringOrDefault(s sql.NullString) string {
	if s.Valid {
		return s.String
	} else {
		return "unknown"
	}
}

func GetInt64OrDefault(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	} else {
		return -1
	}
}

func UpdatePollVotes(poll *db.Poll, session *discordgo.Session) error {
	channel, err := db.ChannelQueryById(poll.ChannelId)
	if err != nil {
		return err
	}
	message, err := session.ChannelMessage(channel.ChannelUid, poll.MessageUid)
	if err != nil {
		return err
	}
	for _, o := range poll.Options {
		r := getReactionById(message, o.ReactionId)
		if r != nil {
			o.Votes = r.Count - 1
		}
	}
	return nil
}

func getReactionById(message *discordgo.Message, reactionId string) *discordgo.MessageReactions {
	for _, r := range message.Reactions {
		if reactionId == r.Emoji.Name {
			return r
		}
	}
	return nil
}

func OpenPollMessage(poll *db.Poll, user *discordgo.User) string {
	message := user.Mention() + " created "
	if poll.Title != "" {
		message += "the poll **" + poll.Title + "**!\n"
	} else {
		message += "a poll!\n"
	}
	for _, o := range poll.Options {
		message += ":" + o.ReactionName + ":  " + o.Description + "\n"
	}
	message += "Poll ID: " + strconv.Itoa(poll.Id)
	return message
}

func ClosePollMessage(poll *db.Poll, user *discordgo.User) string {
	var message string
	if poll.Open {
		if user.ID == poll.UserUid {
			message = user.Mention() + " closed their poll"
		} else {
			message = user.Mention() + " closed " + UserIdToMention(poll.UserUid) + "'s poll"
		}
		if poll.Title != "" {
			message += " **" + poll.Title + "**!\n"
		} else {
			message += "!\n"
		}
	} else {
		if poll.Title != "" {
			message = "Poll **" + poll.Title + "** is already closed!\n"
		} else {
			message = "This poll is already closed!"
		}
	}
	winners := pollWinners(poll)
	if len(winners) == 0 || winners[0].Votes == 0 {
		message += "There are no winners!"
		return message
	}
	if len(winners) > 1 {
		message += "Tied for first place:\n"
	} else {
		message += "Poll winner:\n"
	}
	for _, o := range winners {
		message += ":" + o.ReactionName + ":  " + o.Description + "\n"
	}
	message += "With " + strconv.Itoa(winners[0].Votes) + " votes!"
	return message
}

func pollWinners(poll *db.Poll) []*db.PollOption {
	var winningOptions []*db.PollOption
	maxVotes := 0
	for _, option := range poll.Options {
		if option.Votes > maxVotes {
			maxVotes = option.Votes
		}
	}

	for _, option := range poll.Options {
		if option.Votes == maxVotes {
			winningOptions = append(winningOptions, option)
		}
	}

	return winningOptions
}

func CreatePollOptions(options []string) []*db.PollOption {
	//TODO: Move to a database table?
	optionNames := []string{
		"regional_indicator_a",
		"regional_indicator_b",
		"regional_indicator_c",
		"regional_indicator_d",
		"regional_indicator_e",
		"regional_indicator_f",
		"regional_indicator_g",
		"regional_indicator_h",
		"regional_indicator_i",
		"regional_indicator_j",
		"regional_indicator_k",
		"regional_indicator_l",
		"regional_indicator_m",
		"regional_indicator_n",
		"regional_indicator_o",
		"regional_indicator_p",
		"regional_indicator_q",
		"regional_indicator_r",
		"regional_indicator_s",
		"regional_indicator_t",
		"regional_indicator_u",
		"regional_indicator_v",
		"regional_indicator_w",
		"regional_indicator_x",
		"regional_indicator_y",
		"regional_indicator_z",
	}
	optionIds := []string{
		"🇦",
		"🇧",
		"🇨",
		"🇩",
		"🇪",
		"🇫",
		"🇬",
		"🇭",
		"🇮",
		"🇯",
		"🇰",
		"🇱",
		"🇲",
		"🇳",
		"🇴",
		"🇵",
		"🇶",
		"🇷",
		"🇸",
		"🇹",
		"🇺",
		"🇻",
		"🇼",
		"🇽",
		"🇾",
		"🇿",
	}
	result := []*db.PollOption{}
	for i, s := range options {
		result = append(result, &db.PollOption{
			Description:  strings.Trim(s, " "),
			ReactionId:   optionIds[i],
			ReactionName: optionNames[i],
		})
	}
	return result
}

func GetSpoilerContents(messageParams []string) (title string, text string) {
	if messageParams == nil {
		return "", ""
	}
	reg := regexp.MustCompile("^(\\[.+?\\])")
	return strings.Replace(strings.Replace(reg.FindString(strings.Join(messageParams, " ")), "]", "", 1), "[", "", 1), reg.ReplaceAllString(strings.Join(messageParams, " "), "")
}

func MoveMessage(session *discordgo.Session, message *discordgo.Message, destChannelUid string, deleteOldPin bool) {
	if deleteOldPin {
		session.ChannelMessageDelete(message.ChannelID, message.ID)
	}
	var files []*discordgo.File
	for _, a := range message.Attachments {
		func() {
			response, err := http.Get(a.URL)
			if err != nil {
				return
			}
			defer response.Body.Close()
			b, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Println("Error reading from attachment response", err)
				return
			}
			files = append(files, &discordgo.File{
				Name:        a.Filename,
				Reader:      bytes.NewReader(b),
				ContentType: mime.TypeByExtension(filepath.Ext(a.Filename)),
			})
		}()
	}
	content := message.Author.Mention() + " posted the following message in <#" + message.ChannelID + ">:\n" + message.Content

	session.ChannelMessageSendComplex(destChannelUid, &discordgo.MessageSend{
		Content: content,
		Files:   files,
	})
}
