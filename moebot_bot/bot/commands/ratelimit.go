package commands

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"

	"github.com/bwmarrin/discordgo"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RatelimitCommand struct {
	channels map[string]*activityChecker
	mutex    *sync.Mutex
}

func NewRatelimitCommand() *RatelimitCommand {
	return &RatelimitCommand{
		channels: make(map[string]*activityChecker),
		mutex:    &sync.Mutex{},
	}
}

func (c *RatelimitCommand) Execute(pack *CommPackage) {
	var err error
	commands := ParseCommand(pack.params, []string{"-channel", "-activity", "-remove"})
	chID := strings.Trim(commands["-channel"], " ")
	chID, valid := util.ExtractChannelIdFromString(chID)
	if !valid {
		pack.session.ChannelMessageSend(pack.channel.ID, "Invalid channel specified.")
		return
	}
	channel, err := pack.session.Channel(chID)
	if err != nil || channel.GuildID != pack.channel.GuildID {
		pack.session.ChannelMessageSend(pack.channel.ID, "Invalid channel specified.")
		return
	}
	dbServer, err := db.ServerQueryOrInsert(channel.GuildID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Error while retrieving the server, please try again.")
		return
	}
	dbChannel, err := db.ChannelQueryOrInsert(channel.ID, &dbServer)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Error while retrieving the channel, please try again.")
		return
	}
	if _, ok := commands["-remove"]; ok {
		if c.removeChecker(pack.session, chID) {
			dbChannel.MessageLimit = 0
			db.ChannelUpdate(dbChannel)
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "The specified channel doesn't have a valid check.")
		}
		return
	}
	var activity int
	if activity, err = strconv.Atoi(commands["-activity"]); err != nil || activity <= 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Invalid activity threshold specified.")
		return
	}
	roles, err := pack.session.GuildRoles(pack.channel.GuildID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "There was a problem fetching the '@everyone' role, please try again.")
		return
	}
	role := moeDiscord.FindRoleByName(roles, "@everyone")
	if role == nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "There was a problem fetching the '@everyone' role, please try again.")
		return
	}
	permissions := getBasePermissions(pack.session, channel, role, []int{discordgo.PermissionAttachFiles, discordgo.PermissionEmbedLinks})
	if permissions == nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "There was a problem fetching '@everyone''s permissions, please try again.")
		return
	}
	if !permissions[discordgo.PermissionAttachFiles] && !permissions[discordgo.PermissionEmbedLinks] {
		pack.session.ChannelMessageSend(pack.channel.ID, "Users are already forbidden from sending images and embedding links in this channel.")
		return
	}
	c.addChecker(pack.session, channel.ID, role.ID, permissions, activity)
	dbChannel.MessageLimit = activity
	err = db.ChannelUpdate(dbChannel)
	if err != nil {
		c.removeChecker(pack.session, channel.ID)
		pack.session.ChannelMessageSend(pack.channel.ID, "Cannot add checker to the specified channel, please try again.")
		return
	}
	pack.session.ChannelMessageSend(pack.channel.ID, "Moebot is now checking the specified channel.")
}

func (c *RatelimitCommand) removeChecker(session *discordgo.Session, channelUID string) bool {
	c.mutex.Lock()
	r, ok := c.channels[channelUID]
	c.mutex.Unlock()
	if ok {
		r.removeChecker(session)
		delete(c.channels, channelUID)
		return true
	} else {
		return false
	}
}

func (c *RatelimitCommand) addChecker(session *discordgo.Session, channelUID string, roleUID string, basePermissions map[int]bool, activityLimit int) {
	c.mutex.Lock()
	if existing, ok := c.channels[channelUID]; ok {
		existing.removeChecker(session)
	}
	c.channels[channelUID] = newActivityChecker(session, channelUID, roleUID, basePermissions, activityLimit)
	c.mutex.Unlock()
}

func getBasePermissions(session *discordgo.Session, channel *discordgo.Channel, role *discordgo.Role, flags []int) map[int]bool {
	result := make(map[int]bool)
	permission, ok := moeDiscord.FindPermissionByRoleID(channel.PermissionOverwrites, role.ID)
	if ok {
		mapPermissions(result, permission, flags)
	}
	if !ok || unsetFlags(permission, flags) { //no overwrite defined for the channel, looking in parent category
		parent, _ := session.Channel(channel.ParentID)
		permission, ok = moeDiscord.FindPermissionByRoleID(parent.PermissionOverwrites, role.ID)
		if ok {
			mapPermissions(result, permission, flags)
		}
		if !ok || unsetFlags(permission, flags) { //no overwrite defined for the channel, using role permissions
			permission = &discordgo.PermissionOverwrite{
				ID:   role.ID,
				Type: "role",
			}
			for _, f := range flags {
				if role.Permissions&f != 0 {
					permission.Allow = permission.Allow | f
				} else {
					permission.Deny = permission.Deny | f
				}
			}
			mapPermissions(result, permission, flags)
		}
	}
	return result
}

func unsetFlags(permission *discordgo.PermissionOverwrite, flags []int) bool {
	for _, f := range flags {
		if permission.Allow&f == 0 && permission.Deny&f == 0 {
			return true
		}
	}
	return false
}

func mapPermissions(base map[int]bool, permission *discordgo.PermissionOverwrite, flags []int) {
	for _, f := range flags {
		if _, ok := base[f]; !ok { //only do this if the flag is unset, to allow hierarchy assignations
			if permission.Allow&f != 0 || permission.Deny&f != 0 {
				base[f] = permission.Allow&f != 0
			}
		}
	}
}

func (c *RatelimitCommand) Setup(session *discordgo.Session) {
	channels, err := db.ChannelQueryWithLimit()
	if err != nil {
		return
	}
	for _, channel := range channels {
		go c.setupChannel(session, channel)
	}
}

func (c *RatelimitCommand) setupChannel(session *discordgo.Session, dbChannel db.Channel) {
	channel, err := session.Channel(dbChannel.ChannelUid)
	if err != nil {
		return
	}
	roles, err := session.GuildRoles(channel.GuildID)
	if err != nil {
		return
	}
	role := moeDiscord.FindRoleByName(roles, "@everyone")
	if role == nil {
		return
	}
	permissions := getBasePermissions(session, channel, role, []int{discordgo.PermissionAttachFiles, discordgo.PermissionEmbedLinks})
	if permissions == nil {
		return
	}
	if !permissions[discordgo.PermissionAttachFiles] && !permissions[discordgo.PermissionEmbedLinks] {
		return
	}
	c.addChecker(session, channel.ID, role.ID, permissions, dbChannel.MessageLimit)
}

func (c *RatelimitCommand) GetPermLevel() db.Permission {
	return db.PermMod
}
func (c *RatelimitCommand) GetCommandKeys() []string {
	return []string{"RATELIMIT"}
}
func (c *RatelimitCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s ratelimit -channel <channel> -activity <minMessages>` - Disallow users to post embeds and attachments once the indicated activity per minute is reached.  Use `%[1]s ratelimit -channel <channel> -remove` to remove the rate limit", commPrefix)
}

func (c *RatelimitCommand) Close(session *discordgo.Session) {
	c.mutex.Lock()
	for _, chk := range c.channels {
		chk.removeChecker(session)
	}
	c.mutex.Unlock()
}
