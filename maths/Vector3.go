package maths

// Vector3 is a 3D vector
type Vector3 struct {
	X, Y, Z float64
}

func (v3 Vector3) Add(v Vector3) Vector3 {
	return Vector3{v3.X + v.X, v3.Y + v.Y, v3.Z + v.Z}
}

func (v3 Vector3) Sub(v Vector3) Vector3 {
	return Vector3{v3.X - v.X, v3.Y - v.Y, v3.Z - v.Z}
}

func (v3 Vector3) Mul(v Vector3) Vector3 {
	return Vector3{v3.X * v.X, v3.Y * v.Y, v3.Z * v.Z}
}

func (v3 Vector3) Div(v Vector3) Vector3 {
	return Vector3{v3.X / v.X, v3.Y / v.Y, v3.Z / v.Z}
}

func (v3 Vector3) DistanceTo(v Vector3) float64 {
	return v3.Sub(v).Length()
}

func (v3 Vector3) Length() float64 {
	return (v3.X * v3.X) + (v3.Y * v3.Y) + (v3.Z * v3.Z)
}

func (v3 Vector3) LengthSquared() float64 {
	return v3.Length() * v3.Length()
}

func (v3 Vector3) Normalize() Vector3 {
	return v3.Div(Vector3{v3.Length(), v3.Length(), v3.Length()})
}
