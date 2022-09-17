package _struct

// Event happens in game, and you can receive it from what Game.GetEvent() returns
type Event interface{}

/*
	Here is all events you will register.
	When event happens, match signal will be sending to Game.Event
*/
// JoinGameEvent sent when the client joins the game
type JoinGameEvent struct {
	EntityID int32
}

// DisconnectEvent sent when server disconnect this client. The value is the reason.
type DisconnectEvent ChatMsg

// PlayerSpawnEvent sent when this client is ready to playing.
type PlayerSpawnEvent struct{}

// PlayerDeadEvent sent when player is dead
type PlayerDeadEvent struct{}

// InventoryChangeEvent sent when player's inventory is changed.
// The value is the changed slot id.
// -1 means the cursor (the item currently dragged with the mouse)
// and -2 if the server update all the slots.
type InventoryChangeEvent int16

// BlockChangeEvent sent when a block has been broken or placed
type BlockChangeEvent struct{}

// ChatMessageEvent sent when chat message was received.
// When Pos is 0, this message should be displayed at chat box.
// When it's 1, this is a system message and also at chat box,
// if 2, it's a game info which displayed above hot bar.
type ChatMessageEvent struct {
	Content   string
	Sender    string
	Timestamp int64
	Position  byte
}

type TitleEvent struct {
	Action int32
	Text   string
}

// SoundEffectEvent sent when a sound should be played
// for sound id, check: https://pokechu22.github.io/Burger/1.13.2.html#sounds
// x, y, z is the position the sound played
// volume is the volume of the sound
// pitch is the direction of the sound
type SoundEffectEvent struct {
	Sound         int32
	Category      int32
	X, Y, Z       float64
	Volume, Pitch float32
}

type EntityVelocityEvent struct {
	EntityID int32
	Velocity Vector3
}

// GetEvents returns an int type channel.
// When event happens, an event ID will be sent into this chan
// Note that HandleGame will block if you don't receive from Events
func (g *Game) GetEvents() <-chan Event {
	return g.Events
}
