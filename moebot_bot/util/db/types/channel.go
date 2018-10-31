package types

import "database/sql"

type Channel struct {
	Id             int
	ServerId       int
	ChannelUid     string
	BotAllowed     bool
	MovePins       bool
	MoveTextPins   bool
	DeletePin      bool
	MoveChannelUid sql.NullString
}
