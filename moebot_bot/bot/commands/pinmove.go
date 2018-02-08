package commands

import (
	"database/sql"
	"errors"
	"log"
	"mime"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PinMoveCommand struct {
	Checker        *PermissionChecker
	pinnedMessages map[string][]string
}

func (pc *PinMoveCommand) Execute(pack *CommPackage) {
	if !pc.Checker.HasModPerm(pack.message.Author.ID, pack.member.Roles) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
	if len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you need to specify at least a valid channel.")
		return
	}
	var err error
	var pinChannel string
	regNumbers := regexp.MustCompile("\\d+")
	enableText := false
	for i := 0; i < len(pack.params)-1; i++ {
		if pack.params[i] == "-sendTo" {
			pinChannel = regNumbers.FindString(pack.params[i+1])
		}
		if pack.params[i] == "-text" {
			enableText = true
		}
	}
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if pinChannel == "" && (!server.DefaultPinChannelId.Valid || server.DefaultPinChannelId.Int64 == 0) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there is no default destination channel set. You need to specify at least a valid destination channel.")
		return
	}
	sourceChannelUid := regNumbers.FindString(pack.params[len(pack.params)-1])
	if pinChannel != "" {
		if err = pc.newPinChannel(pinChannel, server, pack); err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, err.Error())
			return
		}
	}
	var pinEnabled bool
	if pinEnabled, err = togglePin(sourceChannelUid, enableText, server, pack); err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, err.Error())
		return
	}
	message := "Message move on pin has been "
	if pinEnabled {
		message += "enabled"
	} else {
		message += "disabled"
	}
	message += " on channel <#" + sourceChannelUid + ">"
	pack.session.ChannelMessageSend(pack.channel.ID, message)
}

func (pc *PinMoveCommand) Setup(session *discordgo.Session) {
	pc.pinnedMessages = make(map[string][]string)
	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Println("Error loading guilds, some functions may not work correctly.", err)
		return
	}
	log.Println("Number of guilds: " + strconv.Itoa(len(guilds)))
	for _, guild := range guilds {
		pc.loadGuild(session, guild)
	}
}

func (pc *PinMoveCommand) EventHandlers() []interface{} {
	return []interface{}{pc.channelMovePinsUpdate}
}

func (pc *PinMoveCommand) loadGuild(session *discordgo.Session, guild *discordgo.UserGuild) {
	server, err := db.ServerQueryOrInsert(guild.ID)
	if err != nil {
		log.Println("Error creating/retrieving server during loading", err)
		return
	}
	channels, err := session.GuildChannels(guild.ID)
	if err != nil {
		log.Println("Error retrieving channels during loading", err)
		return
	}
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText { //only loading text channels for now
			pc.loadChannel(session, &server, channel)
		}
	}
}

func (pc *PinMoveCommand) loadChannel(session *discordgo.Session, server *db.Server, channel *discordgo.Channel) {
	log.Println("Loading channel: " + channel.Name + " (" + channel.ID + ")")

	_, err := db.ChannelQueryOrInsert(channel.ID, server)
	if err != nil {
		log.Println("Error creating/retrieving channel during loading", err)
		return
	}
	pc.loadPinnedMessages(session, channel)
}

func (pc *PinMoveCommand) loadPinnedMessages(session *discordgo.Session, channel *discordgo.Channel) {
	pc.pinnedMessages[channel.ID] = []string{}
	messages, err := session.ChannelMessagesPinned(channel.ID)
	if err != nil {
		log.Println("Error retrieving pinned channel messages", err)
	}
	log.Println("Loading pinned messages > " + strconv.Itoa(len(messages)))
	for _, message := range messages {
		pc.pinnedMessages[channel.ID] = append(pc.pinnedMessages[channel.ID], message.ID)
	}
}

func (pc *PinMoveCommand) newPinChannel(newPinChannelUid string, server db.Server, pack *CommPackage) error {
	var newPinChannel *discordgo.Channel
	var err error
	for _, c := range pack.guild.Channels {
		if c.ID == newPinChannelUid {
			newPinChannel = c
			break
		}
	}
	if newPinChannel == nil {
		return errors.New("Sorry, you need to specify a valid destination channel")
	}
	var currentPinChannel *db.Channel
	if server.DefaultPinChannelId.Valid && server.DefaultPinChannelId.Int64 > 0 {
		currentPinChannel, err = db.ChannelQueryById(int(server.DefaultPinChannelId.Int64))
		if err != nil && err != sql.ErrNoRows {
			return errors.New("Sorry, there was a problem retrieving the current pin channel")
		}
	}
	if currentPinChannel == nil || currentPinChannel.ChannelUid != newPinChannel.ID {
		dbNewPinChannel, err := db.ChannelQueryOrInsert(newPinChannel.ID, &server)
		if err != nil {
			return errors.New("Sorry, there was a problem retrieving the new pin channel")
		}
		err = db.ServerSetDefaultPinChannel(server.Id, dbNewPinChannel.Id)
		if err != nil {
			return errors.New("Sorry, there was a problem setting the new pin channel")
		}
		server.DefaultPinChannelId.Scan(dbNewPinChannel.Id)
	}
	return nil
}

func togglePin(sourceChannelUid string, enableTextPins bool, server db.Server, pack *CommPackage) (bool, error) {
	var sourceChannel *discordgo.Channel
	for _, c := range pack.guild.Channels {
		if c.ID == sourceChannelUid {
			sourceChannel = c
			break
		}
	}
	if sourceChannel == nil {
		return false, errors.New("Sorry, you need to specify a valid source channel")
	}
	dbSourceChannel, err := db.ChannelQueryOrInsert(sourceChannel.ID, &server)
	if err != nil {
		return false, errors.New("Sorry, there was a problem retrieving the source channel")
	}
	err = db.ChannelSetPin(dbSourceChannel.Id, !dbSourceChannel.MovePins, enableTextPins)
	if err != nil {
		return false, errors.New("Sorry, there was a problem setting the pin status")
	}
	return !dbSourceChannel.MovePins, nil
}

func (pc *PinMoveCommand) channelMovePinsUpdate(session *discordgo.Session, pinsUpdate *discordgo.ChannelPinsUpdate) {
	channel, err := session.Channel(pinsUpdate.ChannelID)
	if err != nil {
		log.Println("Error while retrieving channel by UID", err)
		return
	}
	server, err := db.ServerQueryOrInsert(channel.GuildID)
	if err != nil {
		log.Println("Error while retrieving server from database", err)
		return
	}
	if !server.DefaultPinChannelId.Valid || server.DefaultPinChannelId.Int64 == 0 {
		return
	}
	dbChannel, err := db.ChannelQueryOrInsert(pinsUpdate.ChannelID, &server)
	if err != nil {
		log.Println("Error while retrieving source channel from database", err)
		return
	}
	if !dbChannel.MovePins {
		return
	}
	dbDestChannel, err := db.ChannelQueryById(int(server.DefaultPinChannelId.Int64))
	if err != nil {
		log.Println("Error while retrieving destination channel from database", err)
		return
	}
	newPinnedMessages, err := pc.getUpdatePinnedMessages(session, pinsUpdate.ChannelID)
	if err != nil {
		log.Println("Error while retrieving new pinned messages", err)
		return
	}
	if len(newPinnedMessages) == 0 || len(newPinnedMessages) > 1 {
		return //removed pin or the bot is not in sync with the server, abort pinning operation
	}
	newPinnedMessage := newPinnedMessages[0]
	moveMessage := false
	for _, a := range newPinnedMessage.Attachments { //image from direct upload
		if strings.Contains(mime.TypeByExtension(filepath.Ext(a.Filename)), "image") {
			moveMessage = true
			break
		}
	}

	if !moveMessage && len(newPinnedMessage.Embeds) == 1 { //image from link
		if newPinnedMessage.Embeds[0].Type == "image" {
			moveMessage = true
		}
	}
	if len(newPinnedMessage.Attachments) == 0 && len(newPinnedMessage.Embeds) == 0 && dbChannel.MoveTextPins {
		moveMessage = true
	}
	if moveMessage {
		util.MoveMessage(session, newPinnedMessage, dbDestChannel.ChannelUid)
	}
}

func (pc *PinMoveCommand) getUpdatePinnedMessages(session *discordgo.Session, channelId string) ([]*discordgo.Message, error) {
	result := []*discordgo.Message{}
	currentPinnedMessages, err := session.ChannelMessagesPinned(channelId)
	messagesId := []string{}
	if err != nil {
		return result, err
	}
	for _, m := range currentPinnedMessages {
		if !pc.pinnedMessageAlreadyLoaded(m.ID, channelId) {
			result = append(result, m)
		}
		messagesId = append(messagesId, m.ID)
	}
	pc.pinnedMessages[channelId] = messagesId //refreshes pinned messages in case of messages removed from pins
	return result, nil
}

func (pc *PinMoveCommand) pinnedMessageAlreadyLoaded(messageId string, channelId string) bool {
	for _, m := range pc.pinnedMessages[channelId] {
		if messageId == m {
			return true
		}
	}
	return false
}
