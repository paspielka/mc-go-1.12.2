package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func SendPlayerPositionAndLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackDouble(g.Player.X)...)
	data = append(data, pk.PackDouble(g.Player.Y)...)
	data = append(data, pk.PackDouble(g.Player.Z)...)
	data = append(data, pk.PackFloat(g.Player.Yaw)...)
	data = append(data, pk.PackFloat(g.Player.Pitch)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))
	//fmt.Printf("X:%f Y:%f Z:%f Yaw:%f Pitch:%f OnGround:%t\n", g.player.X, g.player.Y, g.player.Z, g.player.Yaw, g.player.Pitch, g.player.OnGround)

	g.SendChan <- pk.Packet{
		ID:   0x0E,
		Data: data,
	}
}
func SendPlayerLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackFloat(g.Player.Yaw)...)
	data = append(data, pk.PackFloat(g.Player.Pitch)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))
	//fmt.Printf("Yaw:%f Pitch:%f OnGround:%t\n", g.player.Yaw, g.player.Pitch, g.player.OnGround)
	g.SendChan <- pk.Packet{
		ID:   0x0F,
		Data: data,
	}
}

func SendPlayerPositionPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackDouble(g.Player.X)...)
	data = append(data, pk.PackDouble(g.Player.Y)...)
	data = append(data, pk.PackDouble(g.Player.Z)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))

	g.SendChan <- pk.Packet{
		ID:   0x10,
		Data: data,
	}
}
