package _struct

import "math"

// Player includes the player's status.
type Player struct {
	LivingEntity
	UUID [2]int64 //128bit UUID

	OnGround bool

	HeldItem  int
	Inventory []Slot

	FoodSaturation float32
}

// GetPosition return the player's position
func (p *Player) GetPosition() Vector3 {
	return p.Position
}

// GetBlockPos return the position of the Block at player's feet
func (p *Player) GetBlockPos() (x, y, z int) {
	return int(math.Floor(p.X)), int(math.Floor(p.Y)), int(math.Floor(p.Z))
}

func (p *Player) GetItemSlotByID(id int) Slot {
	for _, item := range p.Inventory {
		if item.ID == id {
			return item
		}
	}
	return Slot{}
}

func (p *Player) GetItemSlot(slot int16) Slot {
	return p.Inventory[slot]
}

func (p *Player) GetItemAt(slot int16) Slot {
	return p.Inventory[slot]
}
