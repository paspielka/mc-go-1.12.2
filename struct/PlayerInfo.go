package _struct

// PlayerInfo content player info in server.
type PlayerInfo struct {
	EntityID         int
	Gamemode         int
	Hardcore         bool
	Dimension        int
	Difficulty       int
	LevelType        string
	ReducedDebugInfo bool
	SpawnPosition    Position
}
