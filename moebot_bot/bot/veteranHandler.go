package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	messagePoint     = 5
	reactionPoint    = 1
	reactionCooldown = time.Second
)

var (
	reactionCooldownMap = make(map[string]int64)
)

func handleVeteranMessage(m *discordgo.Member, guildUid string) {
	handleVeteranChange(m.User.ID, guildUid, messagePoint)
}

func handleVeteranReaction(userUid string, guildUid string) {
	if lastTime, present := reactionCooldownMap[userUid]; present {
		if lastTime+reactionCooldown.Nanoseconds() > time.Now().UnixNano() {
			return
		}
	}
	// if they don't have a time yet, or the cooldown was passed, give them a new time
	reactionCooldownMap[userUid] = time.Now().UnixNano()
	handleVeteranChange(userUid, guildUid, reactionPoint)
}

func handleVeteranChange(userUid string, guildUid string, point int) {

}
