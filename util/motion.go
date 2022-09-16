package gomcbot

import (
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"time"
)

// CalibratePos moves player to the centre of block and fall player on the ground
func CalibratePos(g *Game) {
	p := g.GetPlayer()
	x, y, z := p.GetBlockPos()
	for NonSolid(g.GetBlock(x, y-1, z).String()) {
		y--
		g.Player.SetPosition(Vector3{float64(x) + 0.5, float64(y), float64(z) + 0.5})
		time.Sleep(time.Millisecond * 50)
	}
	g.Player.SetPosition(Vector3{X: float64(x) + 0.5, Y: float64(y), Z: float64(z) + 0.5})
}
