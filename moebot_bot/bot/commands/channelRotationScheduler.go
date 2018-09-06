package commands

import (
	"log"

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

func (s *ChannelRotationScheduler) Execute(operationId int) {
	channelRotation, err := db.ChannelRotationQuery(operationId)
	if err != nil {
		//TODO
	}
	server, err := db.ServerQueryById(channelRotation.ServerID)
	if err != nil {
		//TODO
	}
	currentIndex := util.StringIndexOf(channelRotation.ChannelUIDList, channelRotation.CurrentChannelUID)
	roles, err := s.session.GuildRoles(server.GuildUid)
	if err != nil {
		//TODO
	}
	role := moeDiscord.FindRoleByName(roles, "@everyone")
	if role == nil {
		//TODO
	}
	err = s.hideChannel(role, channelRotation.CurrentChannelUID)
	if err != nil {
		//TODO
	}
	currentIndex++
	if currentIndex >= len(channelRotation.ChannelUIDList) {
		currentIndex = 0
	}
	err = s.showChannel(role, channelRotation.ChannelUIDList[currentIndex])
	if err != nil {
		//TODO
	}
	err = db.ChannelRotationUpdate(operationId, channelRotation.ChannelUIDList[currentIndex])
	if err != nil {
		//TODO
	}
	err = db.ScheduledOperationUpdateTime(operationId)
	if err != nil {
		//TODO
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

func (s *ChannelRotationScheduler) AddScheduledOperation(comm *CommPackage) {
	params := ParseCommand(comm.params, []string{"-channels", "-interval"})
	if params["-channels"] == "" {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, you need to specify a list of channels.")
		return
	}
	if params["-interval"] == "" {
		comm.session.ChannelMessageSend(comm.channel.ID, "Sorry, you need to specify a time interval.")
		return
	}
}
