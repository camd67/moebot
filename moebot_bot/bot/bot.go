package bot

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

const (
	version = "0.2"
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
	killList             = make(map[string]bool)
	userTicketCooldown   = make(map[string]int64)
)

func SetupMoebot(session *discordgo.Session) {
	addHandlers(session)
	db.SetupDatabase(Config["dbPass"], Config["moeDataPass"])
}

func addHandlers(discord *discordgo.Session) {
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(guildMemberAdd)
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
	if message.Author.ID == session.State.User.ID {
		return
	}
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		// missing channel
		return
	}

	messageTime, _ := message.Timestamp.Parse()

	guild, err := session.Guild(channel.GuildID)

	member, err := session.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		log.Println("ERROR! Unable to get member in messageCreate", err, message)
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
	// temp ignore IHG + Salt
	if err != nil || guild.ID == "84724034129907712" {
		// missing guild
		return
	}

	if strings.HasPrefix(message.Content, ComPrefix) {
		// should add a check here for command spam
		if util.StrContains(member.Roles, oldStarterRole.ID, util.CaseSensitive) {
			// bail out to prevent any new users from using bot commands
			return
		}
		session.ChannelTyping(message.ChannelID)
		RunCommand(session, message.Message, guild, channel, member)
	} else if util.StrContains(allowedNonComServers, guild.ID, util.CaseSensitive) {
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
			session.ChannelMessageSend(message.ChannelID, "We hope you enjoy your stay in our Discord server! Make sure to head over #announcements and #role-request to see whats new and get your roles!")
			session.GuildMemberRoleAdd(guild.ID, member.User.ID, starterRole.ID)
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, oldStarterRole.ID)
			log.Println("Updated user <" + member.User.Username + "> after reading the rules")
		}
	}
	// distribute tickets
	distributeTickets(guild, message, session, messageTime)
}

func distributeTickets(guild *discordgo.Guild, message *discordgo.MessageCreate, session *discordgo.Session, messageTime time.Time) {
	const ticketCooldown = int64(time.Hour * 24)
	if guild.ID == "378336255030722570" || guild.ID == "93799773856862208" {
		if oldTime, hasTime := userTicketCooldown[message.Author.ID]; hasTime {
			// they've got an old time, check for how long it's been
			if oldTime+ticketCooldown > messageTime.UnixNano() {
				// they haven't waited enough, bail
				return
			}
		}
		const maxChance = 100
		const ticketChance = 5
		if rand.Int()%maxChance <= ticketChance {
			// they've won a ticket, let them know and update timestamp + db
			userTicketCooldown[message.Author.ID] = messageTime.UnixNano()
			raffles, err := db.RaffleEntryQuery(message.Author.ID, guild.ID)
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Sorry, there was an issue finding your raffle information")
				return
			}

			// just ignore anything that isn't in the database
			if len(raffles) == 1 {
				db.RaffleEntryUpdate(raffles[0], 1)
				currTickets := raffles[0].TicketCount + 1
				session.ChannelMessageSend("378680855339728918", message.Author.Mention()+", congrats! You just earned another ticket! Your current tickets are: "+strconv.Itoa(currTickets))
			}
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
