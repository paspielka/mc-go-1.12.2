package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	world2 "github.com/edouard127/mc-go-1.12.2/world"
)

func SendPlayerDiggingPacket(g *Game, status int32, x, y, z int, face world2.Face) {
	data := pk.PackVarInt(status)
	data = append(data, pk.PackPosition(x, y, z)...)
	data = append(data, byte(face))

	g.SendChan <- pk.Packet{
		ID:   0x14,
		Data: data,
	}
}
