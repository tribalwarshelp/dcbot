package discord

import (
	"fmt"
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

func FormatLink(text string, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("[``%s``](%s)", text, url)
}