package moeDiscord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
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
