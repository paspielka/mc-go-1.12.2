package entities

import (
	. "github.com/edouard127/mc-go-1.12.2/maths"
)

type Entity struct {
	ID       int32
	UUID     [2]int64
	Type     byte
	Position Vector3
	Rotation Vector2
	Velocity Vector3
	OnGround bool
}

func (e *Entity) SetPosition(v3 Vector3, onGround bool) {
	e.Position = v3
	e.OnGround = onGround
}

func (e *Entity) SetRotation(v Vector2, onGround bool) {
	e.Rotation = v
	e.OnGround = onGround
}

func (e *Entity) SetVelocity(v Vector3) {
	e.Velocity = v
}

func (e *Entity) EntityID() int32 {
	return e.ID
}

func (e *Entity) SetYaw(yaw float32) {
	e.Rotation.X = yaw
}

func (e *Entity) SetPitch(pitch float32) {
	e.Rotation.Y = pitch
}
