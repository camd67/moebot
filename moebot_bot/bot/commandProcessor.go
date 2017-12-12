package bot

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/camd67/moebot/moebot_bot/util"

	"github.com/bwmarrin/discordgo"
)

var commands = map[string]func(pack *commPackage){
	"TEAM": commTeam,
	// used for historical purposes since TEAM used to be ROLE
	"ROLE":      commRole,
	"RANK":      commRank,
	"NSFW":      commNsfw,
	"HELP":      commHelp,
	"CHANGELOG": commChange,
	"RAFFLE":    commRaffle,
	"SUBMIT":    commSubmit,
}

func RunCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commands[command]; commPresent {
		commFunc(&commPackage{session, message, guild, member, channel, messageParts[2:]})
	}
}

func commChange(pack *commPackage) {
	// should load this from db
	prefix := "\n`->` "
	changeMessage := prefix + "Included this command!" +
		prefix + "Updated `Rank` command to prevent removal of lowest role." +
		prefix + "Added random drops for tickets" +
		prefix + "Added `Raffle` related commands... For rafflin'" +
		prefix + "For future reference, previous versions included help, team, rank, and NSFW commands as well as a welcome message to the server."
	pack.session.ChannelMessageSend(pack.channel.ID, "`Moebot update log` (ver "+version+"): \n"+changeMessage)
}

func commHelp(pack *commPackage) {
	pack.session.ChannelMessageSend(pack.channel.ID, "Moebot has the following commands:\n"+
		"`"+ComPrefix+" team <role name>` - Changes your role to one of the approved roles. `"+ComPrefix+" team` to list all teams\n"+
		"`"+ComPrefix+" rank <rank name>` - Changes your rank to one of the approved ranks. `"+ComPrefix+" rank` to list all the ranks\n"+
		"`"+ComPrefix+" changelog` - Displays the changelog for moebot\n"+
		"`"+ComPrefix+" raffle` - Enters you into the raffle (if enabled on the server). Displays ticket count if already in the raffle\n"+
		"`"+ComPrefix+" submit <TYPE> <URL>` - Submits a link for relic/art competitions! Valid types are: art, relic. Valid URLS are from youtube, pastebin, and imgur.\n"+
		"`"+ComPrefix+" NSFW` - Gives you NSFW channel permissions\n"+
		"`"+ComPrefix+" help` - Displays this message")
}

func commSubmit(pack *commPackage) {
	// Salt + MIA
	if !(pack.guild.ID == "378336255030722570" || pack.guild.ID == "93799773856862208") {
		pack.session.ChannelMessageSend(pack.channel.ID, "Raffles are not enabled in this server! Speak to Salt to get your server added to the raffle!")
		return
	}
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a submission type and a URL in order to submit a link.")
		return
	}
	reg := regexp.MustCompile(".*(youtube.com|imgur.com|pastebin.com).*")
	if !reg.MatchString(pack.params[1]) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you must provide a link to an approved site! See submissions rules for more information")
		return
	}
	var raffleDataIndex int
	var raffleWord string
	if strings.ToUpper(pack.params[0]) == "ART" {
		raffleDataIndex = 0
		raffleWord = "art"
	} else if strings.ToUpper(pack.params[0]) == "RELIC" {
		raffleDataIndex = 1
		raffleWord = "relic"
	} else {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, I don't recognize that submission type. Valid types are: art, relic.")
		return
	}
	raffles, err := db.RaffleEntryQuery(pack.message.Author.ID, pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error trying to get your raffle information!")
		return
	}
	if len(raffles) != 1 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information! Make sure you're already in the raffle! (raffle command)")
		return
	}
	raffleData := strings.Split(raffles[0].RaffleData, db.RaffleDataSeparator)
	var ticketsToAdd = 2
	if raffleData[raffleDataIndex] != "NONE" {
		// if they've already got a submission, don't award bonus tickets
		ticketsToAdd = 0
		pack.session.ChannelMessageSend("378354584587862025", "Potential duplicate submission by "+pack.message.Author.Username+". Old url: `"+raffleData[raffleDataIndex]+"`")
	}
	if raffleDataIndex == 0 {
		raffles[0].SetRaffleData(pack.params[1] + db.RaffleDataSeparator + raffleData[1])
	} else if raffleDataIndex == 1 {
		raffles[0].SetRaffleData(raffleData[0] + db.RaffleDataSeparator + pack.params[1])
	}
	db.RaffleEntryUpdate(raffles[0], ticketsToAdd)
	pack.session.ChannelMessageSend(pack.channel.ID, "Submission accepted!")
	pack.session.ChannelMessagePin(pack.channel.ID, pack.message.ID)
	pack.session.ChannelMessageSend("388150028390105089", pack.message.Author.Mention()+" submitted "+raffleWord+": `"+pack.params[1]+"`")
}

func commRaffle(pack *commPackage) {
	// Salt + MIA
	if !(pack.guild.ID == "378336255030722570" || pack.guild.ID == "93799773856862208") {
		pack.session.ChannelMessageSend(pack.channel.ID, "Raffles are not enabled in this server! Speak to Salt to get your server added to the raffle!")
		return
	}
	const startTickets = 5
	raffleEntries, err := db.RaffleEntryQuery(pack.message.Author.ID, pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information!")
		return
	}
	// there should only be 1 of each raffle entry for every user + guild combo
	if len(raffleEntries) > 1 {
		log.Println("Queried for more than one raffle entry: userUid-", raffleEntries[0].UserUid)
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information!")
		return
	}
	if len(raffleEntries) == 0 {
		// haven't joined the raffle yet, make an entry + add tickets
		newRaffle := db.RaffleEntry{
			GuildUid:    pack.guild.ID,
			UserUid:     pack.message.Author.ID,
			RaffleType:  db.RaffleMIA,
			TicketCount: startTickets,
			RaffleData:  "NONE" + db.RaffleDataSeparator + "NONE",
		}
		err := db.RaffleEntryAdd(newRaffle)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue adding your raffle entry!")
			return
		}
		pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+", welcome to the raffle! You get "+
			strconv.Itoa(startTickets)+" tickets for joining!")
	} else {
		// already joined the raffle, let them know their ticket count and other information
		raffleData := strings.Split(raffleEntries[0].RaffleData, db.RaffleDataSeparator)
		pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+" you're already in the raffle! Your ticket count is: "+
			strconv.Itoa(raffleEntries[0].TicketCount)+". Your art submission is: `"+raffleData[0]+"`. Your relic submission is: `"+raffleData[1]+"`")
	}
}

func commRole(pack *commPackage) {
	pack.session.ChannelMessageSend(pack.channel.ID, "`"+ComPrefix+" role` has been renamed to `"+ComPrefix+" team`. Please use team instead!")
}

func commTeam(pack *commPackage) {
	processGuildRole([]string{"Nanachi", "Ozen", "Bondrewd", "Reg", "Riko", "Maruruk"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, false)
}

func commRank(pack *commPackage) {
	processGuildRole([]string{"Red", "Blue"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, true)
}

func commNsfw(pack *commPackage) {
	// force NSFW comm param so we can reuse guild role
	processGuildRole([]string{"NSFW"}, pack.session, []string{"NSFW"}, pack.channel, pack.guild, pack.message, false)
}

/*
Processes a guild role based on a list of allowed role names, and a requested role.
If StrictRole is true then the first role in the list of allowed roles is used if all roles are removed.
*/
func processGuildRole(allowedRoles []string, session *discordgo.Session, params []string, channel *discordgo.Channel, guild *discordgo.Guild, message *discordgo.Message, strictRole bool) {
	if len(params) == 0 || !util.StrContains(allowedRoles, params[0], util.CaseInsensitive) {
		session.ChannelMessageSend(channel.ID, message.Author.Mention()+" please provide one of the approved roles: "+strings.Join(allowedRoles, ", "))
		return
	}

	// get the list of roles and find the one that matches the text
	var roleToAdd *discordgo.Role
	var allRolesToChange []string
	requestedRole := strings.ToUpper(params[0])
	for _, role := range guild.Roles {
		for _, roleToCheck := range allowedRoles {
			if strings.HasPrefix(strings.ToUpper(role.Name), strings.ToUpper(roleToCheck)) {
				allRolesToChange = append(allRolesToChange, role.ID)
			}
		}
		if strings.HasPrefix(strings.ToUpper(role.Name), requestedRole) {
			roleToAdd = role
		}
	}
	if roleToAdd == nil {
		log.Println("Unable to find role", message)
		session.ChannelMessageSend(channel.ID, "Sorry "+message.Author.Mention()+" I don't recognize that role...")
		return
	}
	guildMember, err := session.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		log.Println("Unable to find guild member", err)
		return
	}
	memberRoles := guildMember.Roles
	if strictRole && util.StrContains(memberRoles, roleToAdd.ID, util.CaseSensitive) && strings.HasPrefix(strings.ToUpper(roleToAdd.Name), requestedRole) {
		// we're in strict mode, and got a removal for the first role. prevent that
		session.ChannelMessageSend(channel.ID, "You can't remove your lowest role for this command!")
		return
	}
	removeAllRoles(session, guildMember, allRolesToChange, guild)
	if util.StrContains(memberRoles, roleToAdd.ID, util.CaseSensitive) {
		session.GuildMemberRoleRemove(guild.ID, message.Author.ID, roleToAdd.ID)
		session.ChannelMessageSend(channel.ID, "Removed role: "+roleToAdd.Name+" for "+message.Author.Mention())
		log.Println("Removed role " + roleToAdd.Name + " to user: " + message.Author.Username)
	} else {
		session.GuildMemberRoleAdd(guild.ID, message.Author.ID, roleToAdd.ID)
		session.ChannelMessageSend(channel.ID, "Added role: "+roleToAdd.Name+" for "+message.Author.Mention())
		log.Println("Added role " + roleToAdd.Name + " to user: " + message.Author.Username)
	}
}

func removeAllRoles(session *discordgo.Session, member *discordgo.Member, rolesToRemove []string, guild *discordgo.Guild) {
	for _, roleToCheck := range rolesToRemove {
		if util.StrContains(member.Roles, roleToCheck, util.CaseSensitive) {
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, roleToCheck)
		}
	}
}
