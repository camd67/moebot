package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/camd67/moebot/moebot_bot/bot/commands"
	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	version = "0.4.2"
)

var (
	checker            permissions.PermissionChecker
	ComPrefix          string
	Config             = make(map[string]string)
	operations         []interface{}
	commandsMap        = make(map[string]commands.Command)
	masterId           string
	masterDebugChannel string
)

/*
Run through initial setup steps for Moebot. This is all that's necessary to setup Moebot for use
*/
func SetupMoebot(session *discordgo.Session) {
	masterId = Config["masterId"]
	checker = permissions.PermissionChecker{MasterId: masterId}
	masterDebugChannel = Config["debugChannel"]
	db.SetupDatabase(Config["dbPass"], Config["moeDataPass"])
	addGlobalHandlers(session)
	setupOperations(session)
}

/*
Create all the operations to handle commands and events within moebot.
Whenever a new operation, command, or event is added it should be added to this list
*/
func setupOperations(session *discordgo.Session) {
	operations = []interface{}{
		&commands.RoleCommand{},
		&commands.RoleSetCommand{ComPrefix: ComPrefix},
		&commands.GroupSetCommand{ComPrefix: ComPrefix},
		&commands.HelpCommand{ComPrefix: ComPrefix, Commands: getCommands, Checker: checker}, //using a delegate here because it will remain accurate regardless of what gets added to operations
		&commands.ChangelogCommand{Version: version},
		&commands.RaffleCommand{MasterId: masterId, DebugChannel: masterDebugChannel},
		&commands.SubmitCommand{ComPrefix: ComPrefix},
		&commands.EchoCommand{},
		&commands.PermitCommand{},
		&commands.PingCommand{},
		&commands.SpoilerCommand{},
		&commands.PollCommand{PollsHandler: commands.NewPollsHandler()},
		&commands.MentionCommand{},
		&commands.ServerCommand{ComPrefix: ComPrefix},
		&commands.ProfileCommand{MasterId: masterId},
		&commands.PinMoveCommand{ShouldLoadPins: Config["loadPins"] == "1"},
		commands.NewVeteranHandler(ComPrefix, masterDebugChannel, masterId),
	}

	setupCommands()
	setupHandlers(session)
	setupEvents(session)
}

func getCommands() []commands.Command {
	result := []commands.Command{}
	for _, o := range operations {
		if command, ok := o.(commands.Command); ok {
			result = append(result, command)
		}
	}
	return result
}

/*
Run through each operation and place each command into the command map (including any aliases)
*/
func setupCommands() {
	for _, command := range getCommands() {
		for _, key := range command.GetCommandKeys() {
			commandsMap[key] = command
		}
	}
}

/*
Run through each operation and run through any setup steps required by those operations
*/
func setupHandlers(session *discordgo.Session) {
	for _, o := range operations {
		if setup, ok := o.(commands.SetupHandler); ok {
			setup.Setup(session)
		}
	}
}

/*
Run through each operation and add a handler for any discord events those operations need
*/
func setupEvents(session *discordgo.Session) {
	for _, o := range operations {
		if handler, ok := o.(commands.EventHandler); ok {
			for _, h := range handler.EventHandlers() {
				session.AddHandler(h)
			}
		}
	}
}

/*
These handlers are global for all of moebot such as message creation and ready
*/
func addGlobalHandlers(discord *discordgo.Session) {
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(guildMemberAdd)
}

/*
Global handler for when new guild members join a discord guild. Typically used to welcome them if the server has enabled it.
*/
func guildMemberAdd(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	guild, err := session.Guild(member.GuildID)
	if err != nil {
		log.Println("Error fetching guild during guild member add", err)
		session.ChannelMessageSend(masterDebugChannel, fmt.Sprint("Error fetching guild during guild member add", err, member))
		return
	}
	server, err := db.ServerQueryOrInsert(guild.ID)
	if !server.Enabled {
		return
	}
	// only send out a welcome message is the server has one
	if server.WelcomeMessage.Valid {
		// then decide if we want to PM this welcome or post in a channel
		var channelId string
		if server.WelcomeChannel.Valid {
			channelId = server.WelcomeChannel.String
		} else {
			dmChannel, err := session.UserChannelCreate(member.User.ID)
			if err != nil {
				log.Println("ERROR! Unable to make DM channel with userID ", member.User.ID)
				return
			}
			channelId = dmChannel.ID
		}
		session.ChannelMessageSend(channelId, server.WelcomeMessage.String)
	}
	// then only assign a starter role if they have one set
	if server.StarterRole.Valid {
		var starterRole *discordgo.Role
		for _, guildRole := range guild.Roles {
			if guildRole.ID == server.StarterRole.String {
				starterRole = guildRole
				break
			}
		}
		if starterRole == nil {
			// couldn't find the starter role, try to let them know and then delete the starter role to prevent this error from appearing again
			if server.BotChannel.Valid {
				session.ChannelMessageSend(server.WelcomeChannel.String, "Hello, I couldn't find the starter role for this server! "+
					"Please notify a server admin (Like "+util.UserIdToMention(guild.OwnerID)+") Starter role will be removed.")
			} else if server.WelcomeChannel.Valid {
				session.ChannelMessageSend(server.WelcomeChannel.String, "Hello, I couldn't find the starter role for this server! "+
					"Please notify a server admin (Like "+util.UserIdToMention(guild.OwnerID)+") Starter role will be removed.")
			}
			log.Println("ERROR! Unable to find starter role for guild " + guild.Name + ". Deleting starter role.")
			server.StarterRole.Scan(nil)
			db.ServerFullUpdate(server)
		} else {
			session.GuildMemberRoleAdd(member.GuildID, member.User.ID, starterRole.ID)
		}
	}
}

/*
Global handler for when new messages are sent in any guild. The entry point for commands and other general handling
*/
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// bail out if we have any messages we want to ignore such as bot messages
	if message.Author.ID == session.State.User.ID || message.Author.Bot {
		return
	}

	//I'm guessing this block of code is the cause of our hokago-tea-time issues. Some of these are used just for their IDs while some are used for role checks.
	// Perhaps we could cache some of the information. Maybe timeout the role cache after 1 minute?
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

	server, err := db.ServerQueryOrInsert(guild.ID)
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Sorry, there was an error fetching this server. This is an issue with moebot not discord. "+
			"Please contact a moebot developer/admin.")
		return
	}

	member, err := session.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		log.Println("ERROR! Unable to get member in messageCreate ", err, message)
		return
	}

	// If the server is disabled, then don't allow any message processing
	// HOWEVER, if the user posting the message is this bot's owner or the guild's owner then let it through so they can enable the server
	isMaster := checker.IsMaster(message.Author.ID)
	isGuildOwner := permissions.IsGuildOwner(guild, message.Author.ID)
	if !server.Enabled && !isMaster && !isGuildOwner {
		return
	}

	var baseRole *discordgo.Role
	var starterRole *discordgo.Role
	for _, guildRole := range guild.Roles {
		if server.StarterRole.Valid && guildRole.ID == server.StarterRole.String {
			starterRole = guildRole
		}
		if server.BaseRole.Valid && guildRole.ID == server.BaseRole.String {
			baseRole = guildRole
		}
	}

	// Check if this user is a new user. This will determine what they can/can't do on the server.
	// Masters and guild owners are never a new user
	isNewUser := !isMaster && !isGuildOwner && server.RuleAgreement.Valid && starterRole != nil &&
		util.StrContains(member.Roles, starterRole.ID, util.CaseSensitive)

	if strings.HasPrefix(strings.ToUpper(message.Content), strings.ToUpper(ComPrefix)) {
		// todo: [rate-limit-spam] should add a check here for command spam

		if isNewUser {
			// if a starter role requested a command and the server has rule agreements, let them know they can't do that
			session.ChannelMessageSend(channel.ID, "Sorry "+message.Author.Mention()+", but you have to agree to the rules first to use bot commands! "+
				"Check the rules channel or ask an admin for more info.")
			// We don't need to process anything else since by typing a bot command they couldn't type a rule confirmation
			return
		}
		runCommand(session, message.Message, guild, channel, member)
	}
	// make sure to also check if they agreed to the rules
	if isNewUser {
		sanitizedMessage := util.MakeAlphaOnly(message.Content)
		if strings.HasPrefix(strings.ToUpper(sanitizedMessage), strings.ToUpper(server.RuleAgreement.String)) {
			if baseRole == nil {
				// Server only had a partial setup (rule agreement + starter role but no base role)
				session.ChannelMessageSend(channel.ID, "Hey... this is awkward... It seems like this server's admins setup a rule agreement but no base role. "+
					"Please notify a server admin (Like "+util.UserIdToMention(guild.OwnerID)+") Rule agreement will now be removed.")
				server.RuleAgreement.Scan(nil)
				err = db.ServerFullUpdate(server)
				if err != nil {
					log.Println("Error updateing server", err)
				}
				return
			}
			session.ChannelMessageSend(message.ChannelID, "Welcome "+message.Author.Mention()+"! We hope you enjoy your stay in our Discord server!")
			session.GuildMemberRoleAdd(guild.ID, member.User.ID, baseRole.ID)
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, starterRole.ID)
			log.Println("Updated user <" + member.User.Username + "> after reading the rules")
		}
	}
}

/*
Global handler that is called whenever moebot successfully connects to discord
*/
func ready(session *discordgo.Session, event *discordgo.Ready) {
	status := ComPrefix + " help"
	err := session.UpdateStatus(0, status)
	if err != nil {
		log.Println("Error setting moebot status", err)
	}
	log.Println("Set moebot's status to", status)
}

/*
Helper handler to check if the message provided is a command and if so, executes the command
*/
func runCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	commandKey := strings.ToUpper(messageParts[1])

	if command, commPresent := commandsMap[commandKey]; commPresent {
		params := messageParts[2:]
		if !checker.HasPermission(message.Author.ID, member.Roles, guild, command.GetPermLevel()) {
			session.ChannelMessageSend(channel.ID, "Sorry, you don't have a high enough permission level to access this command.")
			log.Println("!!PERMISSION VIOLATION!! Processing command: " + commandKey + " from user: {" + message.Author.String() + "}| With Params:{" +
				strings.Join(params, ",") + "}")
			return
		}
		log.Println("Processing command: " + commandKey + " from user: {" + message.Author.String() + "}| With Params:{" + strings.Join(params, ",") + "}")
		session.ChannelTyping(message.ChannelID)
		pack := commands.NewCommPackage(session, message, guild, member, channel, params)
		command.Execute(&pack)
	}
}
