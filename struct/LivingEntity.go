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

func (p *LivingEntity) SetPosition(v3 Vector3) {
	p.Position = v3
}

func (p *LivingEntity) SetRotation(v Vector2) {
	p.Yaw = float32(v.X)
	p.Pitch = float32(v.Y)
}

// EntityID get player's entity ID.
func (p *LivingEntity) EntityID() int32 {
	return p.ID
}
