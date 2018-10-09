/*
General utility functions for moebot.

If there are any imports in this package that aren't part of the standard libary (or a /x/ or golang repoo) then that's a signal to move to a
more specific package within util such as event or moeDiscord.
*/
package util

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"

	"github.com/bwmarrin/discordgo"
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

func RetrieveBasePermissions(session *discordgo.Session, channel *discordgo.Channel, role *discordgo.Role, flags []int) map[int]bool {
	result := make(map[int]bool)
	permission, ok := moeDiscord.FindPermissionByRoleID(channel.PermissionOverwrites, role.ID)
	if ok {
		mapPermissions(result, permission, flags)
	}
	if !ok || unsetFlags(permission, flags) { //no overwrite defined for the channel, looking in parent category
		parent, _ := session.Channel(channel.ParentID)
		permission, ok = moeDiscord.FindPermissionByRoleID(parent.PermissionOverwrites, role.ID)
		if ok {
			mapPermissions(result, permission, flags)
		}
		if !ok || unsetFlags(permission, flags) { //no overwrite defined for the channel, using role permissions
			permission = &discordgo.PermissionOverwrite{
				ID:   role.ID,
				Type: "role",
			}
			for _, f := range flags {
				if role.Permissions&f != 0 {
					permission.Allow = permission.Allow | f
				} else {
					permission.Deny = permission.Deny | f
				}
			}
			mapPermissions(result, permission, flags)
		}
	}
	return result
}

func GetEveryoneRoleForServer(session *discordgo.Session, serverID int) *discordgo.Role {
	server, err := db.ServerQueryById(serverID)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to retrieve server informations for Server ID: %v. ", serverID), err)
		return nil
	}
	roles, err := session.GuildRoles(server.GuildUid)
	if err != nil {
		log.Println("Failed to retrieve roles informations for Guild UID: "+server.GuildUid+". ", err)
		return nil
	}
	return moeDiscord.FindRoleByName(roles, "@everyone")
}

func GetCurrentRolePermissionsForChannel(session *discordgo.Session, channelUID string, roleUID string) (*discordgo.PermissionOverwrite, error) {
	channel, err := session.Channel(channelUID)
	if err != nil {
		return nil, err
	}
	if p, ok := moeDiscord.FindPermissionByRoleID(channel.PermissionOverwrites, roleUID); !ok {
		return &discordgo.PermissionOverwrite{
			ID:   roleUID,
			Type: "role",
		}, nil
	} else {
		return p, nil
	}
}

func unsetFlags(permission *discordgo.PermissionOverwrite, flags []int) bool {
	for _, f := range flags {
		if permission.Allow&f == 0 && permission.Deny&f == 0 {
			return true
		}
	}
	return false
}

func mapPermissions(base map[int]bool, permission *discordgo.PermissionOverwrite, flags []int) {
	for _, f := range flags {
		if _, ok := base[f]; !ok { //only do this if the flag is unset, to allow hierarchy assignations
			if permission.Allow&f != 0 || permission.Deny&f != 0 {
				base[f] = permission.Allow&f != 0
			}
		}
	}
}
