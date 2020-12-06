package moeDiscord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

/*
Extensions for moebot that add more functionality or features onto discordGo.
Note that this is only for functionality that directly extend discordGo and not just any discord related functions, those can go in util
*/

/*
Gets a reaction from a message based on it's emoji name
*/
func GetReactionByName(message *discordgo.Message, reactionName string) *discordgo.MessageReactions {
	for _, r := range message.Reactions {
		if reactionName == r.Emoji.Name {
			return r
		}
	}
	return nil
}

func FindRoleByName(roles []*discordgo.Role, toFind string) *discordgo.Role {
	toFind = strings.ToUpper(toFind)
	for _, r := range roles {
		if strings.ToUpper(r.Name) == toFind {
			return r
		}
	}
	return nil
}

func FindRoleById(roles []*discordgo.Role, toFind string) *discordgo.Role {
	// for some reason roleIds have spaces in them...
	toFind = strings.TrimSpace(toFind)
	for _, r := range roles {
		if r.ID == toFind {
			return r
		}
	}
	return nil
}

func FindPermissionByRoleID(overwrites []*discordgo.PermissionOverwrite, toFind string) (*discordgo.PermissionOverwrite, bool) {
	for _, p := range overwrites {
		if p.Type == "role" && p.ID == toFind {
			return p, true
		}
	}
	return nil, false
}

func RetrieveBasePermissions(session *discordgo.Session, channel *discordgo.Channel, role *discordgo.Role, flags []int) map[int]bool {
	result := make(map[int]bool)
	permission, ok := FindPermissionByRoleID(channel.PermissionOverwrites, role.ID)
	if ok {
		mapPermissions(result, permission, flags)
	}
	if !ok || unsetFlags(permission, flags) { //no overwrite defined for the channel, looking in parent category
		parent, _ := session.Channel(channel.ParentID)
		permission, ok = FindPermissionByRoleID(parent.PermissionOverwrites, role.ID)
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

func GetEveryoneRoleForServer(session *discordgo.Session, serverID int) *discordgo.Role {
	server, err := db.ServerQueryById(serverID)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to retrieve server informations for Server ID: %v. ", serverID), err)
		return nil
	}
	roles, err := session.GuildRoles(server.GuildUID)
	if err != nil {
		log.Println("Failed to retrieve roles informations for Guild UID: "+server.GuildUID+". ", err)
		return nil
	}
	return FindRoleByName(roles, "@everyone")
}

func GetCurrentRolePermissionsForChannel(session *discordgo.Session, channelUID string, roleUID string) (*discordgo.PermissionOverwrite, error) {
	channel, err := session.Channel(channelUID)
	if err != nil {
		return nil, err
	}
	if p, ok := FindPermissionByRoleID(channel.PermissionOverwrites, roleUID); !ok {
		return &discordgo.PermissionOverwrite{
			ID:   roleUID,
			Type: "role",
		}, nil
	} else {
		return p, nil
	}
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
