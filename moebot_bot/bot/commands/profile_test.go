package commands

import (
	"database/sql"
	"testing"
)

func TestProfileCommand_ConvertRankToString(t *testing.T) {
	checks := []struct {
		rank      int
		serverMax sql.NullInt64
		out       string
	}{
		// I don't like making this hard coded... but it works...
		// Check each one 0 - 10
		{0, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{1, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{2, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 1** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{3, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 1** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{4, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 2** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{5, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 2** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{6, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 3** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{7, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 3** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{8, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 4** --> Apprentice --> Rookie --> Regular --> Veteran"},
		{9, sql.NullInt64{Int64: 100, Valid: true}, "**Newcomer 4** --> Apprentice --> Rookie --> Regular --> Veteran"},
		// then check just every 1st character
		{10, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		{11, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		// also check some funky numbers
		{7, sql.NullInt64{Int64: 63, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		{27, sql.NullInt64{Int64: 243, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		{12, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		{13, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> **Apprentice** --> Rookie --> Regular --> Veteran"},
		{14, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> **Apprentice 1** --> Rookie --> Regular --> Veteran"},
		{30, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> **Rookie** --> Regular --> Veteran"},
		{60, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> ~~Rookie~~ --> **Regular** --> Veteran"},
		{100, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> ~~Rookie~~ --> ~~Regular~~ --> **Veteran**"},
		{200, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> ~~Rookie~~ --> ~~Regular~~ --> **Veteran 1**"},
		{300, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> ~~Rookie~~ --> ~~Regular~~ --> **Veteran 2**"},
		// make sure the last rank keeps going
		{1500, sql.NullInt64{Int64: 100, Valid: true}, "~~Newcomer~~ --> ~~Apprentice~~ --> ~~Rookie~~ --> ~~Regular~~ --> **Veteran 14**"},
	}
	for _, check := range checks {
		message := convertRankToString(check.rank, check.serverMax)
		if message != check.out {
			t.Errorf("Rank to string was incorrect, got: %s, want: %s.", message, check.out)
		}
	}
}
