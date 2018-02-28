package commands

import (
	"strings"
)

func ParseCommand(commandLine string, arguments []string) map[string]string {
	result := make(map[string]string)
	slicedCommand := strings.Split(commandLine, " ")
	var currentCommand, currentCommandContent string
	for i, s := range slicedCommand {
		if isArgument(s, arguments) {
			if currentCommandContent != "" {
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
			if i == len(slicedCommand)-1 && currentCommandContent != "" {
				result[currentCommand] = currentCommandContent
			}
		}
	}
	return result
}

func isArgument(s string, arguments []string) bool {
	for _, a := range arguments {
		if s == a {
			return true
		}
	}
	return false
}
