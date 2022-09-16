package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func SendClientStatusPacket(g *Game, status int32) {
	data := pk.PackVarInt(status)
	g.SendChan <- pk.Packet{
		ID:   0x03,
		Data: data,
	}
}
