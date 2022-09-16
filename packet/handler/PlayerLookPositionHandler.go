package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/packet/transaction"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handlePlayerPositionAndLookPacket(g *Game, r *bytes.Reader) error {
	x, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	y, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	z, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	yaw, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	pitch, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}

	flags, err := r.ReadByte()
	if err != nil {
		return err
	}

	if flags&0x01 == 0 {
		g.Player.X = x
	} else {
		g.Player.X += x
	}
	if flags&0x02 == 0 {
		g.Player.Y = y
	} else {
		g.Player.Y += y
	}
	if flags&0x04 == 0 {
		g.Player.Z = z
	} else {
		g.Player.Z += z
	}
	if flags&0x08 == 0 {
		g.Player.Yaw = yaw
	} else {
		g.Player.Yaw += yaw
	}
	if flags&0x10 == 0 {
		g.Player.Pitch = pitch
	} else {
		g.Player.Pitch += pitch
	}
	//confirm this packet with Teleport Confirm
	TeleportID, _ := pk.UnpackVarInt(r)
	SendTeleportConfirmPacket(g, TeleportID)
	SendPlayerPositionAndLookPacket(g)
	return nil
}

func handleEntityLookAndRelativeMove(g *Game, r *bytes.Reader) error {
	ID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	E := g.World.Entities[ID]
	if E != nil {
		DeltaX, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaY, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaZ, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}

		yaw, err := r.ReadByte()
		if err != nil {
			return err
		}

		pitch, err := r.ReadByte()
		if err != nil {
			return err
		}
		var rotation = Vector2{
			X: float64(yaw) * 360 / 256,
			Y: float64(pitch) * 360 / 256,
		}
		E.SetRotation(rotation)

		og, err := r.ReadByte()
		if err != nil {
			return err
		}
		E.OnGround = og != 0x00
		E.SetPosition(Vector3{
			X: E.Position.X + float64(DeltaX)/4096,
			Y: E.Position.Y + float64(DeltaY)/4096,
			Z: E.Position.Z + float64(DeltaZ)/4096,
		})
	}
	return nil
}

func handleEntityHeadLookPacket(g *Game, r *bytes.Reader) {
	ID, _ := pk.UnpackVarInt(r)
	E := g.World.Entities[ID]
	if E != nil {
		yaw, _ := r.ReadByte()
		pitch, _ := r.ReadByte()
		E.SetRotation(Vector2{
			X: float64(yaw) * 360 / 256,
			Y: float64(pitch) * 360 / 256,
		})
	}
}

func handleEntityRelativeMovePacket(g *Game, r *bytes.Reader) error {
	ID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	E := g.World.Entities[ID]
	if E != nil {
		DeltaX, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaY, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaZ, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}

		og, err := r.ReadByte()
		if err != nil {
			return err
		}
		E.OnGround = og != 0x00
		E.SetPosition(Vector3{
			X: E.Position.X + float64(DeltaX)/4096,
			Y: E.Position.Y + float64(DeltaY)/4096,
			Z: E.Position.Z + float64(DeltaZ)/4096,
		})
	}
	return nil
}
