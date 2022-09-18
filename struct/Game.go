package _struct

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/edouard127/mc-go-1.12.2/data"
	. "github.com/edouard127/mc-go-1.12.2/data/World"
	. "github.com/edouard127/mc-go-1.12.2/data/entities"
	. "github.com/edouard127/mc-go-1.12.2/maths"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/util"
	"io"
	"math"
	"math/rand"
	"net"
	"time"
)

// Game is the Object used to access Minecraft server
type Game struct {
	addr string
	port int
	Conn net.Conn

	Receiver *bufio.Reader
	Sender   io.Writer

	threshold int
	Info      PlayerInfo
	Abilities PlayerAbilities
	Settings  Settings
	Player    Player
	World     World //the map data
	Server    Server

	SendChan chan pk.Packet  //be used when HandleGame
	recvChan chan *pk.Packet //be used when HandleGame
	Events   chan Event
	Motion   chan func() //used to submit a function and HandleGame do
}

// HandleGame receive server packet and response them correctly.
// Note that HandleGame will block if you don't receive from Events.
func (g *Game) HandleGame() error {
	defer func() {
		close(g.Events)
	}()

	errChan := make(chan error)

	g.SendChan = make(chan pk.Packet, 64)
	go func() {
		for p := range g.SendChan {
			err := g.SendPacket(&p)
			if err != nil {
				errChan <- fmt.Errorf("send packet in game fail: %v", err)
				return
			}
		}
	}()

	g.recvChan = make(chan *pk.Packet, 64)
	go func() {
		for {
			pack, err := g.recvPacket()
			if err != nil {
				close(g.recvChan)
				errChan <- fmt.Errorf("recv packet in game fail: %v", err)
				return
			}

			g.recvChan <- pack
		}
	}()
	for {
		select {
		case err := <-errChan:
			panic(fmt.Sprintf("error: %v\n", err))
		case pack, ok := <-g.recvChan:
			if !ok {
				panic(fmt.Sprintf("packet %v is not ok", pack))
			}
			err := HandlePack(g, pack)
			if err != nil {
				panic(fmt.Errorf("handle packet 0x%X error: %v", pack, err))
			}
		case f := <-g.Motion: // TODO: Fix memory block
			go f()
		}
	}
	return nil
}
func HandlePack(g *Game, p *pk.Packet) (err error) {
	fmt.Printf("recv packet 0x%X\n", p.ID)
	reader := bytes.NewReader(p.Data)

	switch p.ID {
	case 0x00: // Spawn Object
		HandleSpawnObject(g, reader)
	case 0x23:
		err = HandleJoinGamePacket(g, reader)
		g.Events <- JoinGameEvent{
			EntityID: g.Player.EntityID(),
		}
	case 0x25:

	/*case 0x18:
	HandlePluginPacket(g, reader)*/
	case 0x0D:
		err = HandleServerDifficultyPacket(g, reader)
	case 0x46:
		err = HandleSpawnPositionPacket(g, reader)
	case 0x2C:
		err = HandlePlayerAbilitiesPacket(g, reader)
		g.SendChan <- *g.Settings.Pack()
	case 0x3A:
		err = HandleHeldItemPacket(g, reader)
	case 0x20:
		err = HandleChunkDataPacket(g, p)
		g.Events <- BlockChangeEvent{}
	case 0x2F:
		err = HandlePlayerPositionAndLookPacket(g, reader)
	case 0x54:
		// handleDeclareRecipesPacket(g, reader)
	/*case 0x27: // TODO: Handle error
	err = HandleEntityLookAndRelativeMove(g, reader)*/
	/*case 0x28:
	HandleEntityHeadLookPacket(g, reader)*/
	case 0x1F:
		err = HandleKeepAlivePacket(g, reader)
	case 0x26:
		err = HandleEntityRelativeMove(g, reader)
	case 0x05:
		err = HandleSpawnPlayerPacket(g, reader)
	case 0x15:
		err = HandleWindowItemsPacket(g, reader)
	case 0x41:
		err = HandleUpdateHealthPacket(g, reader)
	case 0x0F:
		err = HandleChatMessagePacket(g, reader)
	case 0x0B:
		err = HandleBlockChangePacket(g, reader)
	case 0x10:
		err = HandleMultiBlockChangePacket(g, reader)
		g.Events <- BlockChangeEvent{}
	case 0x1A:
		err = HandleDisconnect(g, reader)
	case 0x17:
	// 	err = handleSetSlotPacket(g, reader)
	case 0x49:
		err = HandleSoundEffect(g, reader)
	case 0x3E: // Entity velocity
		err = HandleEntityVelocity(g, reader)
	case 0x48: // Title
		err = HandleTitle(g, reader)
	case 0x36: // Entity Head Look
		err = HandleEntityHeadLook(g, reader)
	/*case 0x4C: // Entity Teleport
	err = HandleEntityTeleport(g, reader)*/
	/*case 0x1A: // Entity Status
	err = HandleEntityStatus(g, reader)*/
	case 0x32: // Destroy Entities
		err = HandleDestroyEntity(g, reader)
	case 0x47: // Time update
		err = HandleTimeUpdate(g, reader)
	case 0x3C: // Entity Metadata
		err = HandleEntityMetadata(g, reader)
	default:
		fmt.Printf("unhandled packet 0x%X\n", p.ID)
	}
	return nil
}

func HandleSpawnObject(g *Game, reader *bytes.Reader) {
	object := CreateObject{}
	object.EntityID, _ = pk.UnpackVarInt(reader)
	object.ObjectID[0], _ = pk.UnpackInt64(reader)
	object.ObjectID[1], _ = pk.UnpackInt64(reader)
	object.TypeID, _ = pk.UnpackByte(reader)
	object.X, _ = pk.UnpackDouble(reader)
	object.Y, _ = pk.UnpackDouble(reader)
	object.Z, _ = pk.UnpackDouble(reader)
	object.Pitch, _ = pk.UnpackDouble(reader)
	object.Yaw, _ = pk.UnpackDouble(reader)
	object.Data, _ = pk.UnpackInt8(reader)
	object.VelX, _ = pk.UnpackDouble(reader)
	object.VelY, _ = pk.UnpackDouble(reader)
	object.VelZ, _ = pk.UnpackDouble(reader)
	g.World.CreateEntity(object)
}

func HandleDisconnect(g *Game, reader *bytes.Reader) error {
	reason, err := pk.UnpackString(reader)
	if err != nil {
		return err
	}
	g.Events <- DisconnectEvent{Text: reason}
	return nil
}

type EntityMetadataEvent struct {
	EntityID int32
	Metadata []pk.Metadata
}

func HandleEntityMetadata(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	metadata, err := pk.UnpackMetadata(reader)
	if err != nil {
		return err
	}
	for _, m := range metadata {
		switch m.Type {
		case 8: // Position
			fmt.Printf("Position: %v\n", m.Value)
		case 9: // OptPosition
			fmt.Printf("OptPosition: %v\n", m.Value)
		}
	}
	g.Events <- EntityMetadataEvent{
		EntityID: entityID,
		Metadata: metadata,
	}
	return nil
}

func HandleTimeUpdate(g *Game, reader *bytes.Reader) error {
	worldAge, _ := pk.UnpackInt64(reader)
	timeOfDay, _ := pk.UnpackInt64(reader)
	time := WorldTime{
		WorldAge:  worldAge,
		TimeOfDay: timeOfDay,
	}
	t := g.World.SetTime(time)
	g.Events <- TimeUpdateEvent{Time: t}
	return nil
}

func HandleDestroyEntity(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	g.World.DestroyEntity(entityID)
	return nil
}

func HandleEntityTeleport(g *Game, reader *bytes.Reader) error {
	entityID, _ := pk.UnpackVarInt(reader)
	x, _ := pk.UnpackDouble(reader)
	y, _ := pk.UnpackDouble(reader)
	z, _ := pk.UnpackDouble(reader)
	yaw, _ := pk.UnpackFloat(reader)
	pitch, _ := pk.UnpackFloat(reader)
	onGround, _ := pk.UnpackBoolean(reader)
	entity, ok := g.World.Entities[entityID]
	if !ok {
		return fmt.Errorf("entity not found")
	}
	position := Vector3{
		X: x,
		Y: y,
		Z: z,
	}
	rotation := Vector2{
		X: float64(yaw),
		Y: float64(pitch),
	}
	entity.SetPosition(position, onGround)
	entity.SetRotation(rotation, onGround)
	return nil
}

func (g *Game) recvPacket() (*pk.Packet, error) {
	return pk.RecvPacket(g.Receiver, g.threshold > 0)
}

// SendPacket send a packet to server
func (g *Game) SendPacket(p *pk.Packet) error {
	_, err := g.Sender.Write(p.Pack(g.threshold))
	return err
}

// Dig a block in the position and wait
func (g *Game) Dig(v3 Vector3) error {
	b := g.GetBlock(v3)
	if b.IsAir() {
		return fmt.Errorf("block is air")
	}
	g.Events <- DigStartEvent{Block: b}
	SendPlayerDiggingPacket(g, 0, v3, Top) //start
	time.Sleep(100 * time.Millisecond)
	g.Events <- DigStopEvent{Block: b}
	SendPlayerDiggingPacket(g, 2, v3, Top) //finish

	for {
		select {
		case e := <-g.Events:
			switch e.(type) {
			case TickEvent:
				if g.GetBlock(v3) != b {
					break
				}
				g.SwingHand(true)
			}
		}
	}
}

// PlaceBlock place a block in the position and wait
func (g *Game) PlaceBlock(v3 Vector3, face Face) error {
	b := g.GetBlock(v3).Id
	SendPlayerBlockPlacementPacket(g, v3, face, 0, 0, 0, 0)

	for {
		select {
		case e := <-g.Events:
			switch e.(type) {
			case TickEvent:
				if g.GetBlock(v3).Id != b {
					break
				}
				g.SwingHand(true)
			}
		}
	}
}

// Chat send chat message to server
// such as player send message in chat box
// msg can not longger than 256
func (g *Game) Chat(msg string) error {
	data := pk.PackString(msg)
	if len(data) > 256 {
		return fmt.Errorf("message too big")
	}

	pack := pk.Packet{
		ID:   0x02,
		Data: data,
	}

	g.SendChan <- pack
	return nil
}

// GetBlock return the block at (x, y, z)
func (g *Game) GetBlock(v3 Vector3) Block {
	bc := make(chan Block)

	g.Motion <- func() {
		bc <- g.World.GetBlock(v3)
	}

	return <-bc
}

// GetPlayer return the player
func (g *Game) GetPlayer() *Player {
	return &g.Player
}
func (g *Game) Attack(e *Entity) {
	g.LookAt(e.Position)
	g.SwingHand(true)
	SendUseEntityPacket(g, e.ID, 1, e.Position)
}

func (g *Game) Eat() {
	// TODO: Get food slot
	SendUseItemPacket(g, 1)
}

func (g *Game) SetPosition(v3 Vector3, onGround bool) {
	g.Player.SetPosition(v3, onGround)
	SendPlayerPositionPacket(g) // Update the location to the server
}

func (g *Game) SetSpawnPosition(v3 Vector3) {
	g.Info.SetSpawnPosition(v3)
	g.SetPosition(v3, true)
}

func (g *Game) ClosestEntity(r float64) *Entity {
	return g.World.ClosestEntity(g.GetPlayer().Position, r)
}

func (g *Game) WalkTo(x, y, z float64) {
	g.Motion <- func() {
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

func (g *Game) WalkStraight(dist float64) {
	g.Motion <- func() {
		g.Player.Position.X += dist * math.Sin(g.Player.Rotation.X)
		g.Player.Position.Z += dist * math.Cos(g.Player.Rotation.X)
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

// LookAt method turn player's hand and make it look at a point.
func (g *Game) LookAt(v3 Vector3) {
	g.Motion <- func() {
		dx := v3.X - g.Player.Position.X
		dy := v3.Y - g.Player.Position.Y
		dz := v3.Z - g.Player.Position.Z
		r := math.Sqrt(dx*dx + dy*dy + dz*dz)
		yaw := -math.Atan2(dx, dz) / math.Pi * 180
		if yaw < 0 {
			yaw = 360 + yaw
		}
		pitch := -math.Asin(dy/r) / math.Pi * 180
		g.Player.Rotation.X, g.Player.Rotation.Y = yaw, pitch
		g.Player.SetRotation(Vector2{X: yaw, Y: pitch}, true)

		SendPlayerLookPacket(g, float32(yaw), float32(pitch), true) // Update the location to the server
	}
}

// LookYawPitch set player's hand to the direct by yaw and pitch.
// yaw can be [0, 360) and pitch can be (-180, 180).
// if |pitch|>90 the player's hand will be very strange.
func (g *Game) LookYawPitch(yaw, pitch float32) {
	g.Player.Rotation.X, g.Player.Rotation.Y = float64(yaw), float64(pitch)
	SendPlayerLookPacket(g, yaw, pitch, true) // Update the orientation to the server
}

// SwingHand sent when the player's arm swings.
// if hand is true, swing the main hand
func (g *Game) SwingHand(hand bool) {
	if hand {
		SendAnimationPacket(g, 0)
	} else {
		SendAnimationPacket(g, 1)
	}
}

// SendAnimationPacket hand could be 0: main hand, 1: offhand
func SendAnimationPacket(g *Game, hand int32) {
	data := pk.PackVarInt(hand)
	g.SendChan <- pk.Packet{
		ID:   0x1D,
		Data: data,
	}
}

func SendClientStatusPacket(g *Game, status int32) {
	data := pk.PackVarInt(status)
	g.SendChan <- pk.Packet{
		ID:   0x03,
		Data: data,
	}
}

func SendKeepAlivePacket(g *Game, KeepAliveID int64) {
	g.SendChan <- pk.Packet{
		ID:   0x0B,
		Data: pk.PackUint64(uint64(KeepAliveID)),
	}
}

func SendPlayerBlockPlacementPacket(g *Game, v3 Vector3, face Face, i int, i2 int, i3 int, i4 int) {
	var data []byte
	data = append(data, pk.PackPosition(v3)...)
	data = append(data, pk.PackVarInt(int32(face))...)
	data = append(data, pk.PackVarInt(int32(i))...)
	data = append(data, pk.PackVarInt(int32(i2))...)
	data = append(data, pk.PackVarInt(int32(i3))...)
	data = append(data, pk.PackVarInt(int32(i4))...)
	g.SendChan <- pk.Packet{
		ID:   0x1F,
		Data: data,
	}
}

func SendPlayerDiggingPacket(g *Game, status int32, v3 Vector3, face Face) {
	data := pk.PackVarInt(status)
	data = append(data, pk.PackPosition(v3)...)
	data = append(data, byte(face))

	g.SendChan <- pk.Packet{
		ID:   0x14,
		Data: data,
	}
}

func SendPlayerPositionAndLookPacket(g *Game, v3 Vector3, v2 Vector2, onGround bool) {
	var data []byte
	data = append(data, pk.PackPosition(v3)...)
	data = append(data, pk.PackRotation(v2)...)
	data = append(data, pk.PackBoolean(onGround))
	g.SendChan <- pk.Packet{
		ID:   0x0E,
		Data: data,
	}
}
func SendPlayerLookPacket(g *Game, yaw, pitch float32, onGround bool) {
	g.Motion <- func() {
		var data []byte
		data = append(data, pk.PackFloat(yaw)...)
		data = append(data, pk.PackFloat(pitch)...)
		data = append(data, pk.PackBoolean(onGround))
		g.SendChan <- pk.Packet{
			ID:   0x0F,
			Data: data,
		}
	}
}

func SendPlayerPositionPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackPosition(g.Player.Position)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))

	g.SendChan <- pk.Packet{
		ID:   0x0D,
		Data: data,
	}
}

func SendTeleportConfirmPacket(g *Game, TeleportID int32) {
	g.SendChan <- pk.Packet{
		ID:   0x00,
		Data: pk.PackVarInt(TeleportID),
	}
}

func UpdateVelocity(g *Game, entityID int32, velocity Vector3) {
	fmt.Printf("UpdateVelocity: %v %v %v\n", velocity.X, velocity.Y, velocity.Z)
	e := g.World.Entities[entityID]
	if e != nil {
		e.SetPosition(e.Position.Add(velocity), true)
	}
}

func SendUseEntityPacket(g *Game, TargetEntityID int32, Type int32, Pos Vector3) {
	data := pk.PackVarInt(TargetEntityID)
	data = append(data, pk.PackVarInt(Type)...)
	if Type == 2 {
		data = append(data, pk.PackPosition(Pos)...)
	}
	g.SendChan <- pk.Packet{
		ID:   0x0A,
		Data: data,
	}
}

func SendUseItemPacket(g *Game, hand int32) {
	data := pk.PackVarInt(hand)
	g.SendChan <- pk.Packet{
		ID:   0x20,
		Data: data,
	}
}

// ********** Handler ********** //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
//								 //
// ***************************** //

func HandleBlockChangePacket(g *Game, r *bytes.Reader) error {
	if !g.Settings.ReciveMap {
		return nil
	}
	v3, err := pk.UnpackPosition(r)
	if err != nil {
		return err
	}
	c := g.World.Chunks[ChunkLoc{X: int(v3.X) >> 4, Y: int(v3.Z) >> 4}]
	if c != nil {
		id, err := pk.UnpackVarInt(r)
		if err != nil {
			return err
		}
		c.Sections[int(v3.Y)&15].Blocks[int(v3.X)&15][int(v3.Y)&15][int(v3.Z)&15] = Block{Id: uint(id)}
	}

	return nil
}

func HandleChatMessagePacket(g *Game, r *bytes.Reader) error {

	s, err := pk.UnpackString(r)
	if err != nil {
		return err
	}
	pos, err := r.ReadByte()
	if err != nil {
		return err
	}
	cm, err := NewChatMsg([]byte(s))
	if err != nil {
		return err
	}
	sender, content := ExtractContent(cm.String())
	raw := fmt.Sprintf("%s%s", sender, content)
	timestamp := time.Now().UnixMilli()
	g.Events <- ChatMessageEvent{Content: content, Sender: sender, RawString: RawString(raw), Timestamp: timestamp, Position: pos}
	return nil
}

func HandleChunkDataPacket(g *Game, p *pk.Packet) error {
	if !g.Settings.ReciveMap {
		return nil
	}

	c, x, y, err := UnpackChunkDataPacket(p, g.Info.Dimension == 0)
	g.World.Chunks[ChunkLoc{X: x, Y: y}] = c
	return err
}

func handleDeclareRecipesPacket(g *Game, r *bytes.Reader) {
	/*
		NumRecipes, index := pk.UnpackVarInt(r.D)
		for i := 0; i < int(NumRecipes); i++ {
			RecipeID, len := pk.UnpackString(p.Data[index:])
		 	index += len
		 	Type, len := pk.UnpackString(p.Data[index:])
		 	index += len
		 	switch Type {
		 	case "crafting_shapeless":
		 	}
		 }*/
}

func HandleEntityHeadLook(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	yaw, err := reader.ReadByte()
	if err != nil {
		return err
	}
	e := g.World.Entities[entityID]
	if e != nil {
		e.SetYaw(float32(yaw) * 360 / 256)
	}
	return nil
}

func HandleEntityVelocity(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	velX, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	velY, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	velZ, err := pk.UnpackInt16(reader)
	if err != nil {
		return err
	}
	var velocity = Vector3{
		X: float64(velX) / 8000,
		Y: float64(velY) / 8000,
		Z: float64(velZ) / 8000,
	}
	UpdateVelocity(g, entityID, velocity)
	g.Events <- EntityVelocityEvent{
		EntityID: entityID,
		Velocity: velocity,
	}
	return nil
}

func HandleEntityRelativeMove(g *Game, reader *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	entity := g.World.Entities[entityID]
	if entity != nil {
		previous := entity.Position
		deltaX, err := pk.UnpackInt16(reader)
		if err != nil {
			return err
		}
		deltaY, err := pk.UnpackInt16(reader)
		if err != nil {
			return err
		}
		deltaZ, err := pk.UnpackInt16(reader)
		if err != nil {
			return err
		}
		onGround, err := pk.UnpackBoolean(reader)
		var delta = Vector3{
			X: float64(deltaX) / 4096,
			Y: float64(deltaY) / 4096,
			Z: float64(deltaZ) / 4096,
		}
		entity.SetPosition(previous.Add(delta), onGround) // TODO: Fix this cause the position is 1.5 blocks different than the actual position
		g.Events <- EntityRelativeMoveEvent{
			EntityID: entityID,
			Delta:    delta,
		}
	}
	return nil
}

func HandleHeldItemPacket(g *Game, r *bytes.Reader) error {
	hi, err := r.ReadByte()
	if err != nil {
		return err
	}
	g.Player.HeldItem = int(hi)
	return nil
}

func HandleJoinGamePacket(g *Game, r *bytes.Reader) error {
	eid, err := pk.UnpackInt32(r)
	if err != nil {
		return fmt.Errorf("read EntityID fail: %v", err)
	}
	g.Info.EntityID = int(eid)
	gamemode, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("read gamemode fail: %v", err)
	}
	g.Info.Gamemode = int(gamemode & 0x7)
	g.Info.Hardcore = gamemode&0x8 != 0
	dimension, err := pk.UnpackInt32(r)
	if err != nil {
		return fmt.Errorf("read dimension fail: %v", err)
	}
	g.Info.Dimension = int(dimension)
	difficulty, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("read difficulty fail: %v", err)
	}
	g.Info.Difficulty = int(difficulty)
	// ignore Max Players
	_, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("read MaxPlayers fail: %v", err)
	}

	g.Info.LevelType, err = pk.UnpackString(r)
	if err != nil {
		return fmt.Errorf("read LevelType fail: %v", err)
	}
	rdi, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("read ReducedDebugInfo fail: %v", err)
	}
	g.Info.ReducedDebugInfo = rdi != 0x00
	return nil
}

func HandleKeepAlivePacket(g *Game, r *bytes.Reader) (err error) {
	KeepAliveID, err := pk.UnpackInt64(r)
	SendKeepAlivePacket(g, KeepAliveID)
	return nil
}

func HandleMultiBlockChangePacket(g *Game, r *bytes.Reader) error {
	if !g.Settings.ReciveMap {
		return nil
	}

	cX, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	cY, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}

	c := g.World.Chunks[ChunkLoc{X: int(cX), Y: int(cY)}]
	if c != nil {
		RecordCount, err := pk.UnpackVarInt(r)
		if err != nil {
			return err
		}

		for i := int32(0); i < RecordCount; i++ {
			xz, err := r.ReadByte()
			if err != nil {
				return err
			}
			y, err := r.ReadByte()
			if err != nil {
				return err
			}
			BlockID, err := pk.UnpackVarInt(r)
			if err != nil {
				return err
			}
			x, z := xz>>4, xz&0x0F

			c.Sections[y/16].Blocks[x][y%16][z] = Block{Id: uint(BlockID)}
		}
	}

	return nil
}

func HandlePlayerAbilitiesPacket(g *Game, r *bytes.Reader) error {
	f, err := r.ReadByte()
	if err != nil {
		return err
	}
	g.Abilities.Flags = int8(f)
	g.Abilities.FlyingSpeed, err = pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	g.Abilities.FieldOfViewModifier, err = pk.UnpackFloat(r)
	return err
}

func HandlePlayerPositionAndLookPacket(g *Game, r *bytes.Reader) error {
	x, _ := pk.UnpackDouble(r)
	y, _ := pk.UnpackDouble(r)
	z, _ := pk.UnpackDouble(r)
	yaw, _ := pk.UnpackFloat(r)
	pitch, _ := pk.UnpackFloat(r)
	TeleportID, _ := pk.UnpackVarInt(r)
	pos := Vector3{X: x, Y: y, Z: z}
	rot := Vector2{X: float64(yaw), Y: float64(pitch)}
	g.SetPosition(pos, true)
	SendTeleportConfirmPacket(g, TeleportID)
	SendPlayerPositionAndLookPacket(g, pos, rot, true)
	return nil
}

func HandleEntityLookAndRelativeMove(g *Game, r *bytes.Reader) error {
	ID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	e := g.World.Entities[ID]
	if e != nil {
		yaw, err := r.ReadByte()
		pitch, err := r.ReadByte()
		onGround, err := pk.UnpackBoolean(r)
		v2 := Vector2{
			X: float64(yaw) / 256 * 360,
			Y: float64(pitch) / 256 * 360,
		}
		e.SetRotation(v2, onGround)

		err = HandleEntityRelativeMove(g, r)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

/*func HandleEntityHeadLookPacket(g *Game, r *bytes.Reader) {
	ID, _ := pk.UnpackVarInt(r)
	yaw, _ := r.ReadByte()
	e := g.World.Entities[ID]
	if e != nil {

	}
}*/

func HandleSetSlotPacket(g *Game, r *bytes.Reader) error {
	windowID, err := r.ReadByte()
	if err != nil {
		return err
	}
	slot, err := pk.UnpackInt16(r)
	if err != nil {
		return err
	}
	slotData, err := pk.UnpackSlot(r)
	if err != nil {
		return err
	}

	switch int8(windowID) {
	case 0:
		if slot < 32 || slot > 44 {
			return fmt.Errorf("slot out of range")
		}
		fallthrough
	case -2:
		g.Player.Inventory[slot] = slotData
		g.Events <- InventoryChangeEvent(slot)
	}
	return nil
}

func HandleSoundEffect(g *Game, r *bytes.Reader) error {
	SoundID, _ := pk.UnpackVarInt(r)
	SoundCategory, _ := pk.UnpackVarInt(r)
	x, _ := pk.UnpackInt32(r)
	y, _ := pk.UnpackInt32(r)
	z, _ := pk.UnpackInt32(r)
	Volume, _ := pk.UnpackFloat(r)
	Pitch, _ := pk.UnpackFloat(r)
	g.Events <- SoundEffectEvent{Sound: SoundID, Category: SoundCategory, X: float64(x) / 8, Y: float64(y) / 8, Z: float64(z) / 8, Volume: Volume, Pitch: Pitch}

	return nil
}

func HandleSpawnPlayerPacket(g *Game, r *bytes.Reader) (err error) {
	np := new(Player)
	np.ID, err = pk.UnpackVarInt(r)
	np.UUID[0], err = pk.UnpackInt64(r)
	np.UUID[1], err = pk.UnpackInt64(r)
	x, err := pk.UnpackDouble(r)
	y, err := pk.UnpackDouble(r)
	z, err := pk.UnpackDouble(r)
	np.SetPosition(Vector3{
		X: x,
		Y: y,
		Z: z,
	}, true /* Assume the player is on ground */)
	yaw, err := pk.UnpackDouble(r)
	pitch, err := pk.UnpackDouble(r)
	np.SetRotation(Vector2{
		X: yaw,
		Y: pitch,
	}, true /* Assume the player is on ground */)

	g.World.Entities[np.ID] = &np.Entity // Add the player to the world entities
	return nil
}

func HandleSpawnPositionPacket(g *Game, r *bytes.Reader) error {
	v3, err := pk.UnpackPosition(r)
	if err != nil {
		return err
	}
	g.SetSpawnPosition(v3)
	SendPlayerPositionPacket(g)
	return nil
}

func HandleTitle(g *Game, reader *bytes.Reader) error {
	action, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	switch action {
	case 0:
		title, err := pk.UnpackString(reader)
		if err != nil {
			return err
		}
		g.Events <- TitleEvent{Action: action, Text: title}
	}
	return nil
}

func HandleUpdateHealthPacket(g *Game, r *bytes.Reader) (err error) {
	g.Player.Health, err = pk.UnpackFloat(r)
	g.Player.Food, err = pk.UnpackVarInt(r)
	g.Player.FoodSaturation, err = pk.UnpackFloat(r)
	if g.Player.Health < 1 { // Player is dead
		g.Events <- PlayerDeadEvent{} // Dead event
		SendPlayerPositionAndLookPacket(g, g.Info.SpawnPosition, g.Player.Rotation, true)
		time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
		SendClientStatusPacket(g, 0) // Status 0 means perform respawn
	}
	return
}

func HandleWindowItemsPacket(g *Game, r *bytes.Reader) (err error) {
	WindowID, err := r.ReadByte()
	Count, err := pk.UnpackInt16(r)
	slots := make([]Slot, Count)
	for i := int16(0); i < Count; i++ {
		slots[i], err = pk.UnpackSlot(r)
	}
	switch WindowID {
	case 0: //is player inventory
		g.Player.Inventory = slots
		g.Events <- InventoryChangeEvent(-2)
	}
	return nil
}

func HandleServerDifficultyPacket(g *Game, r *bytes.Reader) error {
	diff, err := r.ReadByte()
	if err != nil {
		return err
	}
	g.Info.Difficulty = int(diff)
	return nil
}

// TweenLookAt is the Tween version of LookAt
func TweenLookAt(g *Game, x, y, z float64, t time.Duration) {
	p := g.GetPlayer()
	v3 := p.GetPosition()
	x, y, z = x-v3.X, y-v3.Y, z-v3.Z

	r := math.Sqrt(x*x + y*y + z*z)
	yaw := -math.Atan2(x, z) / math.Pi * 180
	for yaw < 0 {
		yaw = 360 + yaw
	}
	pitch := -math.Asin(y/r) / math.Pi * 180

	TweenLook(g, float32(yaw), float32(pitch), t)
}

// TweenLook do tween animation at player's head.
func TweenLook(g *Game, yaw, pitch float32, t time.Duration) {
	p := g.GetPlayer()
	start := time.Now()
	yaw0, pitch0 := p.Rotation.X, p.Rotation.Y
	for {
		elapsed := time.Since(start)
		if elapsed > t {
			break
		}
		p.SetRotation(Vector2{
			X: float64(float32(yaw0) + (yaw-float32(yaw0))*float32(elapsed)/float32(t)),
			Y: float64(float32(pitch0) + (pitch-float32(pitch0))*float32(elapsed)/float32(t)),
		}, true)
		time.Sleep(10 * time.Millisecond)
	}
}

// TweenLineMove allows you smoothly move on plane. You can't move in Y axis
func TweenLineMove(g *Game, x, z float64) error {
	p := g.GetPlayer()
	start := time.Now()
	v3 := p.GetPosition()

	if similar(v3.X, x) && similar(v3.Z, z) {
		return nil
	}

	v3.Y = math.Floor(v3.Y) + 0.5
	ofstX, ofstZ := x-v3.X, z-v3.Z
	t := time.Duration(float64(time.Second) * (math.Sqrt(ofstX*ofstX+ofstZ*ofstZ) / 4.2))
	for {
		elapsed := time.Since(start)
		if elapsed > t {
			break
		}
		p.SetPosition(Vector3{
			X: v3.X + ofstX*float64(elapsed)/float64(t),
			Y: v3.Y,
			Z: v3.Z + ofstZ*float64(elapsed)/float64(t),
		}, true)
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func similar(a, b float64) bool {
	return a-b < 1 && b-a < 1
}

// TweenJump simulate player jump make no headway
func (g *Game) TweenJump() {
	p := g.GetPlayer()
	v3 := p.GetPosition()
	v3.Y = math.Floor(v3.Y) + 0.5
	p.SetPosition(v3, true)
	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed > 500*time.Millisecond {
			break
		}
		p.SetPosition(Vector3{
			X: v3.X,
			Y: v3.Y + 0.5*math.Sin(float64(elapsed)/float64(500*time.Millisecond)*math.Pi),
			Z: v3.Z,
		}, true)
		time.Sleep(10 * time.Millisecond)
	}
}

// TweenJumpTo simulate player jump up a block
func TweenJumpTo(g *Game, x, z int) {
	p := g.GetPlayer()
	v3 := p.GetPosition()
	v3.Y = math.Floor(v3.Y) + 0.5
	p.SetPosition(v3, true)
	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed > 500*time.Millisecond {
			break
		}
		p.SetPosition(Vector3{
			X: v3.X,
			Y: v3.Y + 0.5*math.Sin(float64(elapsed)/float64(500*time.Millisecond)*math.Pi),
			Z: v3.Z,
		}, true)
		time.Sleep(10 * time.Millisecond)
	}
	v3.X, v3.Z = float64(x)+0.5, float64(z)+0.5
	p.SetPosition(v3, true)
}

func CalibratePos(g *Game) {
	p := g.GetPlayer()
	v3 := p.GetBlockPos()
	v3under := p.GetBlockPosUnder()
	for NonSolid(g.GetBlock(v3under).String()) {
		v3under.Y--
		g.SetPosition(v3.Add(Vector3{X: 0.5, Y: 0.5, Z: 0.5}), true)
		time.Sleep(time.Millisecond * 50)
	}
	g.Player.SetPosition(v3.Add(Vector3{X: 0.5, Y: 1, Z: 0.5}), true)
}
