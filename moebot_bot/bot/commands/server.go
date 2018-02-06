package commands

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type ServerCommand struct {
	checker PermissionChecker
}

func (sc ServerCommand) Execute(pack *CommPackage) {
	if m := sc.checker.HasModPerm(pack.Message.Author.ID, pack.Member.Roles); !m {
		pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
	const possibleConfigMessages = "Possible configs: {VeteranRank -> number}, {VeteranRole -> full role name}, {BotChannel -> channel ID}"
	s, err := db.ServerQueryOrInsert(pack.Guild.ID)

	if len(pack.Params) <= 1 {
		rank := int(util.GetInt64OrDefault(s.VeteranRank))
		role := util.GetStringOrDefault(s.VeteranRole)
		if role != "unknown" {
			role = util.FindRoleById(pack.Guild.Roles, role).Name
		}
		rule := util.GetStringOrDefault(s.RuleAgreement)
		welcome := util.GetStringOrDefault(s.WelcomeMessage)
		botChannel := util.GetStringOrDefault(s.BotChannel)
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "This server's configs: {Rank: "+strconv.Itoa(rank)+"} {Role: "+role+"} {Welcome: "+welcome+
			"} {Rule Confirm: "+rule+"} {BotChannel ID: "+botChannel+"}")
		return
	}
	configKey := strings.ToUpper(pack.Params[0])
	configValue := strings.Join(pack.Params[1:], " ")
	if err != nil {
		log.Println("Error trying to get server", err)
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an error getting the server")
		return
	}
	if configKey == "VETERANRANK" {
		rank, err := strconv.Atoi(configValue)
		if err != nil || rank < 0 {
			// don't bother logging this one. Someone's just given a non-number
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a positive number for the veteran rank")
			return
		}
		s.VeteranRank = sql.NullInt64{
			Int64: int64(rank),
			Valid: true,
		}
	} else if configKey == "VETERANROLE" {
		role := util.FindRoleByName(pack.Guild.Roles, configValue)
		if role == nil {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a valid role and make sure it's the full role name")
			return
		}
		s.VeteranRole = sql.NullString{
			String: role.ID,
			Valid:  true,
		}
	} else if configKey == "BOTCHANNEL" {
		c, err := pack.Session.Channel(configValue)
		if err != nil || c.Type != discordgo.ChannelTypeGuildText {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a valid text channel ID")
			return
		}
		s.BotChannel = sql.NullString{
			String: c.ID,
			Valid:  true,
		}
	} else {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, possibleConfigMessages)
		return
	}
	err = db.ServerFullUpdate(s)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an error updating the server table. Your change was probably not applied.")
		return
	}
	pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Updated this server!")
}

func (sc ServerCommand) Setup() {}
