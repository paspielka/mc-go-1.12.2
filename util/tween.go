package gomcbot

import (
	"fmt"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"math"
	"time"
)

// TweenLookAt is the Tween version of LookAt
func TweenLookAt(g *Game, x, y, z float64, t time.Duration) {
	p := g.GetPlayer()
	v3 := p.GetPosition()
	x, y, z = x-v3.X, y-v3.Y, z-v3.Z

	r := math.Sqrt(x*x + y*y + z*z)
	yaw := -math.Atan2(x, z) / math.Pi * 180
	for yaw < 0 {
		yaw = 360 + yaw
	}
	pitch := -math.Asin(y/r) / math.Pi * 180

	TweenLook(g, float32(yaw), float32(pitch), t)
}

// TweenLook do tween animation at player's head.
func TweenLook(g *Game, yaw, pitch float32, t time.Duration) {
	p := g.GetPlayer()
	start := time.Now()
	yaw0, pitch0 := p.Yaw, p.Pitch
	ofstY, ofstP := yaw-yaw0, pitch-pitch0
	var scale float32
	for scale < 1 {
		scale = float32(time.Since(start)) / float32(t)
		g.LookYawPitch(yaw0+ofstY*scale, pitch0+ofstP*scale)
		time.Sleep(time.Millisecond * 50)
	}
}

// TweenLineMove allows you smoothly move on plane. You can't move in Y axis
func TweenLineMove(g *Game, x, z float64) error {
	p := g.GetPlayer()
	start := time.Now()
	v3 := p.GetPosition()

	if similar(v3.X, x) && similar(v3.Z, z) {
		return nil
	}

	v3.Y = math.Floor(v3.Y) + 0.5
	ofstX, ofstZ := x-v3.X, z-v3.Z
	t := time.Duration(float64(time.Second) * (math.Sqrt(ofstX*ofstX+ofstZ*ofstZ) / 4.2))
	var scale float64
	for scale < 1 {
		scale = float64(time.Since(start)) / float64(t)
		g.SetPosition(Vector3{
			X: v3.X + ofstX*scale,
			Y: v3.Y,
			Z: v3.Z + ofstZ*scale,
		}, g.Player.OnGround)
		time.Sleep(time.Millisecond * 50)
	}

	p = g.GetPlayer()
	if !similar(p.X, x) || !similar(p.Z, z) {
		return fmt.Errorf("wrongly move")
	}
	return nil
}

func similar(a, b float64) bool {
	return a-b < 1 && b-a < 1
}

// TweenJump simulate player jump make no headway
func TweenJump(g *Game) {
	p := g.GetPlayer()
	y := math.Floor(p.Y)
	for tick := 0; tick < 11; tick++ {
		h := -1.7251e-8 + 0.4591*float64(tick) - 0.0417*float64(tick)*float64(tick)

		g.SetPosition(Vector3{
			X: p.X,
			Y: y + h,
			Z: p.Z,
		}, false)
		time.Sleep(time.Millisecond * 50)
	}
	g.SetPosition(Vector3{
		X: p.X,
		Y: y,
		Z: p.Z,
	}, true)
}

// TweenJumpTo simulate player jump up a block
func TweenJumpTo(g *Game, x, z int) {
	p := g.GetPlayer()
	y := math.Floor(p.Y)
	for tick := 0; tick < 7; tick++ {
		h := -1.7251e-8 + 0.4591*float64(tick) - 0.0417*float64(tick)*float64(tick)

		g.SetPosition(Vector3{
			X: p.X,
			Y: y + h,
			Z: p.Z,
		}, false)
		time.Sleep(time.Millisecond * 50)
	}
	err := TweenLineMove(g, float64(x)+0.5, float64(z)+0.5)
	if err != nil {
		return
	}
	CalibratePos(g)
}
