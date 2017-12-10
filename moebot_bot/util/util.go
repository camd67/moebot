package util

import (
	"regexp"
	"strings"
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
