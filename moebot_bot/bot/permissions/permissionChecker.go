package permissions

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermissionChecker struct {
	MasterId string
}

func (p *PermissionChecker) HasAllPerm(userId string, roles []string) bool {
	return p.HasPermission(userId, roles, db.PermAll)
}

func (p *PermissionChecker) HasModPerm(userId string, roles []string) bool {
	return p.HasPermission(userId, roles, db.PermMod)
}

func (p *PermissionChecker) HasPermission(userId string, roles []string, pToCheck db.Permission) bool {
	if pToCheck == db.PermAll {
		// if everyone can use this command, just allow it
		return true
	} else if p.IsMaster(userId) {
		// masters are allowed to do anything
		return true
	} else if pToCheck == db.PermNone {
		// if no one can use this command, never do it
		// make sure this is the last thing to check before
		return false
	}
	// if any of the previous checks fails, then go ahead and check the database for their permission
	perms := db.RoleQueryPermission(roles)
	for _, p := range perms {
		if p <= pToCheck {
			return true
		}
	}
	return false
}

func (p *PermissionChecker) IsGuildOwner(guild *discordgo.Guild, id string) bool {
	return guild.OwnerID == id
}

func (p *PermissionChecker) IsMaster(id string) bool {
	return p.MasterId == id
}
