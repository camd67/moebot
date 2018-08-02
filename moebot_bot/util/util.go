/*
General utility functions for moebot.

If there are any imports in this package that aren't part of the standard libary (or a /x/ or golang repoo) then that's a signal to move to a
more specific package within util such as event or moeDiscord.
*/
package util

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	CaseInsensitive = iota
	CaseSensitive
)

type SyncUIDByChannelMap struct {
	sync.RWMutex
	M map[string][]string
}

type SyncCooldownMap struct {
	sync.RWMutex
	M map[string]int64
}

func IntContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func StrContains(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.EqualFold(e, a) {
				return true
			}
		} else {
			if a == e {
				return true
			}
		}
	}
	return false
}

func StrContainsPrefix(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.HasPrefix(strings.ToUpper(a), strings.ToUpper(e)) {
				return true
			}
		} else {
			if strings.HasPrefix(a, e) {
				return true
			}
		}
	}
	return false
}

func MakeAlphaOnly(s string) string {
	reg := regexp.MustCompile("[^A-Za-z ]+")
	return reg.ReplaceAllString(s, "")
}

func NormalizeNewlines(s string) string {
	reg := regexp.MustCompile("(\r\n|\r|\n)")
	return reg.ReplaceAllString(s, "\n")
}

/*
Converts a user's ID into a mention.
This is useful when you don't have a User object, but want to mention them
*/
func UserIdToMention(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
}

func ExtractChannelIdFromString(message string) (id string, valid bool) {
	// channelIds go with the format of <#1234567>
	if len(message) < 2 || len(message) > 23 {
		return "", false
	}
	id = message[2 : len(message)-1]
	_, err := strconv.ParseInt(id, 10, 64)
	return id, err == nil
}

func MakeStringBold(s string) string {
	return "**" + s + "**"
}

func MakeStringItalic(s string) string {
	return "_" + s + "_"
}

func MakeStringStrikethrough(s string) string {
	return "~~" + s + "~~"
}

func MakeStringCode(s string) string {
	return "`" + s + "`"
}

func GetStringOrDefault(s sql.NullString) string {
	if s.Valid {
		return s.String
	} else {
		return "unknown"
	}
}

func GetInt64OrDefault(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	} else {
		return -1
	}
}
