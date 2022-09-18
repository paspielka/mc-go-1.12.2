package _struct

// World record all the things in the World where player at
type World struct {
	Entities map[int32]*LivingEntity
	Chunks   map[ChunkLoc]*Chunk
	Time     WorldTime
}

type WorldTime struct {
	WorldAge  int64
	TimeOfDay int64
}

func (w *World) SetTime(t WorldTime) {
	w.Time = t
}

func (w *World) SetBlock(x, y, z int, b Block) {
	c := w.Chunks[ChunkLoc{x >> 4, z >> 4}]
	if c != nil {
		cx, cy, cz := x&15, y&15, z&15
		c.Sections[y/16].Blocks[cx][cy][cz] = b
	}
}

func (w *World) UpdateTime() {
	if w.Time.TimeOfDay == 24000 {
		w.Time.WorldAge += w.Time.TimeOfDay
		w.Time.TimeOfDay = 0
	} else {
		w.Time.TimeOfDay += 20
	}
}

func (w *World) ClosestEntity(v Vector3, maxDistance float64) *LivingEntity {
	var (
		closestEntity *LivingEntity
		closestDist   float64
	)

	for _, e := range w.Entities {
		dist := e.Position.DistanceTo(Vector3{v.X, v.Y, v.Z})
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

func (w *World) CreateEntity(id int32) *LivingEntity {
	e := &LivingEntity{
		ID: id,
	}
	w.Entities[id] = e
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
	return BlockNameByID[b.Id]
}
