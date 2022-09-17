package _struct

type LivingEntity struct {
	ID         int32
	X, Y, Z    float64
	Yaw, Pitch float32
	Position   Vector3
	Rotation   Vector2
	Health     float32
	Food       int32
	OnGround   bool
}

func (p *LivingEntity) SetPosition(v3 Vector3, onGround bool) {
	p.X, p.Y, p.Z = v3.X, v3.Y, v3.Z
	p.Position = v3
	p.OnGround = onGround
}

func (p *LivingEntity) SetRotation(v Vector2) {
	p.Yaw, p.Pitch = float32(v.X), float32(v.Y)
	p.Rotation = v
}

// EntityID get player's entity ID.
func (p *LivingEntity) EntityID() int32 {
	return p.ID
}
