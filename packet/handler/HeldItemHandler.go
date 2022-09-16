package handler

import (
	"bytes"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleHeldItemPacket(g *Game, r *bytes.Reader) error {
	hi, err := r.ReadByte()
	if err != nil {
		return err
	}
	g.Player.HeldItem = int(hi)
	return nil
}
