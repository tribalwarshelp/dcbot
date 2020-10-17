package discord

import (
	"fmt"
	"strings"

	"github.com/tribalwarshelp/dcbot/message"
)

func getEmojiForGroupsCommand(val bool) string {
	if val {
		return ":white_check_mark:"
	}
	return ":x:"
}

func getAvailableLanguages() string {
	langTags := []string{}
	for _, langTag := range message.LanguageTags() {
		langTags = append(langTags, langTag.String())
	}
	return strings.Join(langTags, " | ")
}

func isValidLanguageTag(lang string) bool {
	valid := false
	for _, langTag := range message.LanguageTags() {
		if langTag.String() == lang {
			valid = true
			break
		}
	}
	return valid
}

func FormatLink(text string, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("[``%s``](%s)", text, url)
}
