package message

import "github.com/nicksnyder/go-i18n/v2/i18n"

func FallbackMsg(id string, msg string) *i18n.Message {
	return &i18n.Message{
		ID:    id,
		One:   msg,
		Many:  msg,
		Other: msg,
	}
}
