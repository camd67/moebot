package bot

func commHelp(pack *commPackage) {
	pack.session.ChannelMessageSend(pack.channel.ID, "Moebot has the following commands:\n"+
		"`"+ComPrefix+" team <role name>` - Changes your role to one of the approved roles. `"+ComPrefix+" team` to list all teams\n"+
		"`"+ComPrefix+" rank <rank name>` - Changes your rank to one of the approved ranks. `"+ComPrefix+" rank` to list all the ranks\n"+
		"`"+ComPrefix+" changelog` - Displays the changelog for moebot\n"+
		"`"+ComPrefix+" raffle` - Enters you into the raffle (if enabled on the server). Displays ticket count if already in the raffle\n"+
		"`"+ComPrefix+" submit <TYPE> <URL>` - Submits a link for relic/art competitions! Valid types are: art, relic. Valid URLS are from youtube, pastebin, and imgur.\n"+
		"`"+ComPrefix+" NSFW` - Gives you NSFW channel permissions\n"+
		"`"+ComPrefix+" help` - Displays this message")
}
