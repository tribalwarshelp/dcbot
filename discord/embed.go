package discord

import (
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const (
	EmbedColor            = 0x00ff00
	EmbedLimitTitle       = 256
	EmbedLimitDescription = 2048
	EmbedLimitFieldValue  = 1024
	EmbedLimitFieldName   = 256
	EmbedLimitField       = 25
	EmbedLimitFooter      = 2048
	EmbedSizeLimit        = 4500
)

type Embed struct {
	*discordgo.MessageEmbed
}

//NewEmbed returns a new embed object
func NewEmbed() *Embed {
	return &Embed{&discordgo.MessageEmbed{
		Color: EmbedColor,
	}}
}

func (e *Embed) SetTitle(name string) *Embed {
	e.Title = name
	return e
}

func (e *Embed) SetTimestamp(timestamp string) *Embed {
	e.Timestamp = timestamp
	return e
}

func (e *Embed) SetDescription(description string) *Embed {
	if len(description) > EmbedLimitDescription {
		description = description[:EmbedLimitDescription]
	}
	e.Description = description
	return e
}

func (e *Embed) AddField(name, value string) *Embed {
	if len(value) > EmbedLimitFieldValue {
		value = value[:EmbedLimitFieldValue]
	}

	if len(name) > EmbedLimitFieldName {
		name = name[:EmbedLimitFieldName]
	}

	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:  name,
		Value: value,
	})

	return e
}

func (e *Embed) SetFields(fields []*discordgo.MessageEmbedField) *Embed {
	e.Fields = fields
	return e
}

func (e *Embed) SetFooter(args ...string) *Embed {
	iconURL := ""
	text := ""
	proxyURL := ""

	switch {
	case len(args) > 2:
		proxyURL = args[2]
		fallthrough
	case len(args) > 1:
		iconURL = args[1]
		fallthrough
	case len(args) > 0:
		text = args[0]
	case len(args) == 0:
		return e
	}

	e.Footer = &discordgo.MessageEmbedFooter{
		IconURL:      iconURL,
		Text:         text,
		ProxyIconURL: proxyURL,
	}

	return e
}

func (e *Embed) SetImage(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Image = &discordgo.MessageEmbedImage{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

func (e *Embed) SetThumbnail(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

func (e *Embed) SetAuthor(args ...string) *Embed {
	var (
		name     string
		iconURL  string
		URL      string
		proxyURL string
	)

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		name = args[0]
	}
	if len(args) > 1 {
		iconURL = args[1]
	}
	if len(args) > 2 {
		URL = args[2]
	}
	if len(args) > 3 {
		proxyURL = args[3]
	}

	e.Author = &discordgo.MessageEmbedAuthor{
		Name:         name,
		IconURL:      iconURL,
		URL:          URL,
		ProxyIconURL: proxyURL,
	}

	return e
}

func (e *Embed) SetURL(URL string) *Embed {
	e.URL = URL
	return e
}

func (e *Embed) SetColor(clr int) *Embed {
	e.Color = clr
	return e
}

func (e *Embed) InlineAllFields() *Embed {
	for _, v := range e.Fields {
		v.Inline = true
	}
	return e
}

func (e *Embed) Truncate() *Embed {
	e.TruncateDescription()
	e.TruncateFields()
	e.TruncateFooter()
	e.TruncateTitle()
	return e
}

func (e *Embed) TruncateFields() *Embed {
	if len(e.Fields) > 25 {
		e.Fields = e.Fields[:EmbedLimitField]
	}

	for _, v := range e.Fields {

		if len(v.Name) > EmbedLimitFieldName {
			v.Name = v.Name[:EmbedLimitFieldName]
		}

		if len(v.Value) > EmbedLimitFieldValue {
			v.Value = v.Value[:EmbedLimitFieldValue]
		}

	}
	return e
}

func (e *Embed) TruncateDescription() *Embed {
	if len(e.Description) > EmbedLimitDescription {
		e.Description = e.Description[:EmbedLimitDescription]
	}
	return e
}

func (e *Embed) TruncateTitle() *Embed {
	if len(e.Title) > EmbedLimitTitle {
		e.Title = e.Title[:EmbedLimitTitle]
	}
	return e
}

func (e *Embed) TruncateFooter() *Embed {
	if e.Footer != nil && len(e.Footer.Text) > EmbedLimitFooter {
		e.Footer.Text = e.Footer.Text[:EmbedLimitFooter]
	}
	return e
}

type MessageEmbedFieldBuilder struct {
	chunks []string
	index  int
	name   string
	mutex  sync.Mutex
}

func (b *MessageEmbedFieldBuilder) SetName(name string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.name = name
}

func (b *MessageEmbedFieldBuilder) IsEmpty() bool {
	return len(b.chunks) == 0
}

func (b *MessageEmbedFieldBuilder) Append(m string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for len(b.chunks) < b.index+1 {
		b.chunks = append(b.chunks, "")
	}

	if len(m)+len(b.chunks[b.index]) > EmbedLimitFieldValue {
		b.chunks = append(b.chunks, m)
		b.index++
		return
	}

	b.chunks[b.index] += m
}

func (b *MessageEmbedFieldBuilder) ToMessageEmbedFields() []*discordgo.MessageEmbedField {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	var fields []*discordgo.MessageEmbedField
	name := b.name
	if name == "" {
		name = "Field"
	}
	for i, chunk := range b.chunks {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  name + " " + strconv.Itoa(i+1),
			Value: chunk,
		})
	}
	return fields
}

func splitEmbedFields(e *Embed) [][]*discordgo.MessageEmbedField {
	fields := e.Fields
	baseNumberOfCharacters := len(e.Description) + len(e.Title)
	if e.Author != nil {
		baseNumberOfCharacters += len(e.Author.Name)
	}
	if e.Footer != nil {
		baseNumberOfCharacters += len(e.Footer.Text)
	}

	var splitFields [][]*discordgo.MessageEmbedField
	characters := baseNumberOfCharacters
	fromIndex := 0
	fieldsLen := len(fields)
	for index, field := range fields {
		fNameLen := len(field.Name)
		fValLen := len(field.Value)
		if characters+fNameLen+fValLen > EmbedSizeLimit || index == fieldsLen-1 {
			splitFields = append(splitFields, fields[fromIndex:index+1])
			fromIndex = index
			characters = baseNumberOfCharacters
		}
		characters += fNameLen
		characters += fValLen
	}

	return splitFields
}
