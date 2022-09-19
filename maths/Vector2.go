package maths

import "fmt"

// Vector2 is a 2D vector
type Vector2 struct {
	X, Y float64
}

func (v2 Vector2) Add(v Vector2) Vector2 {
	return Vector2{v2.X + v.X, v2.Y + v.Y}
}

func (v2 Vector2) Sub(v Vector2) Vector2 {
	return Vector2{v2.X - v.X, v2.Y - v.Y}
}

func (v2 Vector2) Mul(v Vector2) Vector2 {
	return Vector2{v2.X * v.X, v2.Y * v.Y}
}

func (v2 Vector2) Div(v Vector2) Vector2 {
	return Vector2{v2.X / v.X, v2.Y / v.Y}
}

func (v2 Vector2) DistanceTo(v Vector2) float64 {
	return v2.Sub(v).Length()
}

func (v2 Vector2) Length() float64 {
	return (v2.X * v2.X) + (v2.Y * v2.Y)
}

func (v2 Vector2) LengthSquared() float64 {
	return v2.Length() * v2.Length()
}

func (v2 Vector2) Normalize() Vector2 {
	return v2.Div(Vector2{v2.Length(), v2.Length()})
}

func (v2 Vector2) String() string {
	return fmt.Sprintf("Vector2{X: %f, Y: %f}", v2.X, v2.Y)
}
