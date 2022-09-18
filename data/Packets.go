package data

const PacketHandshake = 0x00
const ProtocolVersion = 340

// Login Client bound
const (
	LoginDisconnect = iota
	LoginEncryptionRequest
	LoginSuccess
	LoginCompression
)

// Status Client bound
const (
	StatusResponse = iota
	StatusPongResponse
)

// Packets Client bound
const (
	SpawnObject = iota
	SpawnExperienceOrb
	SpawnGlobalEntity
	SpawnMob
	SpawnPainting
	SpawnPlayer
	Animation
	Statistics
	BlockBreakAnimation
	UpdateBlockEntity
	BlockAction
	BlockChange
	BossBar
	ServerDifficulty
	TabComplete
	ChatMessage
	MultiBlockChange
	ConfirmTransaction
	CloseWindow
	OpenWindow
	WindowItems
	WindowProperty
	SetSlot
	SetCooldown
	PluginMessage
	NamedSoundEffect
	Disconnect
	EntityStatus
	Explosion
	UnloadChunk
	ChangeGameState
	KeepAlive
	ChunkData
	Effect
	Particle
	JoinGame
	Map
	Entity
	EntityRelativeMove
	EntityLookAndRelativeMove
	EntityLook
	VehicleMove
	OpenSignEditor
	CraftRecipeResponse
	PlayerAbilities
	CombatEvent
	PlayerListItem
	PlayerPositionAndLook
	UseBed
	DestroyEntities
	RemoveEntityEffect
	ResourcePackSend
	Respawn
	EntityHeadLook
	SelectAdvancementTab
	WorldBorder
	Camera
	HeldItemChange
	DisplayScoreboard
	EntityMetadata
	AttachEntity
	EntityVelocity
	EntityEquipment
	SetExperience
	UpdateHealth
	ScoreboardObjective
	SetPassengers
	Teams
	UpdateScore
	SpawnPosition
	TimeUpdate
	Title
	SoundEffect
	PlayerListHeaderAndFooter
	CollectItem
	EntityTeleport
	Advancements
	EntityProperties
	EntityEffect
)
