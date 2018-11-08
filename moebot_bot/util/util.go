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
	_, err := strconv.Atoi(id)
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

// Forces title case for the given string. This is necessary when your string may contain upper case characters
func ForceTitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

func GetInt64OrDefault(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	} else {
		return -1
	}
}

func StringIndexOf(s []string, search string) int {
	for i, v := range s {
		if v == search {
			return i
		}
	}
	return -1
}

func ParseIntervalToISO(interval string) (string, error) {
	intervalsOrder := []string{"Y", "M", "W", "D", "h", "m"}
	//this regex is parsing interval strings in the format of nYnMnWnDnhnm, for example 1Y2M3W4D5h6m.
	//later, the string is converted into ISO interval format, so (same example input) P1Y2M3W4DT5H6M
	rx, _ := regexp.Compile("^(\\d+Y){0,1}(\\d+M){0,1}(\\d+W){0,1}(\\d+D){0,1}(\\d+h){0,1}(\\d+m){0,1}$")
	if !rx.MatchString(interval) {
		return "", fmt.Errorf("Invalid interval string")
	}

	matches := rx.FindAllStringSubmatch(interval, -1)[0][1:]
	var b strings.Builder
	b.WriteString("P")
	for _, indicator := range intervalsOrder {
		for _, match := range matches {
			if strings.Contains(match, indicator) {
				b.WriteString(strings.ToUpper(match))
			}
		}
		if indicator == "D" { //Adds time separator after day segment
			if b.Len() != 1 {
				b.WriteString("T")
			} else {
				b.WriteString("0DT")
			}
		}
	}
	intervalString := strings.Trim(b.String(), "T")
	return intervalString, nil
}
