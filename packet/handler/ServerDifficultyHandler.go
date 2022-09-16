package handler

import (
	"bytes"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleServerDifficultyPacket(g *Game, r *bytes.Reader) error {
	diff, err := r.ReadByte()
	if err != nil {
		return err
	}
	g.Info.Difficulty = int(diff)
	return nil
}
