package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleTitle(g *Game, reader *bytes.Reader) error {
	action, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	switch action {
	case 0:
		title, err := pk.UnpackString(reader)
		if err != nil {
			return err
		}
		g.Events <- TitleEvent{Action: action, Text: title}
	}
	return nil
}
