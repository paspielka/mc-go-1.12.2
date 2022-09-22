package World

import (
	. "github.com/edouard127/mc-go-1.12.2/data"
	. "github.com/edouard127/mc-go-1.12.2/data/entities"
	. "github.com/edouard127/mc-go-1.12.2/maths"
)

// World record all the things in the World where player at
type World struct {
	Entities map[int32]*Entity
	Columns  map[ChunkPos]*Chunk
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

func (w *World) SetBlock(v3 Vector3, b Block) {
	// TODO
}

func (w *World) UpdateBlock(chunk Vector2, v3 Vector3, id int32) {
	// TODO
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
			X: float32(object.Yaw),
			Y: float32(object.Pitch),
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

// Block is the base of world
type Block struct {
	Id       uint
	Metadata uint
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
	// TODO
	return Block{}
}

func (b Block) String() string {
	return BlockNameByID[b.Id]
}
