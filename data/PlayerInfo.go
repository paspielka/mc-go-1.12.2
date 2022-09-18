package data

import . "github.com/edouard127/mc-go-1.12.2/maths"

// PlayerInfo content player info in server.
type PlayerInfo struct {
	EntityID         int
	Gamemode         int
	Hardcore         bool
	Dimension        int
	Difficulty       int
	LevelType        string
	ReducedDebugInfo bool
	SpawnPosition    Vector3
}

func (p *PlayerInfo) SetSpawnPosition(v3 Vector3) {
	p.SpawnPosition = v3
}
