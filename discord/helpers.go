package discord

import (
	"strings"

	"github.com/tribalwarshelp/dcbot/message"
)

func getAvailableLanguages() string {
	langTags := []string{}
	for _, langTag := range message.LanguageTags() {
		langTags = append(langTags, langTag.String())
	}
	return strings.Join(langTags, " | ")
}
