package db

type Server struct {
	id             int
	GuildUid       string
	WelcomeMessage string
	RuleAgreement  string
}

const serverTable = `CREATE TABLE IF NOT EXISTS server(
		id SERIAL NOT NULL PRIMARY KEY,
		GuildUid VARCHAR(20) NOT NULL UNIQUE,
		WelcomeMessage VARCHAR(500),
		RuleAgreement VARCHAR(50)
	)`
