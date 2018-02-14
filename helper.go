package main

import "fmt"

func PrintQuotedString(str string) string {
	return fmt.Sprintf("%q", str)
}

func BooleanToString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}
