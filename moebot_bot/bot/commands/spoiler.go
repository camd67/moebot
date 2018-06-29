package commands

import (
	"bytes"
	"log"
	"regexp"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
)

type SpoilerCommand struct{}

func (sc *SpoilerCommand) Execute(pack *CommPackage) {
	content := pack.message.Author.Mention() + " sent a spoiler"
	for i := 0; i < 2; i++ {
		err := pack.session.ChannelMessageDelete(pack.channel.ID, pack.message.ID)
		if err == nil {
			break
		}
		log.Println("Error while deleting message", err)
	}

	spoilerTitle, spoilerText := getSpoilerContents(pack.params)
	if spoilerTitle != "" {
		content += ": **" + spoilerTitle + "**"
	}
	spoilerGif := util.MakeGif(spoilerText)
	pack.session.ChannelMessageSendComplex(pack.channel.ID, &discordgo.MessageSend{
		Content: content,
		File: &discordgo.File{
			Name:        "Spoiler.gif",
			ContentType: "image/gif",
			Reader:      bytes.NewReader(spoilerGif),
		},
	})
}

func (sc *SpoilerCommand) GetPermLevel() db.Permission {
	return db.PermNone
}

func (sc *SpoilerCommand) GetCommandKeys() []string {
	return []string{"SPOILER"}
}

func (sc *SpoilerCommand) GetCommandHelp(commPrefix string) string {
	return ""
}

func getSpoilerContents(messageParams []string) (title string, text string) {
	if messageParams == nil {
		return "", ""
	}
	reg := regexp.MustCompile("^(\\[.+?\\])")
	return strings.Replace(strings.Replace(reg.FindString(strings.Join(messageParams, " ")), "]", "", 1), "[", "", 1), reg.ReplaceAllString(strings.Join(messageParams, " "), "")
}
