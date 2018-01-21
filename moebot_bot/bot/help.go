package bot

func commHelp(pack *commPackage) {
	if len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Moebot has the following commands:\n"+
			"`"+ComPrefix+" team <team name>` - Changes your team to one of the approved teams. `"+ComPrefix+" team` to list all teams\n"+
			"`"+ComPrefix+" rank <rank name>` - Changes your rank to one of the approved ranks. `"+ComPrefix+" rank` to list all the ranks\n"+
			"`"+ComPrefix+" changelog` - Displays the changelog for moebot\n"+
			"`"+ComPrefix+" NSFW` - Gives you NSFW channel permissions\n"+
			"`"+ComPrefix+" spoiler [<spoiler title>] <spoiler text>` - Creates a spoiler gif with the given text and (optional) title\n"+
			"`"+ComPrefix+" permit <perm level> <role name>` - Master/All only. Grants permission to the selected role.\n"+
			"`"+ComPrefix+" custom <command name> <role name>` - Master/All/Mod Links up a role to be toggable by the command name. Type `"+ComPrefix+" role <command name> to toggle`\n"+
			"`"+ComPrefix+" poll -options <option 1, option 2, option 3, ...> -title <poll title>` - Master/All/Mod set up a poll with the given options. Type `"+ComPrefix+" poll -close <poll id> to close`\n"+
			"`"+ComPrefix+" pinmove [-sendTo <#destChannel>] [-text] <#channel>` - Enables moving messages from the specified channel to the server's destination channel. The `-sendTo` option sets/changes the default destination channel. The `-text` option enables moving text on pin\n"+
			"`"+ComPrefix+" role <role name>` - Changes your role to one of the approved roles. `"+ComPrefix+" role` to list all the roles\n"+
			"`"+ComPrefix+" help` - Displays this message")
	}
}
