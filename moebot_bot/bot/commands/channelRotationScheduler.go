package commands

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type ChannelRotationScheduler struct {
	schedulerType db.SchedulerType
	session       *discordgo.Session
}

func NewChannelRotationScheduler(schedulerType db.SchedulerType, session *discordgo.Session) *ChannelRotationScheduler {
	return &ChannelRotationScheduler{schedulerType, session}
}

func (s *ChannelRotationScheduler) Execute(operationId int64) {
	channelRotation, err := db.ChannelRotationQuery(operationId)
	if err != nil {
		log.Println("Failed to retrieve operation informations (operation is possibly being created) ", err)
		return
	}
	server, err := db.ServerQueryById(channelRotation.ServerID)
	if err != nil {
		log.Println("Failed to retrieve server informations ", err)
		return
	}
	currentIndex := util.StringIndexOf(channelRotation.ChannelUIDList, channelRotation.CurrentChannelUID)
	roles, err := s.session.GuildRoles(server.GuildUid)
	if err != nil {
		log.Println("Failed to retrieve roles informations ", err)
		return
	}
	role := moeDiscord.FindRoleByName(roles, "@everyone")
	if role == nil {
		log.Println("Failed to retrieve everyone role informations ", err)
		return
	}
	err = s.hideChannel(role, channelRotation.CurrentChannelUID)
	if err != nil {
		log.Println("Failed to change channel permissions (hide) ", err)
		return
	}
	currentIndex++
	if currentIndex >= len(channelRotation.ChannelUIDList) {
		currentIndex = 0
	}
	err = s.showChannel(role, channelRotation.ChannelUIDList[currentIndex])
	if err != nil {
		log.Println("Failed to change channel permissions (show) ", err)
		return
	}
	err = db.ChannelRotationUpdate(operationId, channelRotation.ChannelUIDList[currentIndex])
	if err != nil {
		log.Println("Failed to update current channel in operation ", err)
		return
	}
	err = db.ScheduledOperationUpdateTime(operationId)
	if err != nil {
		log.Println("Failed to update operation time ", err)
		return
	}
}

func (s *ChannelRotationScheduler) hideChannel(role *discordgo.Role, channelUID string) error {
	permissions, err := getCurrentPermissions(s.session, channelUID, role.ID)
	//check if err != nil
	//check if channel is already not visible
	permissions.Allow = permissions.Allow &^ discordgo.PermissionReadMessages
	permissions.Deny = permissions.Deny | discordgo.PermissionReadMessages
	err = s.session.ChannelPermissionSet(channelUID, role.ID, "role", permissions.Allow, permissions.Deny)
	if err != nil {
		log.Println("Error while setting channel permissions:", err)
	}
	return err
}

func (s *ChannelRotationScheduler) showChannel(role *discordgo.Role, channelUID string) error {
	permissions, err := getCurrentPermissions(s.session, channelUID, role.ID)
	//check if err != nil
	//check if channel is already not visible
	permissions.Allow = permissions.Allow | discordgo.PermissionReadMessages
	permissions.Deny = permissions.Deny &^ discordgo.PermissionReadMessages
	err = s.session.ChannelPermissionSet(channelUID, role.ID, "role", permissions.Allow, permissions.Deny)
	if err != nil {
		log.Println("Error while setting channel permissions:", err)
	}
	return err
}

func getCurrentPermissions(session *discordgo.Session, channelUID string, roleUID string) (*discordgo.PermissionOverwrite, error) {
	channel, err := session.Channel(channelUID)
	if err != nil {
		return nil, err
	}
	if p, ok := moeDiscord.FindPermissionByRoleID(channel.PermissionOverwrites, roleUID); !ok {
		return &discordgo.PermissionOverwrite{
			ID:   roleUID,
			Type: "role",
		}, nil
	} else {
		return p, nil
	}
}

func (s *ChannelRotationScheduler) Keyword() string {
	return "ChannelRotation"
}

func (s *ChannelRotationScheduler) Help() string {
	return "`" + s.Keyword() + " -channels <channels> -interval <interval>`: Rotates through the specified channels, making them visible one at a time. \r\n " +
		"The command doesn't edit already existing permissions, so when using this command make sure that only the first channel in the list is currently visible."
}

func (s *ChannelRotationScheduler) AddScheduledOperation(comm *CommPackage) error {
	params := ParseCommand(comm.params, []string{"-channels", "-interval"})
	if params["-channels"] == "" {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, you need to specify a list of channels.")
		return fmt.Errorf("-channels parameter empty")
	}
	if params["-interval"] == "" {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, you need to specify a time interval.")
		return fmt.Errorf("-interval parameter empty")
	}

	intervalString, err := parseCommandInterval(params["-interval"])
	if err != nil {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, the interval you specified is invalid. You need to specify the interval in the format `XWXDXh`, for example `5W6D4h` for 5 weeks, 6 days and 4 hours.")
		return err
	}

	server, err := db.ServerQueryOrInsert(comm.guild.ID)
	if err != nil {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, there was a problem retrieving the current server informations. Please try again.")
		return err
	}
	channels := []string{}
	for _, c := range strings.Split(params["-channels"], " ") {
		channels = append(channels, strings.Trim(c, "<#>"))
	}
	for _, chUID := range channels {
		ch, err := comm.session.Channel(chUID)
		if err != nil {
			comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, there was a problem retrieving channel informations. Please try again.")
			return err
		}
		if ch.GuildID != comm.guild.ID {
			comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, there was a problem retrieving channel informations. Please try again.")
			return err
		}
	}

	err = db.ChannelRotationAdd(server.Id, channels[0], channels, intervalString)
	if err != nil {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, there was a problem adding the rotation to the server.")
		return err
	}
	comm.session.ChannelMessageSend(comm.channel.ID, "Channel rotation successfully added")
	return nil
}

func (s *ChannelRotationScheduler) OperationDescription(operationID int64) string {
	channelRotation, err := db.ChannelRotationQuery(operationID)
	if err != nil {
		return "Failed to retrieve channel list"
	}
	result := "Rotating channels"
	for _, c := range channelRotation.ChannelUIDList {
		result += " <#" + c + ">"
	}
	return result
}

func parseCommandInterval(interval string) (string, error) {
	intervalsOrder := []string{"Y", "M", "W", "D", "h", "m"}
	rx, _ := regexp.Compile("^(\\d+[YMWDhm]){1}(\\d+[YMWDhm]){0,1}(\\d+[YMWDhm]){0,1}(\\d+[YMWDhm]){0,1}(\\d+[YMWDhm]){0,1}$")
	if !rx.MatchString(interval) {
		return "", fmt.Errorf("Invalid interval string")
	}

	matches := rx.FindAllStringSubmatch(interval, -1)[0][1:]
	intervalString := "P"
	for _, indicator := range intervalsOrder {
		for _, match := range matches {
			if strings.Contains(match, indicator) {
				intervalString += strings.ToUpper(match)
			}
		}
		if indicator == "D" { //Adds time separator after day segment
			if intervalString != "P" {
				intervalString += "T"
			} else {
				intervalString += "0DT"
			}
		}
	}
	intervalString = strings.Trim(intervalString, "T")
	return intervalString, nil
}
