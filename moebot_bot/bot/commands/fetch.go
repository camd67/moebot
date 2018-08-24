package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

/*
Fetches information about something within discord. Useful for debugging why something isn't working correctly.
*/
type FetchCommand struct {
	MasterId string
}

func (fc *FetchCommand) Execute(pack *CommPackage) {
	args := ParseCommand(pack.params, []string{"-item", "-arg1", "-arg2", "-fromCache"})

	item, hasItem := args["-item"]
	arg1, hasArg1 := args["-arg1"]
	arg2, hasArg2 := args["-arg2"]
	_, hasFromCache := args["-fromCache"]

	if !hasItem || !hasArg1 {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "This command requires both an -item and an -arg.")
		return
	}

	item = strings.ToUpper(item)
	switch item {
	case "MEMBER":
		if !hasArg2 {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a guildID and a userID as the two args.")
			return
		}
		var guildMember *discordgo.Member
		var err error
		// Force getting from cache or not, and skip over our caching functions
		if hasFromCache {
			guildMember, err = pack.session.State.Member(arg1, arg2)
		} else {
			guildMember, err = pack.session.GuildMember(arg1, arg2)
		}
		if err != nil {
			if err == discordgo.ErrStateNotFound {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "The given guild/user ID combo isn't known to moebot's cache.")
			} else {
				pack.session.ChannelMessageSend(pack.message.ChannelID, "The given guild/user ID combo don't match a known guild member for moebot.")
			}
			return
		}
		var message strings.Builder
		if hasFromCache {
			message.WriteString("`{From Cache!}` ")
		}
		message.WriteString("Information on: ")
		message.WriteString(guildMember.User.Username)
		message.WriteString(" (#")
		message.WriteString(guildMember.User.Discriminator)
		message.WriteString(")\n")
		message.WriteString("JoinedAt: `")
		message.WriteString(string(guildMember.JoinedAt))
		if guildMember.Nick != "" {
			message.WriteString("`\nNickname: `")
			message.WriteString(guildMember.Nick)
		}
		message.WriteString("`\nRole IDs: `")
		message.WriteString(strings.Join(guildMember.Roles, ", "))
		message.WriteString("`")
		pack.session.ChannelMessageSend(pack.message.ChannelID, message.String())
		break
	default:
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Unknown item. Please provide a supported item type.")
		break
	}
}
func (fc *FetchCommand) GetPermLevel() db.Permission {
	// Temporarily only for master until further investigation on if it's not too much information...
	return db.PermMaster
}
func (fc *FetchCommand) GetCommandKeys() []string {
	return []string{"FETCH"}
}
func (fc *FetchCommand) GetCommandHelp(commPrefix string) string {
	return "`" + commPrefix + " fetch` - Fetches information on something within discord."
}
