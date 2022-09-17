package _struct

import (
	. "github.com/edouard127/mc-go-1.12.2/packet"
)

// Vector3 is a 3D vector
type Vector3 struct {
	X, Y, Z float64
}

func (v3 Vector3) Add(v Vector3) Vector3 {
	return Vector3{v3.X + v.X, v3.Y + v.Y, v3.Z + v.Z}
}

func PackVector3(v3 Vector3) (p []byte) {
	var data []byte
	data = append(data, PackDouble(v3.X)...)
	data = append(data, PackDouble(v3.Y)...)
	data = append(data, PackDouble(v3.Z)...)
	return data
}
