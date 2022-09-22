package entities

import (
	"github.com/edouard127/mc-go-1.12.2/data/World"
	. "github.com/edouard127/mc-go-1.12.2/maths"
	"math"
)

type LivingEntity struct {
	Entity
	Health   float32
	OnGround bool
}

func (p *LivingEntity) SetPosition(v3 Vector3) {
	p.Entity.SetPosition(v3, p.ShouldSendGround())
}

func (p *LivingEntity) SetRotation(v2 Vector2) {
	p.Entity.SetRotation(v2, p.ShouldSendGround())
}

// EntityID get entity ID.
func (p *LivingEntity) EntityID() int32 {
	return p.Entity.ID
}

// EntityUUID get entity UUID.
func (p *LivingEntity) EntityUUID() [2]int64 {
	return p.Entity.UUID
}

// GetBlockPosUnder return the position of the Block under player's feet
func (p *LivingEntity) GetBlockPosUnder() Vector3 {
	return Vector3{X: math.Floor(p.Position.X), Y: math.Floor(p.Position.Y) - 1, Z: math.Floor(p.Position.Z)}
}

// ShouldSendGround return true if the player is on ground
func (p *LivingEntity) ShouldSendGround() bool {
	b := p.GetBlockPosUnder()
	if p.Position.Y-b.Y < 0.1 {
		return true
	}
	return false
}

// GetFacing return the direction the player is facing
/*
	0: South
	1: West
	2: North
	3: East
*/
func (p *LivingEntity) GetFacing() World.Direction {
	yaw := p.Rotation.X
	switch {
	case yaw >= 45 && yaw < 135:
		return 1
	case yaw >= 135 && yaw < 225:
		return 2
	case yaw >= 225 && yaw < 315:
		return 3
	default:
		return 0
	}
}
