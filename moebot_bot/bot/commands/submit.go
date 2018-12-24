package commands

import (
	"regexp"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

type SubmitCommand struct {
	ComPrefix string
}

func (sc *SubmitCommand) Execute(pack *CommPackage) {
	// Previous servers
	if pack.guild.ID == "378336255030722570" {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, submissions are closed!")
		return
	}
	// Salt
	if pack.guild.ID != "93799773856862208" {
		pack.session.ChannelMessageSend(pack.channel.ID, "Raffles are not enabled in this server! Speak to Salt to get your server added to the raffle!")
		return
	}

	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a submission type and a URL in order to submit a link.")
		return
	}
	// not a perfect pattern match, but if someone submits a link with a random "youtube.com" later in the url then it can be removed manually
	reg := regexp.MustCompile(".*(youtube.com|imgur.com|pastebin.com).*")
	if !reg.MatchString(pack.params[1]) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you must provide a link to an approved site! See submissions rules for more information")
		return
	}
	var raffleDataIndex int
	if strings.ToUpper(pack.params[0]) == "ART" {
		raffleDataIndex = 0
	} else if strings.ToUpper(pack.params[0]) == "RELIC" {
		raffleDataIndex = 1
	} else {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, I don't recognize that submission type. Valid types are: art, relic.")
		return
	}
	raffles, err := db.RaffleEntryQuery(pack.message.Author.ID, pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error trying to get your raffle information!")
		return
	}
	if len(raffles) != 1 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your raffle information! Make sure you're already in the raffle! "+
			"(Join the raffle first via `"+sc.ComPrefix+" raffle`)")
		return
	}
	raffleData := strings.Split(raffles[0].RaffleData, db.RaffleDataSeparator)
	var ticketsToAdd = 2
	if raffleData[raffleDataIndex] != "NONE" {
		// if they've already got a submission, don't award bonus tickets
		ticketsToAdd = 0
	}
	if raffleDataIndex == 0 {
		raffles[0].RaffleData = pack.params[1] + db.RaffleDataSeparator + raffleData[1]
	} else if raffleDataIndex == 1 {
		raffles[0].RaffleData = raffleData[0] + db.RaffleDataSeparator + pack.params[1]
	}
	db.RaffleEntryUpdate(raffles[0], ticketsToAdd)
	pack.session.ChannelMessageSend(pack.channel.ID, "Submission accepted!")
	pack.session.ChannelMessagePin(pack.channel.ID, pack.message.ID)
}

func (sc *SubmitCommand) GetPermLevel() types.Permission {
	return types.PermNone
}

func (sc *SubmitCommand) GetCommandKeys() []string {
	return []string{"SUBMIT"}
}

func (sc *SubmitCommand) GetCommandHelp(commPrefix string) string {
	return ""
}
