package util

import "strings"

func IntContains(s []int, e int) bool {
	return IntIndexOf(s, e) > -1
}

func IntRemove(s []int, e int) []int {
	index := IntIndexOf(s, e)
	if index > -1 {
		s = append(s[:index], s[index+1:]...)
	}
	return s
}

func IntIndexOf(s []int, e int) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
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

//Subtract removes all values in slice2 from slice1
func Subtract(slice1 []int, slice2 []int) []int {
	var result []int
	for _, a := range slice1 {
		if !IntContains(slice2, a) {
			result = append(result, a)
		}
	}
	return result
}
