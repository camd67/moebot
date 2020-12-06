package commands

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PollsHandler struct {
	pollsList models.PollSlice
}

func NewPollsHandler() *PollsHandler {
	h := &PollsHandler{}
	h.loadFromDb()
	return h
}

func (handler *PollsHandler) loadFromDb() {
	polls, _ := db.PollsOpenQuery()
	handler.pollsList = polls
}

func (handler *PollsHandler) openPoll(pack *CommPackage) {
	var options []string
	var title string
	for i := 0; i < len(pack.params); i++ {
		if pack.params[i] == "-options" {
			options = parseOptions(pack.params[i+1:])
		}
		if pack.params[i] == "-title" {
			title = parseTitle(pack.params[i+1:])
		}
	}
	if len(options) <= 1 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you must specify at least two options to create a poll.")
		return
	}
	if len(options) > 25 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there can only be a maximum of 25 options per poll.")
		return
	}
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem creating the poll. Please try again.")
		return
	}
	channel, err := db.ChannelQueryOrInsert(pack.channel.ID, server)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem creating the poll. Please try again.")
		return
	}
	poll := &models.Poll{
		Title:     title,
		UserUID:   pack.message.Author.ID,
		ChannelID: channel.ID,
		Open:      true,
	}
	err = db.PollAdd(poll)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem creating the poll. Please try again.")
		return
	}
	poll = db.PollSetOptions(poll, createPollOptions(poll, options))
	db.PollOptionAdd(poll)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem creating the poll. Please try again.")
		return
	}
	message, _ := pack.session.ChannelMessageSend(pack.channel.ID, openPollMessage(poll, pack.message.Author))
	for _, o := range poll.R.PollOptions {
		err = pack.session.MessageReactionAdd(pack.channel.ID, message.ID, o.ReactionUID)
		if err != nil {
			log.Println("Cannot add reaction to poll message", err)
		}
	}
	poll.MessageUID.SetValid(message.ID)
	err = db.PollSetMessageId(poll)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem updating the poll. Please delete and create it again.")
	}
	handler.pollsList = append(handler.pollsList, poll)
}

func parseOptions(params []string) []string {
	for i := 0; i < len(params); i++ {
		if params[i][0] == '-' {
			return strings.Split(strings.Join(params[:i], " "), ",")
		}
	}
	return strings.Split(strings.Join(params, " "), ",")
}

func parseTitle(params []string) string {
	for i := 0; i < len(params); i++ {
		if params[i][0] == '-' {
			return strings.Join(params[:i], " ")
		}
	}
	return strings.Join(params, " ")
}

func (handler *PollsHandler) closePoll(pack *CommPackage) {
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you have to specify a valid ID for the poll")
		return
	}
	id, err := strconv.Atoi(pack.params[1])
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem parsing the poll ID, please check if it's a valid ID")
		return
	}
	poll := handler.pollFromId(id)
	if poll == nil {
		poll, err = db.PollQuery(id)
		if err == sql.ErrNoRows {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there is no valid poll with the given ID")
			return
		} else if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem retreiving the poll with the given ID")
			return
		}
		handler.pollsList = append(handler.pollsList, poll)
	}
	channel, err := db.ChannelQueryById(poll.ChannelID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem retrieving poll data")
		return
	}
	if channel.ChannelUID != pack.channel.ID {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you can't close a poll opened in another channel")
		return
	}
	if !poll.Open {
		pack.session.ChannelMessageSend(pack.channel.ID, closePollMessage(poll, pack.message.Author))
		return
	}
	err = updatePollVotes(poll, pack.session)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem retrieving the votes count for the given Poll")
		return
	}
	db.PollOptionUpdateVotes(poll)
	err = db.PollClose(id)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem closing the poll.")
		return
	}
	pack.session.ChannelMessageSend(pack.channel.ID, closePollMessage(poll, pack.message.Author))
	poll.Open = false
}

func (handler *PollsHandler) pollFromId(id int) *models.Poll {
	for _, p := range handler.pollsList {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (handler *PollsHandler) checkSingleVote(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	var err error
	for _, p := range handler.pollsList {
		if p.MessageUID.String == reactionAdd.MessageID {
			p.R.PollOptions, err = db.PollOptionQuery(p.ID)
			if err != nil {
				log.Println("Cannot retrieve poll options informations", err)
				return
			}
			//If the user is reacting to a poll, we check if he has already cast a vote and remove it
			handler.handleSingleVote(session, p, reactionAdd)
			return
		}
	}
}

func (handler *PollsHandler) handleSingleVote(session *discordgo.Session, poll *models.Poll, reactionAdd *discordgo.MessageReactionAdd) {
	channel, err := db.ChannelQueryById(poll.ChannelID)
	if err != nil {
		log.Println("Cannot retrieve poll channel informations", err)
		return
	}
	message, err := session.ChannelMessage(channel.ChannelUID, poll.MessageUID.String)
	if err != nil {
		log.Println("Cannot retrieve poll message informations", err)
		return
	}
	if message.Author.ID == reactionAdd.UserID {
		return //The bot is modifying its own reactions
	}
	for _, r := range message.Reactions {
		if !reactionIsOption(poll.R.PollOptions, r.Emoji.Name) {
			continue
		}
		//Getting a list of users for every reaction
		users, err := session.MessageReactions(channel.ChannelUID, poll.MessageUID.String, r.Emoji.Name, 100)
		if err != nil {
			log.Println("Cannot retrieve reaction informations", err)
			return
		}
		for _, u := range users {
			//If the user has other votes, we remove them
			if u.ID == reactionAdd.UserID && r.Emoji.Name != reactionAdd.Emoji.Name && reactionIsOption(poll.R.PollOptions, reactionAdd.Emoji.Name) {
				session.MessageReactionRemove(channel.ChannelUID, poll.MessageUID.String, r.Emoji.Name, u.ID)
				break
			}
		}
	}
}

func reactionIsOption(options models.PollOptionSlice, emojiID string) bool {
	for _, o := range options {
		if o.ReactionUID == emojiID {
			return true
		}
	}
	return false
}

func updatePollVotes(poll *models.Poll, session *discordgo.Session) error {
	channel, err := db.ChannelQueryById(poll.ChannelID)
	if err != nil {
		return err
	}
	message, err := session.ChannelMessage(channel.ChannelUID, poll.MessageUID.String)
	if err != nil {
		return err
	}
	for _, o := range poll.R.PollOptions {
		r := moeDiscord.GetReactionByName(message, o.ReactionUID)
		if r != nil {
			o.Votes = r.Count - 1
		}
	}
	return nil
}

func openPollMessage(poll *models.Poll, user *discordgo.User) string {
	message := user.Mention() + " created "
	if poll.Title != "" {
		message += "the poll **" + poll.Title + "**!\n"
	} else {
		message += "a poll!\n"
	}
	for _, o := range poll.R.PollOptions {
		message += ":" + o.ReactionName + ":  " + o.Description + "\n"
	}
	message += "Poll ID: " + strconv.Itoa(poll.ID)
	return message
}

func closePollMessage(poll *models.Poll, user *discordgo.User) string {
	var message string
	if poll.Open {
		if user.ID == poll.UserUID {
			message = user.Mention() + " closed their poll"
		} else {
			message = user.Mention() + " closed " + util.UserIdToMention(poll.UserUID) + "'s poll"
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

func pollWinners(poll *models.Poll) models.PollOptionSlice {
	var winningOptions models.PollOptionSlice
	maxVotes := 0
	for _, option := range poll.R.PollOptions {
		if option.Votes > maxVotes {
			maxVotes = option.Votes
		}
	}

	for _, option := range poll.R.PollOptions {
		if option.Votes == maxVotes {
			winningOptions = append(winningOptions, option)
		}
	}

	return winningOptions
}

func createPollOptions(poll *models.Poll, options []string) models.PollOptionSlice {
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
	var result models.PollOptionSlice
	for i, s := range options {
		opt := &models.PollOption{
			Description:  strings.Trim(s, " "),
			ReactionUID:  optionIds[i],
			ReactionName: optionNames[i],
		}
		opt.PollID.SetValid(poll.ID)
		result = append(result, opt)
	}
	return result
}
