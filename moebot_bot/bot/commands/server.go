package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

const serverPossibleCommands = "Possible configs: {WelcomeMessage -> string; max length " + db.ServerMaxWelcomeMessageLengthString + "} " +
	"{WelcomeChannel: ChannelId} {VeteranRank -> number} {VeteranRole -> full role name} {BotChannel -> channel ID} {RuleAgreement -> string; max length " +
	db.ServerMaxRuleAgreementLengthString + "} {StarterRole -> full role name} {BaseRole -> full role name} {Enabled -> true/false}"

type ServerCommand struct {
	ComPrefix string
}

func (sc *ServerCommand) Execute(pack *CommPackage) {
	s, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Error getting server information. This is an issue with moebot and not discord. Please let a moebot "+
			"dev or admin know!")
		return
	}

	if len(pack.params) < 1 {
		pack.session.ChannelMessageSend(pack.channel.ID, "This server's configuration is: "+db.ServerSprint(s))
		return
	}

	// where to start looking for params to this function
	configKeyIndex := 0
	// If they asked for a clear, pass that in when processing the config key
	shouldClear := false
	if strings.EqualFold(pack.params[0], "-CLEAR") {
		shouldClear = true
		configKeyIndex = 1
	}
	configKey := strings.ToUpper(pack.params[configKeyIndex])
	var configValue string
	if len(pack.params)-configKeyIndex == 1 {
		// they didn't  provide any arguments, so it's a help command instead
		configValue = ""
	} else {
		// We need to go one further than the start index so we can skip over the config key
		configValue = strings.Join(pack.params[configKeyIndex+1:], " ")
	}
	if sc.processServerConfigKey(configKey, configValue, pack, &s, shouldClear) {
		err = db.ServerFullUpdate(s)
		if err != nil {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an error updating the server table. Your change was probably not applied.")
			return
		}
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Updated this server!")
	}
}

func (sc *ServerCommand) processServerConfigKey(configKey string, configValue string, pack *CommPackage, s *db.Server, shouldClear bool) (success bool) {
	// todo: This function is starting to become a little cumbersome... Some has been refactored but more can be done
	// We only have a help command if we got an empty config value and it's not a clear command
	isHelp := configValue == "" && !shouldClear
	if configKey == "VETERANRANK" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "VeteranRank: "+strconv.Itoa(int(util.GetInt64OrDefault(s.VeteranRank))))
		} else if shouldClear {
			s.VeteranRank.Scan(nil)
		} else {
			rank, err := strconv.Atoi(configValue)
			if err != nil || rank < 0 {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a positive number for the veteran rank")
				return false
			}
			s.VeteranRank.Scan(int64(rank))
		}
	} else if configKey == "VETERANROLE" {
		if !sc.defaultServerRoleSet(pack, configValue, &s.VeteranRole, isHelp, "VeteranRole", shouldClear) {
			return
		}
	} else if configKey == "BOTCHANNEL" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "BotChannel: "+util.GetStringOrDefault(s.BotChannel))
		} else if shouldClear {
			s.BotChannel.Scan(nil)
		} else {
			c, err := pack.session.Channel(configValue)
			if err != nil || c.Type != discordgo.ChannelTypeGuildText || c.GuildID != pack.guild.ID {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid text channel ID")
				return false
			}
			s.BotChannel.Scan(c.ID)
		}
	} else if configKey == "WELCOMEMESSAGE" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "WelcomeMessage:"+util.GetStringOrDefault(s.WelcomeMessage))
		} else if shouldClear {
			s.WelcomeMessage.Scan(nil)
		} else {
			if len(configValue) > db.ServerMaxWelcomeMessageLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this property has a max length of: "+strconv.Itoa(db.ServerMaxWelcomeMessageLength))
				return false
			}
			if strings.HasPrefix(configValue, sc.ComPrefix) {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you can't use moebot's prefix in your welcome message.")
				return false
			}
			s.WelcomeMessage.Scan(configValue)
		}
	} else if configKey == "WELCOMECHANNEL" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "WelcomeChannel: "+util.GetStringOrDefault(s.BotChannel))
		} else if shouldClear {
			s.WelcomeChannel.Scan(nil)
		} else {
			c, err := pack.session.Channel(configValue)
			if err != nil || c.Type != discordgo.ChannelTypeGuildText || c.GuildID != pack.guild.ID {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid text channel ID")
				return false
			}
			s.WelcomeChannel.Scan(c.ID)
		}
	} else if configKey == "RULEAGREEMENT" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "RuleAgreement: "+util.GetStringOrDefault(s.RuleAgreement))
		} else if shouldClear {
			s.RuleAgreement.Scan(nil)
		} else {
			if len(configValue) > db.ServerMaxRuleAgreementLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this property has a max length of: "+strconv.Itoa(db.ServerMaxRuleAgreementLength))
				return false
			}
			if strings.HasPrefix(configValue, sc.ComPrefix) {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you can't use moebot's prefix in your rule agreement.")
				return false
			}
			s.RuleAgreement.Scan(configValue)
		}
	} else if configKey == "BASEROLE" {
		if !sc.defaultServerRoleSet(pack, configValue, &s.BaseRole, isHelp, "BaseRole", shouldClear) {
			return
		}
	} else if configKey == "STARTERROLE" {
		if !sc.defaultServerRoleSet(pack, configValue, &s.StarterRole, isHelp, "StarterRole", shouldClear) {
			return
		}
	} else if configKey == "ENABLED" {
		if isHelp {
			pack.session.ChannelMessageSend(pack.channel.ID, "Enabled: "+strconv.FormatBool(s.Enabled))
		} else if shouldClear {
			s.Enabled = false
		} else {
			newBool, err := strconv.ParseBool(configValue)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, I don't recognize that as a boolean. Please provide either true/false.")
				return
			}
			s.Enabled = newBool
		}
	} else {
		pack.session.ChannelMessageSend(pack.message.ChannelID, serverPossibleCommands)
		return false
	}
	// if we are in a help state, then we never succeeded, otherwise we always did if we got to this point
	return !isHelp
}

func (sc *ServerCommand) defaultServerRoleSet(pack *CommPackage, configValue string, toSet *sql.NullString, isHelp bool, name string,
	shouldClear bool) (shouldReturn bool) {

	if isHelp {
		pack.session.ChannelMessageSend(pack.channel.ID, name+": "+util.GetStringOrDefault(*toSet))
		return false
	} else if shouldClear {
		toSet.Scan(nil)
	} else {
		role := util.FindRoleByName(pack.guild.Roles, configValue)
		if role == nil {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid role and make sure it's the full role name")
			return false
		}
		toSet.Scan(role.ID)
	}
	return true
}

func (sc *ServerCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (sc *ServerCommand) GetCommandKeys() []string {
	return []string{"SERVER"}
}

func (sc *ServerCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s server <config setting> <value>` - Master/Mod Changes a config setting on the server to a given value. `%[1]s server` to list configs.", commPrefix)
}
