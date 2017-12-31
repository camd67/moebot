package bot

import (
	"log"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

func HasAllPerm(userId string, roles []string) bool {
	return hasPermission(userId, roles, db.PermAll)
}

func HasModPerm(userId string, roles []string) bool {
	return hasPermission(userId, roles, db.PermMod)
}

func hasPermission(userId string, roles []string, pToCheck db.Permission) bool {
	// masters are allowed to do anything
	if isMaster(userId) {
		return true
	}
	perms := db.RoleQueryPermission(roles)
	for _, p := range perms {
		log.Println(p)
		if p <= pToCheck {
			return true
		}
	}
	return false
}

func isMaster(id string) bool {
	return id == Config["masterId"]
}

func checkValidMasterId(pack *commPackage) bool {
	if isMaster(pack.message.Author.ID) {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, only my master can use this command!")
		return false
	}
	return true
}
