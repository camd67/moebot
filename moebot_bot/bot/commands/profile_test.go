package commands

import (
	"database/sql"
	"testing"
)

func TestProfileCommand_ConvertRankToString(t *testing.T) {
	rankPrefixes := []string{"Newcomer", "Apprentice", "Rookie", "Regular", "Veteran"}
	checks := []struct {
		rank      int
		serverMax sql.NullInt64
		out       string
	}{
		{0, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0]},
		{1, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0]},
		{2, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 1"},
		{3, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 1"},
		{4, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 2"},
		{5, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 2"},
		{6, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 3"},
		{7, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 3"},
		{8, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 4"},
		{9, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 4"},
		{10, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1]},
		{11, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1]},
		{12, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1]},
		{13, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1]},
		{14, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " 1"},
		{30, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[2]},
		{60, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[3]},
		{100, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4]},
		{200, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 1"},
		{300, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 2"},
		{1500, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 14"},
	}
	for _, check := range checks {
		message := convertRankToString(check.rank, check.serverMax)
		if message != check.out {
			t.Errorf("Rank to string was incorrect, got: %s, want: %s.", message, check.out)
		}
	}
}
