package world

import (
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

// World record all the things in the World where player at
type World struct {
	Entities map[int32]*LivingEntity
	Chunks   map[ChunkLoc]*Chunk
}

// Chunk store a 256*16*16 column blocks
type Chunk struct {
	Sections [16]Section
}

// Section store a 16*16*16 cube blocks
type Section struct {
	Blocks [16][16][16]Block
}

// Block is the base of world
type Block struct {
	Id uint
}

type ChunkLoc struct {
	X, Y int
}

// Face is a face of a block
type Face byte

// All six faces in a block
const (
	Bottom Face = iota
	Top
	North
	South
	West
	East
)

// GetBlock return the block in the position (x, y, z)
func (w *World) GetBlock(x, y, z int) Block {
	c := w.Chunks[ChunkLoc{x >> 4, z >> 4}]
	if c != nil {
		cx, cy, cz := x&15, y&15, z&15
		/*
			n = n&(16-1)

			is equal to

			n %= 16
			if n < 0 { n += 16 }
		*/

		return c.Sections[y/16].Blocks[cx][cy][cz]
	}

	return Block{Id: 0}
}

func (b Block) String() string {
	return blockNameByID[b.Id]
}
