package models

import "strings"

func addPrefixToColumnName(column, prefix string) string {
	if prefix != "" && !strings.HasPrefix(column, prefix+".") {
		column = wrapStringInDoubleQuotes(prefix) + "." + wrapStringInDoubleQuotes(column)
	} else {
		column = wrapStringInDoubleQuotes(column)
	}
	return column
}

func wrapStringInDoubleQuotes(str string) string {
	return `"` + str + `"`
}
