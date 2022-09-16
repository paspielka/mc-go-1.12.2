package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/packet/transaction"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleEntityVelocity(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	velX, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	velY, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	velZ, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	var velocity = Vector3{
		X: float64(velX) / 8000,
		Y: float64(velY) / 8000,
		Z: float64(velZ) / 8000,
	}
	g.Events <- EntityVelocityEvent{
		EntityID: entityID,
		Velocity: velocity,
	}
	UpdateVelocity(g, entityID, velocity)
	return nil
}
