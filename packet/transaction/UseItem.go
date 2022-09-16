package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func SendUseItemPacket(g *Game, hand int32) {
	data := pk.PackVarInt(hand)
	g.SendChan <- pk.Packet{
		ID:   0x20,
		Data: data,
	}
}
