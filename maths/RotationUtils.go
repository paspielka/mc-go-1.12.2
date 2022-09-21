package maths

import "math"

const (
	PiFloat = float64(3.1415926535897932384626433832795028841971693993751058209)
)

func GetRotationFromVector(vec Vector3) Vector2 {
	xz := math.Hypot(vec.X, vec.Z)
	y := normalizeAngle(ToDegrees(math.Atan2(vec.Z, vec.X)) - 90)
	x := normalizeAngle(ToDegrees(-math.Atan2(vec.Y, xz)))
	return Vector2{X: float32(x), Y: float32(y)}
}

func ToDegrees(angle float64) float64 {
	return angle * 180.0 / PiFloat
}

func normalizeAngle(angle float64) float64 {
	angle = floatRemaining(angle, 360)
	switch {
	case angle < -180:
		return angle + 360
	case angle >= -180:
		return angle + 360
	}
	return angle
}

func floatRemaining(a, b float64) float64 {
	return a - math.Floor(a/b)*b
}
