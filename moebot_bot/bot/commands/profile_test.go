package commands

import (
	"database/sql"
	"strings"
	"testing"
)

func TestProfileCommand_ConvertRankToString(t *testing.T) {
	rankPrefixes := []string{"Newcomer", "Apprentice", "Rookie", "Regular", "Veteran"}
	rankSeparator := "-->"
	checks := []struct {
		rank      int
		serverMax sql.NullInt64
		out       string
	}{
		// Check each one 0 - 10
		{0, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{1, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{2, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 1" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{3, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 1" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{4, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 2" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{5, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 2" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{6, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 3" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{7, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 3" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{8, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 4" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		{9, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[0] + " 4" + " " + strings.Join(rankPrefixes[0:], rankSeparator)},
		// then check just every 1st character
		{10, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{11, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		// also check some funky numbers
		{7, sql.NullInt64{Int64: 63, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{27, sql.NullInt64{Int64: 243, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{12, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{13, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{14, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[1] + " 1" + " " + strings.Join(rankPrefixes[1:], rankSeparator)},
		{30, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[2] + " " + strings.Join(rankPrefixes[2:], rankSeparator)},
		{60, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[3] + " " + strings.Join(rankPrefixes[3:], rankSeparator)},
		{100, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " " + strings.Join(rankPrefixes[4:], rankSeparator)},
		{200, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 1" + " " + strings.Join(rankPrefixes[4:], rankSeparator)},
		{300, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 2" + " " + strings.Join(rankPrefixes[4:], rankSeparator)},
		// make sure the last rank keeps going
		{1500, sql.NullInt64{Int64: 100, Valid: true}, rankPrefixes[4] + " 14" + " " + strings.Join(rankPrefixes[4:], rankSeparator)},
	}
	for _, check := range checks {
		message := convertRankToString(check.rank, check.serverMax)
		if message != check.out {
			t.Errorf("Rank to string was incorrect, got: %s, want: %s.", message, check.out)
		}
	}
}
