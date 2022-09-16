package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleSpawnPositionPacket(g *Game, r *bytes.Reader) (err error) {
	g.Info.SpawnPosition.X, g.Info.SpawnPosition.Y, g.Info.SpawnPosition.Z, err =
		pk.UnpackPosition(r)
	return
}
