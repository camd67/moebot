package bot

func commHelp(pack *commPackage) {
	if len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Moebot has the following commands:\n"+
			"`"+ComPrefix+" team <role name>` - Changes your role to one of the approved roles. `"+ComPrefix+" team` to list all teams\n"+
			"`"+ComPrefix+" rank <rank name>` - Changes your rank to one of the approved ranks. `"+ComPrefix+" rank` to list all the ranks\n"+
			"`"+ComPrefix+" changelog` - Displays the changelog for moebot\n"+
			"`"+ComPrefix+" NSFW` - Gives you NSFW channel permissions\n"+
			"`"+ComPrefix+" help` - Displays this message")
	}
}
