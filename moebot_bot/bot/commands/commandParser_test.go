package commands

import (
	"reflect"
	"testing"
)

type checkType struct {
	test     string
	testArgs []string
	expected map[string]string
}

type testArgs struct {
	argName string
	value   string
}

func TestCommandParser_ParseCommand(t *testing.T) {
	resultMaps := [][]testArgs{
		[]testArgs{{"", ""}},
		[]testArgs{{"", "test"}},
		[]testArgs{{"", "test"}, {"-arg1", "test1"}, {"-arg2", "test2"}},
		[]testArgs{{"-arg1", "test arg1 multiple words"}, {"-arg2", "test2 hi there"}},
		[]testArgs{{"-arg1", "test1 test aaa"}},
		[]testArgs{{"", "test"}, {"-arg1", "test\r\nwith\r\nlinebreak"}},
	}
	checks := []checkType{}
	for _, m := range resultMaps {
		var check checkType
		check.expected = make(map[string]string)
		for _, ts := range m {
			if ts.argName != "" {
				check.test += " " + ts.argName + " " + ts.value
				check.testArgs = append(check.testArgs, ts.argName)
			} else {
				check.test += ts.value
			}
			if !(ts.argName == "" && ts.value == "") {
				check.expected[ts.argName] = ts.value
			}
		}
		checks = append(checks, check)
	}
	for _, check := range checks {
		result := ParseCommand(check.test, check.testArgs)
		t.Logf("Tested '%s'", check.test)
		if !reflect.DeepEqual(check.expected, result) {
			t.Errorf("Command parse was incorrect, got: %v, want: %v.", result, check.expected)
		}
	}
}
