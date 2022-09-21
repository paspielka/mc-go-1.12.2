package entities

import (
	_struct "github.com/edouard127/mc-go-1.12.2/data"
	. "github.com/edouard127/mc-go-1.12.2/maths"
	"math"
)

// Player includes the player's status.
type Player struct {
	LivingEntity
	UUID [2]int64 //128bit UUID

	OnGround bool

	HeldItem       int
	Inventory      []_struct.Slot
	Food           int32
	FoodSaturation float32
}

// GetPosition return the player's position
func (p *Player) GetPosition() Vector3 {
	return p.Position
}

// GetBlockPos return the position of the Block at player's feet
func (p *Player) GetBlockPos() Vector3 {
	return Vector3{X: math.Floor(p.Position.X), Y: math.Floor(p.Position.Y), Z: math.Floor(p.Position.Z)}
}

// GetBlockPosUnder return the position of the Block under player's feet
func (p *Player) GetBlockPosUnder() Vector3 {
	return p.LivingEntity.GetBlockPosUnder()
}

// ShouldSendGround return true if the player is on ground
func (p *Player) ShouldSendGround() bool {
	return p.LivingEntity.ShouldSendGround()
}

func (p *Player) GetItemSlotByID(id int) _struct.Slot {
	for _, item := range p.Inventory {
		if item.ID == id {
			return item
		}
	}
	return _struct.Slot{}
}

func (p *Player) GetFirstStackSlot() int16 {
	for i, item := range p.Inventory {
		if item.Count == 64 {
			return int16(i)
		}
	}
	return -1
}

func (p *Player) GetItemAt(slot int16) _struct.Slot {
	return p.Inventory[slot]
}
