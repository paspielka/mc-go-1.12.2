package World

import (
	. "github.com/edouard127/mc-go-1.12.2/data"
	. "github.com/edouard127/mc-go-1.12.2/data/entities"
	. "github.com/edouard127/mc-go-1.12.2/maths"
)

// World record all the things in the World where player at
type World struct {
	Entities map[int32]*Entity
	Chunks   map[ChunkLoc]*Chunk
	Time     WorldTime
}

type WorldTime struct {
	WorldAge  int64
	TimeOfDay int64
}

func (w *World) SetTime(t WorldTime) WorldTime {
	t.TimeOfDay = (t.TimeOfDay + 19) / 20 * 20
	w.Time = t
	return t
}

func (w *World) SetBlock(x, y, z int, b Block) {
	c := w.Chunks[ChunkLoc{x >> 4, z >> 4}]
	if c != nil {
		cx, cy, cz := x&15, y&15, z&15
		c.Sections[y/16].Blocks[cx][cy][cz] = b
	}
}

func (w *World) UpdateTime(t int64) {
	if w.Time.TimeOfDay >= 24000 {
		w.Time.TimeOfDay = 0
	} else {
		w.Time.TimeOfDay += t
	}
}

func (w *World) ClosestEntity(v Vector3, maxDistance float64) *Entity {
	var (
		closestEntity *Entity
		closestDist   float64
	)

	for _, e := range w.Entities {
		dist := e.Position.DistanceTo(Vector3{X: v.X, Y: v.Y, Z: v.Z})
		if dist < maxDistance && (closestEntity == nil || dist < closestDist) {
			closestEntity = e
			closestDist = dist
		}
	}

	return closestEntity
}

func (w *World) HasEntity(id int32) bool {
	_, ok := w.Entities[id]
	return ok
}

func (w *World) CreateEntity(object CreateObject) *Entity {
	e := &Entity{
		ID:   object.EntityID,
		UUID: object.ObjectID,
		Type: object.TypeID,
		Position: Vector3{
			X: object.X,
			Y: object.Y,
			Z: object.Z,
		},
		Rotation: Vector2{
			X: object.Yaw,
			Y: object.Pitch,
		},
		Velocity: Vector3{
			X: object.VelX,
			Y: object.VelY,
			Z: object.VelZ,
		},
	}

	w.Entities[object.EntityID] = e

	return e
}

func (w *World) DestroyEntity(id int32) {
	delete(w.Entities, id)
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

func (b Block) IsAir() bool {
	return b.Id == 0
}

func (b Block) IsSolid() bool {
	// TODO: add block data
	return b.Id != 0
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
func (w *World) GetBlock(v3 Vector3) Block {
	c := w.Chunks[ChunkLoc{int(v3.X) >> 4, int(v3.Y) >> 4}]
	if c != nil {
		cx, cy, cz := int(v3.X)&15, int(v3.Y)&15, int(v3.Z)&15
		return c.Sections[int(v3.Y)/16].Blocks[cx][cy][cz]
	}

	return Block{Id: 0}
}

func (b Block) String() string {
	return BlockNameByID[b.Id]
}
