package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"github.com/edouard127/mc-go-1.12.2/world"
)

func SendPlayerBlockPlacementPacket(g *Game, x int, y int, z int, face world.Face, i int, i2 int, i3 int, i4 int) {
	var data []byte
	data = append(data, pk.PackPosition(x, y, z)...)
	data = append(data, pk.PackVarInt(int32(face))...)
	data = append(data, pk.PackVarInt(int32(i))...)
	data = append(data, pk.PackVarInt(int32(i2))...)
	data = append(data, pk.PackVarInt(int32(i3))...)
	data = append(data, pk.PackVarInt(int32(i4))...)
	g.SendChan <- pk.Packet{
		ID:   0x1F,
		Data: data,
	}
}
