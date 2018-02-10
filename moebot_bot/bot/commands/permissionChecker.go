package commands

import (
	"log"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermissionChecker struct {
	MasterId string
}

func (p *PermissionChecker) HasAllPerm(userId string, roles []string) bool {
	return p.hasPermission(userId, roles, db.PermAll)
}

func (p *PermissionChecker) HasModPerm(userId string, roles []string) bool {
	return p.hasPermission(userId, roles, db.PermMod)
}

func (p *PermissionChecker) hasPermission(userId string, roles []string, pToCheck db.Permission) bool {
	// masters are allowed to do anything
	if p.isMaster(userId) {
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

func (p *PermissionChecker) isMaster(id string) bool {
	return id == p.MasterId
}

func (p *PermissionChecker) hasValidMasterId(pack *CommPackage) bool {
	if !p.isMaster(pack.message.Author.ID) {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, only my master can use this command!")
		return false
	}
	return true
}
