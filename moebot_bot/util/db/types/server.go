package types

import "database/sql"

/*
A Server (guild in discord terms) that stores information related to what over server settings are.
*/
type Server struct {
	Id             int
	GuildUid       string
	WelcomeMessage sql.NullString // Message user gets sent when they first join the server, either via PM or public message depending on WelcomeChannel
	RuleAgreement  sql.NullString // Message to type when the user agrees to the rules
	VeteranRank    sql.NullInt64  // Rank at which a user can be promoted to the veteran role
	VeteranRole    sql.NullString // Role to apply when a user can become a veteran
	BotChannel     sql.NullString // Where any bot related information or errors get sent to.
	Enabled        bool           // defaults to true, so that new servers that add moebot can immediately start using her. This can be turned off later
	WelcomeChannel sql.NullString // Channel to post a welcome message. If null, send via PM's
	StarterRole    sql.NullString // The role that is added when someone first joins a server
	BaseRole       sql.NullString // The role that is added when someone types the RuleAgreement message. Should only exist when RuleAgreement isn't null
}
