package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/camd67/moebot/moebot_bot/bot/commands"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	version = "0.3"
)

var (
	// need to move this to the db
	allowedNonComServers = []string{"378336255030722570", "93799773856862208"}
	ComPrefix            string
	Config               = make(map[string]string)
	pinnedMessages       = make(map[string][]string)
	commandsMap          = make(map[string]commands.Command)
)

func SetupMoebot(session *discordgo.Session) {
	db.SetupDatabase(Config["dbPass"], Config["moeDataPass"])
	setupCommands(session)
	addHandlers(session)
}

func setupCommands(session *discordgo.Session) {
	checker := commands.PermissionChecker{MasterId: Config["masterId"]}
	commandsMap["TEAM"] = commands.TeamCommand{}
	commandsMap["ROLE"] = commands.RoleCommand{}
	commandsMap["RANK"] = commands.RankCommand{}
	commandsMap["NSFW"] = commands.NsfwCommand{}
	commandsMap["HELP"] = commands.HelpCommand{ComPrefix: ComPrefix}
	commandsMap["CHANGELOG"] = commands.ChangelogCommand{Version: version}
	commandsMap["RAFFLE"] = commands.RaffleCommand{MasterId: Config["masterId"], DebugChannel: Config["debugChannel"]}
	commandsMap["SUBMIT"] = commands.SubmitCommand{ComPrefix: ComPrefix}
	commandsMap["ECHO"] = commands.EchoCommand{Checker: checker}
	commandsMap["PERMIT"] = commands.PermitCommand{Checker: checker}
	commandsMap["CUSTOM"] = commands.CustomCommand{Checker: checker, ComPrefix: ComPrefix}
	commandsMap["PING"] = commands.PingCommand{}
	commandsMap["SPOILER"] = commands.SpoilerCommand{}
	commandsMap["POLL"] = commands.PollCommand{Checker: checker, PollsHandler: commands.NewPollsHandler()}
	commandsMap["TOGGLEMENTION"] = commands.MentionCommand{Checker: checker}
	commandsMap["SERVER"] = commands.ServerCommand{Checker: checker}
	commandsMap["PROFILE"] = commands.ProfileCommand{}
	commandsMap["PINMOVE"] = commands.PinMoveCommand{Checker: checker}

	for _, com := range commandsMap {
		com.Setup(session)
	}
}

func addHandlers(discord *discordgo.Session) {
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(guildMemberAdd)
	discord.AddHandler(messageReactionAdd)
	for _, com := range commandsMap {
		for _, h := range com.EventHandlers() {
			discord.AddHandler(h)
		}
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
		runCommand(session, message.Message, guild, channel, member)
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
}

func messageReactionAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	// should make some local caches for channels and guilds...
	channel, err := session.Channel(reactionAdd.ChannelID)
	if err != nil {
		log.Println("Error trying to get channel", err)
		return
	}

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

func ready(session *discordgo.Session, event *discordgo.Ready) {
	status := ComPrefix + " help"
	err := session.UpdateStatus(0, status)
	if err != nil {
		log.Println("Error setting moebot status", err)
	}
	log.Println("Set moebot's status to", status)
}

func runCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commandsMap[command]; commPresent {
		params := messageParts[2:]
		log.Println("Processing command: " + command + " from user: {" + fmt.Sprintf("%+v", message.Author) + "}| With Params:{" + strings.Join(params, ",") + "}")
		session.ChannelTyping(message.ChannelID)
		pack := commands.NewCommPackage(session, message, guild, member, channel, params)
		commFunc.Execute(&pack)
	}
}
