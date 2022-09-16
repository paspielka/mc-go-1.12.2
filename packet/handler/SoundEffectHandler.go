package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleSoundEffect(g *Game, r *bytes.Reader) error {
	SoundID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	SoundCategory, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}

	x, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	y, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	z, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	Volume, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	Pitch, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	g.Events <- SoundEffectEvent{Sound: SoundID, Category: SoundCategory, X: float64(x) / 8, Y: float64(y) / 8, Z: float64(z) / 8, Volume: Volume, Pitch: Pitch}

	return nil
}
