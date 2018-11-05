package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type ChannelRotationScheduler struct {
	schedulerType types.SchedulerType
	session       *discordgo.Session
}

func NewChannelRotationScheduler(schedulerType types.SchedulerType, session *discordgo.Session) *ChannelRotationScheduler {
	return &ChannelRotationScheduler{schedulerType, session}
}

func (s *ChannelRotationScheduler) Execute(operationId int64) {
	channelRotation, err := db.ChannelRotationQuery(operationId)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to retrieve operation informations for Operation ID: %v (operation is possibly being created). ", operationId), err)
		return
	}
	role := moeDiscord.GetEveryoneRoleForServer(s.session, channelRotation.ServerID)
	if role == nil {
		log.Println(fmt.Sprintf("Failed to retrieve everyone role informations for Server ID: %v. ", channelRotation.ServerID), err)
		return
	}
	if !s.rotateChannels(role, channelRotation.CurrentChannelUID, channelRotation.NextChannelUID()) {
		return
	}

	s.updateChannelRotationOperation(channelRotation)
}

func (s *ChannelRotationScheduler) rotateChannels(role *discordgo.Role, channelToHideUID string, channelToShowUID string) bool {
	if channelToHideUID != "" {
		err := s.hideChannel(role, channelToHideUID)
		if err != nil {
			log.Println("Failed to change channel permissions (hide) for Channel UID: "+channelToHideUID+". ", err)
			return false
		}
	}
	if channelToShowUID != "" {
		err := s.showChannel(role, channelToShowUID)
		if err != nil {
			log.Println("Failed to change channel permissions (show) for Channel UID: "+channelToShowUID+". ", err)
			return false
		}
	}
	return true
}

func (s *ChannelRotationScheduler) hideChannel(role *discordgo.Role, channelUID string) error {
	permissions, err := moeDiscord.GetCurrentRolePermissionsForChannel(s.session, channelUID, role.ID)
	if err != nil {
		log.Println("Error while hiding channel, failed to get permissions for Role UID: " + role.ID + ".")
		return err
	}
	if permissions.Deny&discordgo.PermissionReadMessages != 0 {
		return nil //no need to do anything, channel is already hidden
	}
	permissions.Allow = permissions.Allow &^ discordgo.PermissionReadMessages
	permissions.Deny = permissions.Deny | discordgo.PermissionReadMessages
	err = s.session.ChannelPermissionSet(channelUID, role.ID, "role", permissions.Allow, permissions.Deny)
	if err != nil {
		log.Println("Error while setting channel permissions to hidden:", err)
	}
	return err
}

func (s *ChannelRotationScheduler) showChannel(role *discordgo.Role, channelUID string) error {
	permissions, err := moeDiscord.GetCurrentRolePermissionsForChannel(s.session, channelUID, role.ID)
	if err != nil {
		log.Println("Error while showing channel, failed to get permissions for Role UID: " + role.ID + ".")
		return err
	}
	if permissions.Allow&discordgo.PermissionReadMessages != 0 {
		return nil //no need to do anything, channel is already visible
	}
	permissions.Allow = permissions.Allow | discordgo.PermissionReadMessages
	permissions.Deny = permissions.Deny &^ discordgo.PermissionReadMessages
	err = s.session.ChannelPermissionSet(channelUID, role.ID, "role", permissions.Allow, permissions.Deny)
	if err != nil {
		log.Println("Error while setting channel permissions to visible:", err)
	}
	return err
}

func (s *ChannelRotationScheduler) updateChannelRotationOperation(channelRotation *types.ChannelRotation) bool {
	err := db.ChannelRotationUpdate(channelRotation.ID, channelRotation.NextChannelUID())
	if err != nil {
		log.Println(fmt.Sprintf("Failed to update current channel in Operation ID: %v. ", channelRotation.ID), err)
		return false
	}
	_, err = db.ScheduledOperationUpdateTime(channelRotation.ID)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to update operation time in Operation ID: %v. ", channelRotation.ID), err)
		return false
	}
	return true
}

func (s *ChannelRotationScheduler) Keyword() string {
	return "ChannelRotation"
}

func (s *ChannelRotationScheduler) Help() string {
	return "`" + s.Keyword() + " -channels <channels> -interval <interval>`: Rotates through the specified channels, making them visible one at a time. \n " +
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

	intervalString, err := util.ParseIntervalToISO(params["-interval"])
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
	if len(strings.Join(channels, " ")) > 1000 {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, you specified too many channels. Please specify less channels for the rotation.")
		return fmt.Errorf("Too many channels passed as an argument")
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
	if len(channelRotation.ChannelUIDList) == 1 {
		return "Rotating channel " + channelRotation.ChannelUIDList[0]
	}
	var b strings.Builder
	b.WriteString("Rotating channels")
	for _, c := range channelRotation.ChannelUIDList {
		b.WriteString(" <#" + c + ">")
	}
	return b.String()
}
