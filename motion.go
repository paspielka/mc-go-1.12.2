package gomcbot

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	"math"
	"time"
)

// SetPosition method move your character around.
// Server will ignore this if changes too much.
func (g *Game) SetPosition(x, y, z float64, onGround bool) {
	g.motion <- func() {
		g.player.X, g.player.Y, g.player.Z = x, y, z
		g.player.OnGround = onGround
		sendPlayerPositionPacket(g) // Update the location to the server
	}
}

// CalcCollisionBox calculate the collision box of the block
func CalcCollisionBox(b Block) []Vector3 {
	// TODO: calculate the collision box of the block
	return []Vector3{}
}

// PlaceBlock place a block in the position and wait
func (g *Game) PlaceBlock(x, y, z int, face Face) error {
	b := g.GetBlock(x, y, z).id
	sendPlayerBlockPlacementPacket(g, x, y, z, face, 0, 0, 0, 0)

	for {
		time.Sleep(time.Millisecond * 50)
		if g.GetBlock(x, y, z).id != b {
			break
		}
		g.SwingHand(true)
	}

	return nil
}

func sendPlayerBlockPlacementPacket(g *Game, x int, y int, z int, face Face, i int, i2 int, i3 int, i4 int) {
	var data []byte
	data = append(data, pk.PackPosition(x, y, z)...)
	data = append(data, pk.PackVarInt(int32(face))...)
	data = append(data, pk.PackVarInt(int32(i))...)
	data = append(data, pk.PackVarInt(int32(i2))...)
	data = append(data, pk.PackVarInt(int32(i3))...)
	data = append(data, pk.PackVarInt(int32(i4))...)
	g.sendChan <- pk.Packet{
		ID:   0x1F,
		Data: data,
	}
}

// LookAt method turn player's hand and make it look at a point.
func (g *Game) LookAt(x, y, z float64) {
	x0, y0, z0 := g.player.X, g.player.Y, g.player.Z
	x, y, z = x-x0, y-y0, z-z0

	r := math.Sqrt(x*x + y*y + z*z)
	yaw := -math.Atan2(x, z) / math.Pi * 180
	for yaw < 0 {
		yaw = 360 + yaw
	}
	pitch := -math.Asin(y/r) / math.Pi * 180

	g.LookYawPitch(float32(yaw), float32(pitch))
}

// LookYawPitch set player's hand to the direct by yaw and pitch.
// yaw can be [0, 360) and pitch can be (-180, 180).
// if |pitch|>90 the player's hand will be very strange.
func (g *Game) LookYawPitch(yaw, pitch float32) {
	g.motion <- func() {
		g.player.Yaw, g.player.Pitch = yaw, pitch
		sendPlayerLookPacket(g) // Update the orientation to the server
	}
}

// SwingHand sent when the player's arm swings.
// if hand is true, swing the main hand
func (g *Game) SwingHand(hand bool) {
	if hand {
		sendAnimationPacket(g, 0)
	} else {
		sendAnimationPacket(g, 1)
	}
}
func (g *Game) Attack(e LivingEntity) {
	sendUseEntityPacket(g, e.EntityID(), 1, e.Position)
}

func sendUseEntityPacket(g *Game, target int32, action int32, targetPos Vector3) {
	var data []byte
	data = append(data, pk.PackVarInt(target)...)
	data = append(data, pk.PackVarInt(action)...)
	data = append(data, pk.PackDouble(targetPos.X)...)
	data = append(data, pk.PackDouble(targetPos.Y)...)
	data = append(data, pk.PackDouble(targetPos.Z)...)
	g.sendChan <- pk.Packet{
		ID:   0x0A,
		Data: data,
	}
}

// Dig a block in the position and wait
func (g *Game) Dig(x, y, z int) error {
	b := g.GetBlock(x, y, z).id
	sendPlayerDiggingPacket(g, 0, x, y, z, Top) //start
	sendPlayerDiggingPacket(g, 2, x, y, z, Top) //end

	for {
		time.Sleep(time.Millisecond * 50)
		if g.GetBlock(x, y, z).id != b {
			break
		}
		g.SwingHand(true)
	}

	return nil
}

// UseItem use the item in hand.
// if hand is true, swing the main hand
func (g *Game) UseItem(hand bool) {
	if hand {
		sendUseItemPacket(g, 0)
	} else {
		sendUseItemPacket(g, 1)
	}
}
