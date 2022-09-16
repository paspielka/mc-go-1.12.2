package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	world2 "github.com/edouard127/mc-go-1.12.2/world"
)

func handleBlockChangePacket(g *Game, r *bytes.Reader) error {
	if !g.Settings.ReciveMap {
		return nil
	}
	x, y, z, err := pk.UnpackPosition(r)
	if err != nil {
		return err
	}
	c := g.World.Chunks[world2.ChunkLoc{X: x >> 4, Y: z >> 4}]
	if c != nil {
		id, err := pk.UnpackVarInt(r)
		if err != nil {
			return err
		}
		c.Sections[y/16].Blocks[x&15][y&15][z&15] = world2.Block{Id: uint(id)}
	}

	return nil
}
