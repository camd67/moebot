package bot

const changeLogPrefix = "\n`->` "

// probably want to move this to the DB, but not bad to have it here
var changeLog = map[string]string{

	"0.2.4": changeLogPrefix + "Added spoiler command" +
		changeLogPrefix + "Added poll command",

	"0.2.3": changeLogPrefix + "Added ping command" +
		changeLogPrefix + "Added permit command" +
		changeLogPrefix + "Added custom command",

	"0.2.2": changeLogPrefix + "Added echo command for master only" +
		changeLogPrefix + "added `raffle winner` and `raffle count` to get the raffle winner and vote counts" +
		changeLogPrefix + "removed ticket generation",

	"0.2.1": changeLogPrefix + "Updated raffle art/relic submissions to post all submissions on command instead of over time.",

	"0.2": changeLogPrefix + "Included this command!" +
		changeLogPrefix + "Updated `Rank` command to prevent removal of lowest role." +
		changeLogPrefix + "Added random drops for tickets" +
		changeLogPrefix + "Fixed the cooldown so users wouldn't be spammed due to high luck stat" +
		changeLogPrefix + "Added `Raffle` related commands... For rafflin'" +
		changeLogPrefix + "For future reference, previous versions included help, team, rank, and NSFW commands as well as a welcome message to the server.",
}

func commChange(pack *commPackage) {
	if len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Moebot update log `(ver "+version+")`: \n"+changeLog[version])
	} else if log, present := changeLog[pack.params[0]]; present {
		pack.session.ChannelMessageSend(pack.channel.ID, "Moebot update log `(ver "+pack.params[0]+")`: \n"+log)
	} else {
		pack.session.ChannelMessageSend(pack.channel.ID, "Unknown version number. Latest log:\nMoebot update log `(ver "+version+")`: \n"+changeLog[version])
	}
}
