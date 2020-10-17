package models

import "strings"

func addPrefixToFieldName(field, prefix string) string {
	if prefix != "" && !strings.HasPrefix(field, prefix+".") {
		field = wrapStringInDoubleQuotes(prefix) + "." + wrapStringInDoubleQuotes(field)
	} else {
		field = wrapStringInDoubleQuotes(field)
	}
	return field
}

func wrapStringInDoubleQuotes(str string) string {
	return `"` + str + `"`
}
