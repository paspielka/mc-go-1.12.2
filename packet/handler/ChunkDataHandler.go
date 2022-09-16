package handler

import (
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	world2 "github.com/edouard127/mc-go-1.12.2/world"
)

func handleChunkDataPacket(g *Game, p *pk.Packet) error {
	if !g.Settings.ReciveMap {
		return nil
	}

	c, x, y, err := world2.UnpackChunkDataPacket(p, g.Info.Dimension == 0)
	g.World.Chunks[world2.ChunkLoc{X: x, Y: y}] = c
	return err
}
