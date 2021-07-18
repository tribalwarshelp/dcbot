package discord

import (
	"fmt"
	"strings"

	"github.com/tribalwarshelp/dcbot/message"
)

func boolToEmoji(val bool) string {
	if val {
		return ":white_check_mark:"
	}
	return ":x:"
}

func getAvailableLanguages() string {
	var langTags []string
	for _, langTag := range message.LanguageTags() {
		langTags = append(langTags, langTag.String())
	}
	return strings.Join(langTags, " | ")
}

func isValidLanguageTag(lang string) bool {
	for _, langTag := range message.LanguageTags() {
		if langTag.String() == lang {
			return true
		}
	}
	return false
}

func BuildLink(text string, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("[`%s`](%s)", text, url)
}
