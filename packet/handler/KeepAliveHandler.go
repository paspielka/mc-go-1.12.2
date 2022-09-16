package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/packet/transaction"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleKeepAlivePacket(g *Game, r *bytes.Reader) (err error) {
	KeepAliveID, err := pk.UnpackInt64(r)
	SendKeepAlivePacket(g, KeepAliveID)
	return nil
}
