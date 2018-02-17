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
}

func (sc *ServerCommand) Execute(pack *CommPackage) {
	const possibleConfigMessages = "Possible configs: {VeteranRank -> number}, {VeteranRole -> full role name}, {BotChannel -> channel ID}"
	s, err := db.ServerQueryOrInsert(pack.guild.ID)

	if len(pack.params) <= 1 {
		rank := int(util.GetInt64OrDefault(s.VeteranRank))
		role := util.GetStringOrDefault(s.VeteranRole)
		if role != "unknown" {
			role = util.FindRoleById(pack.guild.Roles, role).Name
		}
		rule := util.GetStringOrDefault(s.RuleAgreement)
		welcome := util.GetStringOrDefault(s.WelcomeMessage)
		botChannel := util.GetStringOrDefault(s.BotChannel)
		pack.session.ChannelMessageSend(pack.message.ChannelID, "This server's configs: {Rank: "+strconv.Itoa(rank)+"} {Role: "+role+"} {Welcome: "+welcome+
			"} {Rule Confirm: "+rule+"} {BotChannel ID: "+botChannel+"}")
		return
	}
	configKey := strings.ToUpper(pack.params[0])
	configValue := strings.Join(pack.params[1:], " ")
	if err != nil {
		log.Println("Error trying to get server", err)
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an error getting the server")
		return
	}
	if configKey == "VETERANRANK" {
		rank, err := strconv.Atoi(configValue)
		if err != nil || rank < 0 {
			// don't bother logging this one. Someone's just given a non-number
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a positive number for the veteran rank")
			return
		}
		s.VeteranRank = sql.NullInt64{
			Int64: int64(rank),
			Valid: true,
		}
	} else if configKey == "VETERANROLE" {
		role := util.FindRoleByName(pack.guild.Roles, configValue)
		if role == nil {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid role and make sure it's the full role name")
			return
		}
		s.VeteranRole = sql.NullString{
			String: role.ID,
			Valid:  true,
		}
	} else if configKey == "BOTCHANNEL" {
		c, err := pack.session.Channel(configValue)
		if err != nil || c.Type != discordgo.ChannelTypeGuildText {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid text channel ID")
			return
		}
		s.BotChannel = sql.NullString{
			String: c.ID,
			Valid:  true,
		}
	} else {
		pack.session.ChannelMessageSend(pack.message.ChannelID, possibleConfigMessages)
		return
	}
	err = db.ServerFullUpdate(s)
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an error updating the server table. Your change was probably not applied.")
		return
	}
	pack.session.ChannelMessageSend(pack.message.ChannelID, "Updated this server!")
}

func (sc *ServerCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (sc *ServerCommand) GetCommandKeys() []string {
	return []string{"SERVER"}
}
