package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
	"github.com/camd67/moebot/moebot_bot/util/rolerules"
)

type RoleCommand struct {
	ComPrefix   string
	PermChecker permissions.PermissionChecker
}

func (rc *RoleCommand) Execute(pack *CommPackage) {
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error loading server information!")
		return
	}
	var vetRole *discordgo.Role
	if server.VeteranRole.Valid {
		vetRole = moeDiscord.FindRoleById(pack.guild.Roles, server.VeteranRole.String)
	}
	if len(pack.params) == 0 {
		printAllRoles(server, vetRole, pack)
	} else {
		var role *discordgo.Role
		var dbRole *models.Role

		var roleNameBuf strings.Builder
		for _, param := range pack.params {
			if !strings.HasPrefix(param, "-") {
				roleNameBuf.WriteString(param)
				roleNameBuf.WriteString(" ")
			}
		}
		roleNameString := strings.TrimSpace(roleNameBuf.String())
		dbRole, err = db.RoleQueryTrigger(roleNameString, server.ID)
		// an invalid trigger should pretty much never happen, but checking for it anyways
		// however an error may indicate that there were simply no roles in the result set
		if err != nil || !dbRole.Trigger.Valid {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the role, or the role you provided doesn't exist. "+
				"Please provide a valid role. `"+rc.ComPrefix+" role` to list all roles for this server.")
			return
		}
		role = moeDiscord.FindRoleById(pack.guild.Roles, dbRole.RoleUID)
		if role == nil {
			log.Println("Nil dbRole when searching for dbRole id:" + dbRole.RoleUID)
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue finding that role in this server. It may have been deleted.")
			return
		}
		rules, err := rolerules.GetRulesForRole(server, dbRole, rc.ComPrefix)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem fetching the apply rules for the given role. Please try again.")
			return
		}
		usrRank, _ := db.UserServerRankQuery(pack.message.Author.ID, pack.guild.ID)
		action := &rolerules.RoleAction{
			Role:            dbRole,
			UserRank:        usrRank,
			Member:          pack.member,
			Guild:           pack.guild,
			Channel:         pack.channel,
			OriginalMessage: pack.message,
		}
		if util.StrContains(pack.member.Roles, role.ID, util.CaseSensitive) {
			action.Action = rolerules.RoleRemove
		} else {
			action.Action = rolerules.RoleAdd
		}
		success, message := checkRules(rules, action, pack)
		if message != "" {
			pack.session.ChannelMessageSend(pack.channel.ID, message)
		}
		if !success {
			return
		}
		success, message = applyRules(rules, action, pack)
		if !success {
			return
		}
		var builder strings.Builder
		if action.Action == rolerules.RoleRemove {
			pack.session.GuildMemberRoleRemove(pack.guild.ID, pack.message.Author.ID, role.ID)
			builder.WriteString("Removed role `" + role.Name + "` for " + pack.message.Author.Mention())
		} else {
			pack.session.GuildMemberRoleAdd(pack.guild.ID, pack.message.Author.ID, role.ID)
			builder.WriteString("Added role `" + role.Name + "` for " + pack.message.Author.Mention())
		}
		builder.WriteString(message) //messages from the apply functions are sent after the role change confirmation
		pack.session.ChannelMessageSend(pack.channel.ID, builder.String())
	}
}

func checkRules(rules []rolerules.RoleRule, action *rolerules.RoleAction, pack *CommPackage) (bool, string) {
	var builder strings.Builder
	for _, rule := range rules {
		check, message := rule.Check(pack.session, action)
		builder.WriteString(message)
		if !check {
			return false, builder.String()
		}
	}
	return true, builder.String()
}

func applyRules(rules []rolerules.RoleRule, action *rolerules.RoleAction, pack *CommPackage) (bool, string) {
	var builder strings.Builder
	for _, rule := range rules {
		check, message := rule.Apply(pack.session, action)
		builder.WriteString(message)
		if !check {
			return false, builder.String()
		}
	}
	return true, builder.String()
}

func (rc *RoleCommand) GetPermLevel() types.Permission {
	return types.PermAll
}

func (rc *RoleCommand) GetCommandKeys() []string {
	return []string{"ROLE"}
}

func (rc *RoleCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s role <role name>` - Changes your role to one of the approved roles. `%[1]s role` to list all the roles", commPrefix)
}
func printAllRoles(server *models.Server, vetRole *discordgo.Role, pack *CommPackage) {
	triggersByGroup := make(map[string][]string)
	// go find all the roles for this server
	roles, err := db.RoleQueryServer(server)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot!")
		return
	}
	// Then find all the groups for the server
	roleGroups, err := db.RoleGroupQueryServer(server)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the roles for this server. This is an issue with moebot!")
		return
	}
	if vetRole != nil {
		// triggers = append(triggers, vetRole.Name)
		// TODO: this should be the name of the role, but role is restricted to one word right now...
		triggersByGroup["Veteran"] = append(triggersByGroup["Veteran"], "veteran")
	}
	for _, role := range roles {
		if !role.Trigger.Valid {
			// skip any invalid triggers. We don't want people thinking that they can choose roles they actually can't
			continue
		}
		if vetRole != nil && role.RoleUID == vetRole.ID { //ignore veteran role since we already added it manually
			continue
		}
		// Could maybe make a map here, but the group size is going to be pretty small
		foundGroup := false
		for _, group := range roleGroups {
			if group.Type != types.GroupTypeNoMultiples {
				for _, gr := range role.R.RoleGroups {
					if gr.ID == group.Id {
						triggersByGroup[group.Name] = append(triggersByGroup[group.Name], role.Trigger.String)
						foundGroup = true
					}
				}
			}
		}
		if !foundGroup {
			log.Println("!!! WARNING !!! Failed to find group for a role! This is most likely a programming error")
			triggersByGroup["Uncategorized"] = append(triggersByGroup["Uncategorized"], role.Trigger.String)
		}
	}
	var message strings.Builder
	if len(triggersByGroup) == 0 {
		message.WriteString("Looks like there aren't any roles I can assign to you in this server!")
	} else {
		message.WriteString("This server's roles (highlighted `like this`): ")
		for groupName, triggerList := range triggersByGroup {
			message.WriteString("\nGroup (")
			message.WriteString(groupName)
			// TODO: add group type string here
			message.WriteString("): `")
			message.WriteString(strings.Join(triggerList, "`, `"))
			message.WriteString("`. ")
		}
	}
	pack.session.ChannelMessageSend(pack.channel.ID, message.String())
}
