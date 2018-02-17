package commands

import (
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type CustomCommand struct {
	ComPrefix string
}

func (cc *CustomCommand) Execute(pack *CommPackage) {
	// should have params: command name - role name or delete - id
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Please provide command name followed by the role name")
		return
	}
	if strings.ToUpper(pack.params[0]) == "DELETE" {
		id, err := strconv.Atoi(pack.params[1])
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a valid ID to delete")
			return
		}
		count := db.CustomRoleDelete(id)
		if count == -1 {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue deleting that custom role. Perhaps it doesn't exist?")
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "Deleted "+strconv.FormatInt(count, 10)+" custom role commands")
		}
	} else {
		// get the role and server
		server, err := db.ServerQueryOrInsert(pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Error storing server information")
			return
		}
		roleName := strings.Join(pack.params[1:], " ")
		r := util.FindRoleByName(pack.guild.Roles, roleName)
		if r == nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, it doesn't seem like that role exists on this server.")
			return
		}
		role, err := db.RoleQueryOrInsert(db.Role{
			ServerId: server.Id,
			RoleUid:  r.ID,
		})
		oldId, exists := db.CustomRoleRowExists(pack.params[0], server.GuildUid)
		if !exists {
			err = db.CustomRoleAdd(pack.params[0], server.Id, role.Id)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Error adding custom role")
				return
			}
			pack.session.ChannelMessageSend(pack.channel.ID, "Added custom command `"+pack.params[0]+"` tied to the role: "+roleName)
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "Custom command already exists. Delete with `"+cc.ComPrefix+" custom delete "+strconv.Itoa(oldId)+"`")
		}
	}
}

func (cc *CustomCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (cc *CustomCommand) GetCommandKeys() []string {
	return []string{"CUSTOM"}
}
