package _struct

import (
	"encoding/json"
	"fmt"
	"github.com/edouard127/mc-go-1.12.2/locales"
	"strings"
)

// ChatMsg is a message sent by other
type ChatMsg jsonChat

type jsonChat struct {
	Text string `json:"text"`

	Bold          bool   `json:"bold"`
	Italic        bool   `json:"Italic"`
	UnderLined    bool   `json:"underlined"`
	StrikeThrough bool   `json:"strikethrough"`
	Obfuscated    bool   `json:"obfuscated"`
	Color         string `json:"color"`

	Translate string            `json:"translate"`
	With      []json.RawMessage `json:"with"` // How can go handle an JSON array with Object and String?
	Extra     []jsonChat        `json:"extra"`
}

func NewChatMsg(jsonMsg []byte) (jc ChatMsg, err error) {
	if jsonMsg[0] == '"' {
		err = json.Unmarshal(jsonMsg, &jc.Text)
	} else {
		err = json.Unmarshal(jsonMsg, &jc)
	}
	return
}
func ExtractSenderName(msg string) string {
	if len(msg) > 0 {
		if msg[0] == '<' {
			// Remove minecraft color code
			msg = strings.Replace(msg, "§", "", 1)
			return msg[1 : strings.Index(msg, ">")-2]
		}
	}
	return ""
}
func ExtractContent(msg string) (string, string) {
	if len(msg) > 0 {
		if msg[0] == '<' {
			s := ExtractSenderName(msg)
			return s, msg[len(s):]
		}
	}
	return "", msg
}

var colors = map[string]int{
	"black":        30,
	"dark_blue":    34,
	"dark_green":   32,
	"dark_aqua":    36,
	"dark_red":     31,
	"dark_purple":  35,
	"gold":         33,
	"gray":         37,
	"dark_gray":    90,
	"blue":         94,
	"green":        92,
	"aqua":         96,
	"red":          91,
	"light_purple": 95,
	"yellow":       93,
	"white":        97,
}

// String return the message with escape sequence for ansi color.
// On Windows, you may want print this string using
// github.com/mattn/go-colorable.
func (c ChatMsg) String() (s string) {
	var format string
	if c.Bold {
		format += "1;"
	}
	if c.Italic {
		format += "3;"
	}
	if c.UnderLined {
		format += "4;"
	}
	if c.StrikeThrough {
		format += "9;"
	}
	if c.Color != "" {
		format += fmt.Sprintf("%d;", colors[c.Color])
	}

	if format != "" {
		s = "\033[" + format[:len(format)-1] + "m"
	}

	s += c.Text

	if format != "" {
		s += "\033[0m"
	}

	//handle translate
	if c.Translate != "" {
		args := make([]interface{}, len(c.With))
		for i, v := range c.With {
			args[i], _ = NewChatMsg(v) //ignore error
		}

		s += fmt.Sprintf(locales.EnUs[c.Translate], args...)
	}

	if c.Extra != nil {
		for i := range c.Extra {
			s += ChatMsg(c.Extra[i]).String()
		}
	}
	return
}
