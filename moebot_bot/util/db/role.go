package db

// Permission enum
type Permission int

const (
	_ Permission = iota
	PermAll
	PermMod
	PermNone
)

// RoleType enum
type RoleType int

const (
	_ RoleType = iota
	// the starter role you get when joining a server (if enabled)
	RoleStarter
	// the default role you get AFTER agreeing to the rules
	RoleDefault
	RoleRank
	RoleTeam
	RoleNone
)

type Role struct {
	id         int
	serverId   int
	RoleUid    string
	Permission Permission
	RoleType   RoleType
}

const roleTable = `CREATE TABLE IF NOT EXISTS role(
		id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER REFERENCES server(id) ON DELETE CASCADE,
		RoleId VARCHAR(20) NOT NULL UNIQUE,
		Permission SMALLINT NOT NULL,
		RoleType SMALLINT NOT NULL
	)`
