package _struct

// Vector3 is a 3D vector
type Vector3 struct {
	X, Y, Z float64
}

func (v3 Vector3) Add(v Vector3) Vector3 {
	return Vector3{v3.X + v.X, v3.Y + v.Y, v3.Z + v.Z}
}
