package types

import "database/sql"

// Permission enum
type Permission int

const RoleCodeSearchText = "[code]"
const RoleCodeLength = 6

const (
	// Default permission level, no permissions regarding what can or can't be done
	PermAll Permission = 2
	// Mod level permission, allowed to do some server changing commands
	PermMod Permission = 50
	// Guild Owner permission. Essentially a master
	PermGuildOwner Permission = 90
	// Used to disable something, no one can have this permission level
	PermNone Permission = 100
	// Master level permission, can't ever be ignored or disabled
	PermMaster Permission = 101
)

type Role struct {
	Id                         int
	ServerId                   int
	Groups                     []int
	RoleUid                    string
	Permission                 Permission
	ConfirmationMessage        sql.NullString
	ConfirmationSecurityAnswer sql.NullString
	Trigger                    sql.NullString
}
