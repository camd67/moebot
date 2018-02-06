package bot

import (
	"fmt"
	"log"
	"math/rand"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	version = "0.3"
)

type commPackage struct {
	session *discordgo.Session
	message *discordgo.Message
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	params  []string
}

var (
	// need to move this to the db
	allowedNonComServers = []string{"378336255030722570", "93799773856862208"}
	ComPrefix            string
	Config               = make(map[string]string)
	pollsHandler         = new(PollsHandler)
	pinnedMessages       = make(map[string][]string)
)

func SetupMoebot(session *discordgo.Session) {
	addHandlers(session)
	db.SetupDatabase(Config["dbPass"], Config["moeDataPass"])
	pollsHandler.loadFromDb()
	loadGuilds(session)
}

func addHandlers(discord *discordgo.Session) {
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(guildMemberAdd)
	discord.AddHandler(messageReactionAdd)
	discord.AddHandler(channelPinsUpdate)
}

func loadGuilds(session *discordgo.Session) {
	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Println("Error loading guilds, some functions may not work correctly.", err)
		return
	}
	log.Println("Number of guilds: " + strconv.Itoa(len(guilds)))
	for _, guild := range guilds {
		loadGuild(session, guild)
	}
}

func loadGuild(session *discordgo.Session, guild *discordgo.UserGuild) {
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
			loadChannel(session, &server, channel)
		}
	}
}

func loadChannel(session *discordgo.Session, server *db.Server, channel *discordgo.Channel) {
	log.Println("Loading channel: " + channel.Name + " (" + channel.ID + ")")

	_, err := db.ChannelQueryOrInsert(channel.ID, server)
	if err != nil {
		log.Println("Error creating/retrieving channel during loading", err)
		return
	}
	loadPinnedMessages(session, channel)
}

func loadPinnedMessages(session *discordgo.Session, channel *discordgo.Channel) {
	pinnedMessages[channel.ID] = []string{}
	messages, err := session.ChannelMessagesPinned(channel.ID)
	if err != nil {
		log.Println("Error retrieving pinned channel messages", err)
	}
	log.Println("Loading pinned messages > " + strconv.Itoa(len(messages)))
	for _, message := range messages {
		pinnedMessages[channel.ID] = append(pinnedMessages[channel.ID], message.ID)
	}
}

func guildMemberAdd(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	guild, err := session.Guild(member.GuildID)
	// temp ignore IHG + Salt
	if guild.ID == "84724034129907712" || guild.ID == "93799773856862208" {
		return
	}
	if err != nil {
		log.Println("Error when getting guild for user when joining guild", member)
		return
	}

	var starterRole *discordgo.Role
	for _, guildRole := range guild.Roles {
		if guildRole.Name == "Bell" {
			starterRole = guildRole
			break
		}
	}
	if starterRole == nil {
		log.Println("ERROR! Unable to find starter role for guild " + guild.Name)
		return
	}
	dmChannel, err := session.UserChannelCreate(member.User.ID)
	if err != nil {
		log.Println("ERROR! Unable to make DM channel with userID ", member.User.ID)
		return
	}
	session.ChannelMessageSend(dmChannel.ID, "Hello "+member.User.Mention()+" and welcome to "+guild.Name+"! Please read the #rules channel for more information")
	session.GuildMemberRoleAdd(member.GuildID, member.User.ID, starterRole.ID)
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// bail out if we have any messages we want to ignore such as bot messages
	if message.Author.ID == session.State.User.ID || message.Author.Bot {
		return
	}

	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		// missing channel
		log.Println("ERROR! Unable to get guild in messageCreate ", err, channel)
		return
	}

	guild, err := session.Guild(channel.GuildID)
	if err != nil {
		log.Println("ERROR! Unable to get guild in messageCreate ", err, guild)
		return
	}

	member, err := session.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		log.Println("ERROR! Unable to get member in messageCreate ", err, message)
		return
	}
	// temp ignore IHG
	if err != nil || guild.ID == "84724034129907712" {
		// missing guild
		return
	}
	// should change this to store the ID of the starting role
	var starterRole *discordgo.Role
	var oldStarterRole *discordgo.Role
	for _, guildRole := range guild.Roles {
		if strings.EqualFold(guildRole.Name, "Red Whistle") {
			starterRole = guildRole
		} else if strings.EqualFold(guildRole.Name, "Bell") {
			oldStarterRole = guildRole
		}
	}

	// ignore some common bot prefixes
	if !(strings.HasPrefix(message.Content, "->") || strings.HasPrefix(message.Content, "~") || strings.HasPrefix(message.Content, ComPrefix)) {
		changedUsers, err := handleVeteranMessage(member, guild.ID)
		if err != nil {
			session.ChannelMessageSend(Config["debugChannel"], fmt.Sprint("An error occurred when trying to update veteran users ", err))
		} else {
			for _, user := range changedUsers {
				session.ChannelMessageSend(user.SendTo, "Congrats "+util.UserIdToMention(user.UserUid)+" you can become a server veteran! Type `"+
					ComPrefix+" role veteran` In this channel.")
			}
		}
	}

	if strings.HasPrefix(message.Content, ComPrefix) {
		// should add a check here for command spam
		if oldStarterRole != nil && util.StrContains(member.Roles, oldStarterRole.ID, util.CaseSensitive) {
			// bail out to prevent any new users from using bot commands
			return
		}
		RunCommand(session, message.Message, guild, channel, member)
	} else if util.StrContains(allowedNonComServers, guild.ID, util.CaseSensitive) && oldStarterRole != nil {
		// message may have other bot related commands, but not with a prefix
		readRules := "I want to venture into the abyss"
		sanitizedMessage := util.MakeAlphaOnly(message.Content)
		if strings.HasPrefix(strings.ToUpper(sanitizedMessage), strings.ToUpper(readRules)) {
			if !util.StrContains(member.Roles, oldStarterRole.ID, util.CaseSensitive) {
				// bail out if the user doesn't have the first starter role. Preventing any duplicate role creations/DMs
				return
			}
			if starterRole == nil || oldStarterRole == nil {
				log.Println("ERROR! Unable to find roles for guild " + guild.Name)
				return
			}
			session.ChannelMessageSend(message.ChannelID, "Welcome "+message.Author.Mention()+"! We hope you enjoy your stay in our Discord server! Make sure to head over #announcements and #bot-stuff to see whats new and get your roles!")
			session.GuildMemberRoleAdd(guild.ID, member.User.ID, starterRole.ID)
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, oldStarterRole.ID)
			log.Println("Updated user <" + member.User.Username + "> after reading the rules")
		}
	}
	// distribute tickets
	// temporarily disable ticket distribution
	//distributeTickets(guild, message, session, messageTime)
}

func messageReactionAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	// should make some local caches for channels and guilds...
	channel, err := session.Channel(reactionAdd.ChannelID)
	if err != nil {
		log.Println("Error trying to get channel", err)
		return
	}

	pollsHandler.checkSingleVote(session, reactionAdd)
	changedUsers, err := handleVeteranReaction(reactionAdd.UserID, channel.GuildID)
	if err != nil {
		session.ChannelMessageSend(Config["debugChannel"], fmt.Sprint("An error occurred when trying to update veteran users ", err))
	} else {
		for _, user := range changedUsers {
			session.ChannelMessageSend(user.SendTo, "Congrats "+util.UserIdToMention(user.UserUid)+" you can become a server veteran! Type `"+
				ComPrefix+" role veteran` In this channel.")
		}
	}
}

func reactionIsOption(options []*db.PollOption, emojiID string) bool {
	for _, o := range options {
		if o.ReactionId == emojiID {
			return true
		}
	}
	return false
}

func channelPinsUpdate(session *discordgo.Session, pinsUpdate *discordgo.ChannelPinsUpdate) {
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
	newPinnedMessages, err := getUpdatePinnedMessages(session, pinsUpdate.ChannelID)
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

func getUpdatePinnedMessages(session *discordgo.Session, channelId string) ([]*discordgo.Message, error) {
	result := []*discordgo.Message{}
	currentPinnedMessages, err := session.ChannelMessagesPinned(channelId)
	messagesId := []string{}
	if err != nil {
		return result, err
	}
	for _, m := range currentPinnedMessages {
		if !pinnedMessageAlreadyLoaded(m.ID, channelId) {
			result = append(result, m)
		}
		messagesId = append(messagesId, m.ID)
	}
	pinnedMessages[channelId] = messagesId //refreshes pinned messages in case of messages removed from pins
	return result, nil
}

func pinnedMessageAlreadyLoaded(messageId string, channelId string) bool {
	for _, m := range pinnedMessages[channelId] {
		if messageId == m {
			return true
		}
	}
	return false
}

func distributeTickets(guild *discordgo.Guild, message *discordgo.MessageCreate, session *discordgo.Session, messageTime time.Time) {
	if false {
		const maxChance = 100
		const ticketChance = 5
		if rand.Int()%maxChance <= ticketChance {
			raffles, err := db.RaffleEntryQuery(message.Author.ID, guild.ID)
			if err != nil {
				session.ChannelMessageSend(Config["debugChannel"], "Error loading raffle information during ticket distribution"+fmt.Sprintf("%+v | %+v", guild, message))
				return
			}

			if len(raffles) != 1 {
				// if they're not in the raffle, bail out
				return
			}
			r := raffles[0]
			// check to see if their last ticket time is more than the cooldown
			if r.LastTicketUpdate+ticketCooldown > messageTime.UnixNano() {
				// they haven't waited enough, bail
				return
			}
			// they've won a ticket and passed the timestamp check, let them know and update db
			r.LastTicketUpdate = messageTime.UnixNano()
			db.RaffleEntryUpdate(r, 1)
			currTickets := r.TicketCount + 1
			session.ChannelMessageSend("378680855339728918", message.Author.Mention()+", congrats! You just earned another ticket! Your current tickets are: "+strconv.Itoa(currTickets))
		}
	}
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	status := ComPrefix + " help"
	err := session.UpdateStatus(0, status)
	if err != nil {
		log.Println("Error setting moebot status", err)
	}
	log.Println("Set moebot's status to", status)
}
