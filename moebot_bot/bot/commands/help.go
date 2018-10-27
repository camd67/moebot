package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type HelpCommand struct {
	ComPrefix string
	Commands  func() []Command
	Checker   permissions.PermissionChecker
}

func (hc *HelpCommand) Execute(pack *CommPackage) {
	if len(pack.params) == 1 {
		// User should've given a command to get details for
		hc.listCommandDetails(pack)
	} else {
		// Any errors, bad params, or no params will just list everything off
		hc.listAllCommands(pack)
	}
}

// Lists details for a specific command, if the user has permission, in the given package
func (hc *HelpCommand) listCommandDetails(pack *CommPackage) {
	requestedCommandKey := strings.ToUpper(pack.params[0])
	for _, command := range hc.Commands() {
		if !hc.currentUserHasCommandPermission(pack, command) {
			// Skip any commands which the user doesn't have access to
			continue
		}
		for _, commandKey := range command.GetCommandKeys() {
			if commandKey == requestedCommandKey {
				var message strings.Builder
				message.WriteString("**Details for command**: `")
				message.WriteString(util.ForceTitleCase(commandKey))
				message.WriteString("`:\n")
				message.WriteString(command.GetCommandHelp(hc.ComPrefix))
				pack.session.ChannelMessageSend(pack.channel.ID, message.String())
				return
			}
		}
	}
	// Getting to this point means we didn't find the message. List all commands
	hc.listAllCommands(pack)
}

// Lists out all commands for the given user
func (hc *HelpCommand) listAllCommands(pack *CommPackage) {
	var message strings.Builder
	message.WriteString("For details on each command, use the detailed help command or check out the wiki! <https://github.com/camd67/moebot/wiki>\n")
	message.WriteString("**Moebot has the following commands**:\n")
	for commIndex, command := range hc.Commands() {
		if hc.currentUserHasCommandPermission(pack, command) {
			for commKeyIndex, commKey := range command.GetCommandKeys() {
				// Don't add a , unless we're not the first command
				if commIndex > 0 || commKeyIndex > 0 {
					message.WriteString(", ")
				}
				message.WriteString("`")
				message.WriteString(util.ForceTitleCase(commKey))
				message.WriteString("`")
			}
		}
	}
	message.WriteString("\nYou can also use `")
	message.WriteString(hc.ComPrefix)
	message.WriteString(" help <command name>` for more details.")
	_, err := pack.session.ChannelMessageSend(pack.channel.ID, message.String())
	if err != nil {
		log.Println("An error occurred sending help message to channel: "+pack.channel.ID+" with error: ", err)
		return
	}
}

func (hc *HelpCommand) currentUserHasCommandPermission(pack *CommPackage, command Command) bool {
	return hc.Checker.HasPermission(pack.message.Author.ID, pack.member.Roles, pack.guild, command.GetPermLevel()) &&
		command.GetCommandHelp(hc.ComPrefix) != ""
}

func (hc *HelpCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (hc *HelpCommand) GetCommandKeys() []string {
	return []string{"HELP"}
}

func (hc *HelpCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s help` - Displays this message", commPrefix)
}
