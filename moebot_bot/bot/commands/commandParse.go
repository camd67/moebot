package commands

import "strings"

/*
Parses a given param list, mapping them to the given arguments if possible.
*/
func ParseCommand(commParams []string, arguments []string) map[string]string {
	result := make(map[string]string)
	var currentCommand, currentCommandContent string
	for i, s := range commParams {
		if isArgument(s, arguments) {
			if currentCommand != "" || currentCommandContent != "" {
				result[currentCommand] = currentCommandContent
			}
			currentCommand = s
			currentCommandContent = ""
		} else {
			if currentCommandContent == "" {
				currentCommandContent = s
			} else {
				currentCommandContent += " " + s
			}
		}
		if i == len(commParams)-1 && (currentCommand != "" || currentCommandContent != "") {
			result[currentCommand] = currentCommandContent
		}
	}
	return result
}

func isArgument(s string, arguments []string) bool {
	for _, a := range arguments {
		if strings.EqualFold(s, a) {
			return true
		}
	}
	return false
}
