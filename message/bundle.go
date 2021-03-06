package message

import (
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"golang.org/x/text/language"
)

var lang = language.English
var bundle = i18n.NewBundle(lang)

func GetDefaultLanguage() language.Tag {
	return lang
}

func NewLocalizer(l ...string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, append(l, lang.String())...)
}

func LanguageTags() []language.Tag {
	return bundle.LanguageTags()
}

func LoadMessages(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path != root {
			bundle.MustLoadMessageFile(path)
		}
		return nil
	})
}
