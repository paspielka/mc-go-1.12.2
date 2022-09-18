package entities

import . "github.com/edouard127/mc-go-1.12.2/maths"

type LivingEntity struct {
	Entity
	Health   float32
	OnGround bool
}

func (p *LivingEntity) SetPosition(v3 Vector3, onGround bool) {
	p.Entity.SetPosition(v3, onGround)
}

func (p *LivingEntity) SetRotation(v2 Vector2, onGround bool) {
	p.Entity.SetRotation(v2, onGround)
}

// EntityID get entity ID.
func (p *LivingEntity) EntityID() int32 {
	return p.Entity.ID
}

// EntityUUID get entity UUID.
func (p *LivingEntity) EntityUUID() [2]int64 {
	return p.Entity.UUID
}
