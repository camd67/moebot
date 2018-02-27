package commands

import (
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type HelpCommand struct {
	ComPrefix string
}

func (hc *HelpCommand) Execute(pack *CommPackage) {
	if len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Moebot has the following commands:\n"+
			"`"+hc.ComPrefix+" team <team name>` - Changes your team to one of the approved teams. `"+hc.ComPrefix+" team` to list all teams\n"+
			"`"+hc.ComPrefix+" rank <rank name>` - Changes your rank to one of the approved ranks. `"+hc.ComPrefix+" rank` to list all the ranks\n"+
			"`"+hc.ComPrefix+" changelog` - Displays the changelog for moebot\n"+
			"`"+hc.ComPrefix+" NSFW` - Gives you NSFW channel permissions\n"+
			"`"+hc.ComPrefix+" spoiler [<spoiler title>] <spoiler text>` - Creates a spoiler gif with the given text and (optional) title\n"+
			"`"+hc.ComPrefix+" permit <role name> [-permission <perm level>] [-securityAnswer <answer>] [-confirmationMessage <message>]` - Master/All only. Edits the selected role to grant permission or add a confirmation procedure.\n"+
			"`"+hc.ComPrefix+" custom <command name> <role name>` - Master/All/Mod Links up a role to be toggable by the command name. Type `"+hc.ComPrefix+" role <command name> to toggle`\n"+
			"`"+hc.ComPrefix+" poll -options <option 1, option 2, option 3, ...> -title <poll title>` - Master/All/Mod set up a poll with the given options. Type `"+hc.ComPrefix+" poll -close <poll id> to close`\n"+
			"`"+hc.ComPrefix+" pinmove [-sendTo <#destChannel>] [-text] <#channel>` - Enables moving messages from the specified channel to the server's destination channel. The `-sendTo` option sets/changes the default destination channel. The `-text` option enables moving text on pin\n"+
			"`"+hc.ComPrefix+" role <role name>` - Changes your role to one of the approved roles. `"+hc.ComPrefix+" role` to list all the roles\n"+
			"`"+hc.ComPrefix+" server <config setting> <value>` - Master/Mod Changes a config setting on the server to a given value. `"+hc.ComPrefix+" server` to list configs.\n"+
			"`"+hc.ComPrefix+" profile - Displays your server profile"+
			"`"+hc.ComPrefix+" help` - Displays this message")
	}
}

func (hc *HelpCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (hc *HelpCommand) GetCommandKeys() []string {
	return []string{"HELP"}
}
