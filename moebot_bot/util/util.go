package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	CaseInsensitive = iota
	CaseSensitive
)

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

func FindRoleByName(roles []*discordgo.Role, toFind string) *discordgo.Role {
	toFind = strings.ToUpper(toFind)
	for _, r := range roles {
		if strings.ToUpper(r.Name) == toFind {
			return r
		}
	}
	return nil
}

func FindRoleById(roles []*discordgo.Role, toFind string) *discordgo.Role {
	// for some reason roleIds have spaces in them...
	toFind = strings.TrimSpace(toFind)
	for _, r := range roles {
		if r.ID == toFind {
			return r
		}
	}
	return nil
}

func GetSpoilerContents(messageParams []string) (title string, text string) {
	if messageParams == nil {
		return "", ""
	}
	reg := regexp.MustCompile("^(\\[.+?\\])")
	return strings.Replace(strings.Replace(reg.FindString(strings.Join(messageParams, " ")), "]", "", 1), "[", "", 1), reg.ReplaceAllString(strings.Join(messageParams, " "), "")
}
