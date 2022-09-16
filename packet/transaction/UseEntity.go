package transaction

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func SendUseEntityPacket(g *Game, TargetEntityID int32, Type int32, Pos Vector3) {
	data := pk.PackVarInt(TargetEntityID)
	data = append(data, pk.PackVarInt(Type)...)
	if Type == 2 {
		data = append(data, pk.PackVector3(Pos)...)
	}
	g.SendChan <- pk.Packet{
		ID:   0x02,
		Data: data,
	}
}
