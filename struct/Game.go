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
	go func() {
		for {
			time.Sleep(time.Millisecond * 50)
			g.Events <- TickEvent{}
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
	//fmt.Printf("recv packet 0x%X\n", p.ID)
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
	case 0x0B: // Block Change
		err = HandleBlockChange(g, reader)
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
	case 0x27: // TODO: Handle error
		err = HandleEntityLookAndRelativeMove(g, reader)
	case 0x28:
		HandleEntityLook(g, reader)
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
	case 0x10:
		err = HandleMultiBlockChangePacket(g, reader)
		g.Events <- BlockChangeEvent{}
	case 0x1A:
		err = HandleDisconnect(g, reader)
	case 0x17:
		err = HandleSetSlotPacket(g, reader)
	case 0x49:
		err = HandleSoundEffect(g, reader)
	case 0x3E: // Entity velocity
		err = HandleEntityVelocity(g, reader)
	case 0x48: // Title
		err = HandleTitle(g, reader)
	case 0x36: // Entity Head Look
		err = HandleEntityHeadLook(g, reader)
	case 0x4C: // Entity Teleport
		err = HandleEntityTeleport(g, reader)
	case 0x32: // Destroy Entities
		err = HandleDestroyEntity(g, reader)
	case 0x47: // Time update
		err = HandleTimeUpdate(g, reader)
	case 0x3C: // Entity Metadata
		err = HandleEntityMetadata(g, reader)
	default:
		//fmt.Printf("unhandled packet 0x%X\n", p.ID)
	}
	return nil
}

func HandleEntityLook(g *Game, reader *bytes.Reader) {
	entityID, _ := pk.UnpackVarInt(reader)
	yaw, _ := pk.UnpackAngle(reader)
	pitch, _ := pk.UnpackAngle(reader)
	onGround, _ := pk.UnpackBoolean(reader)
	rotation := Vector2{
		X: yaw,
		Y: pitch,
	}

	e := g.World.Entities[entityID]
	if e != nil {
		e.SetRotation(rotation, onGround)
	}
}

func HandleBlockChange(g *Game, reader *bytes.Reader) error {
	position, _ := pk.UnpackPosition(reader)
	blockID, _ := pk.UnpackVarInt(reader)
	g.World.UpdateBlock(position.ToChunkPos(), position, blockID)
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
		X: yaw,
		Y: pitch,
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
	g.LookAt(v3)
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

func (g *Game) SetPosition(v3 Vector3) {
	g.GetPlayer().SetPosition(v3)
	SendPlayerPositionPacket(g) // Update the location to the server
}

func (g *Game) SetRotation(v2 Vector2) {
	g.GetPlayer().SetRotation(v2)
	SendPlayerLookPacket(g) // Update the rotation to the server
}

func (g *Game) SetPositionAndRotation(v3 Vector3, v2 Vector2) {
	g.SetPosition(v3)
	g.SetRotation(v2)
	SendPlayerPositionAndLookPacket(g) // Update the location and rotation to the server
}

func (g *Game) SetSpawnPosition(v3 Vector3) {
	g.Info.SetSpawnPosition(v3)
	g.SetPosition(v3)
}

func (g *Game) ClosestEntity(r float64) *Entity {
	return g.World.ClosestEntity(g.GetPlayer().Position, r)
}

func (g *Game) WalkTo(x, y, z float64) {
	g.Motion <- func() {
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

func (g *Game) WalkToVector(v3 Vector3) {
	go func() {
		dir := g.GetPlayer().GetFacing()
		dist := g.Player.GetPosition().DistanceTo(v3)
		// The walk speed is 0.2806 blocks per tick
		path := GeneratePathFromDirection(dir, int(dist), float32(0.2806))
		initPos := g.GetPlayer().GetPosition()
		for {
			select {
			case e := <-g.Events:
				switch e.(type) {
				case TickEvent:
					if len(path) == 0 {
						return
					}
					// TODO: Check if the block is walkable
					initPos = initPos.Add(path[0])
					g.SetPosition(initPos)
					path = path[1:]
				}
			}
		}
	}()
}

func (g *Game) WalkStraight(dist int) {
	go func() {
		dir := g.GetPlayer().GetFacing()
		// The walk speed is 0.2806 blocks per tick
		path := GeneratePathFromDirection(dir, dist, float32(0.2806))
		for {
			select {
			case e := <-g.Events:
				switch e.(type) {
				case TickEvent:
					if len(path) == 0 {
						return
					}
					// TODO: Check if the block is walkable
					g.SetPosition(g.GetPlayer().GetPosition().Add(path[0]))
					path = path[1:]
				}
			}
		}
	}()
}

func GeneratePathFromDirection(dir Direction, length int, speed float32) []Vector3 {
	var path []Vector3
	for i := 0; i < length; i++ {
		switch dir {
		case DNorth:
			path = append(path, Vector3{X: 0, Y: 0, Z: float64(-speed)})
		case DSouth:
			path = append(path, Vector3{X: 0, Y: 0, Z: float64(speed)})
		case DWest:
			path = append(path, Vector3{X: float64(-speed), Y: 0, Z: 0})
		case DEast:
			path = append(path, Vector3{X: float64(speed), Y: 0, Z: 0})
		}
	}
	return path
}

// LookAt method turn player's hand and make it look at a point.
func (g *Game) LookAt(v3 Vector3) {
	dx := v3.X - g.GetPlayer().Position.X
	dy := v3.Y - g.GetPlayer().Position.Y
	dz := v3.Z - g.GetPlayer().Position.Z
	r := math.Sqrt(dx*dx + dy*dy + dz*dz)
	yaw := -math.Atan2(dx, dz) / math.Pi * 180
	if yaw < 0 {
		yaw = 360 + yaw
	}
	pitch := -math.Asin(dy/r) / math.Pi * 180
	g.SetRotation(Vector2{X: float32(yaw), Y: float32(pitch)})
}

// LookYawPitch set player's hand to the direct by yaw and pitch.
// yaw can be [0, 360) and pitch can be (-180, 180).
// if |pitch|>90 the player's hand will be very strange.
func (g *Game) LookYawPitch(yaw, pitch float32, onGround bool) {
	rotation := Vector2{
		X: yaw,
		Y: pitch,
	}
	g.SetRotation(rotation)
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

func SendPlayerPositionAndLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackPosition(g.GetPlayer().Position)...)
	data = append(data, pk.PackRotation(g.GetPlayer().Rotation)...)
	data = append(data, pk.PackBoolean(g.GetPlayer().OnGround))
	g.SendChan <- pk.Packet{
		ID:   0x0E,
		Data: data,
	}
}
func SendPlayerLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackRotation(g.GetPlayer().Rotation)...)
	data = append(data, pk.PackBoolean(g.GetPlayer().OnGround))
	g.SendChan <- pk.Packet{
		ID:   0x0F,
		Data: data,
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
	e := g.World.Entities[entityID]
	if e != nil {
		fmt.Println("UpdateVelocity", entityID, velocity)
		e.SetPosition(e.Position.Add(velocity), true)
		SendPlayerPositionPacket(g)
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
// ***************************** //

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
	// TODO
	/*c, x, y, err := UnpackChunkDataPacket(p, g.Info.Dimension == 0)
	g.World.Chunks[ChunkLoc{X: x, Y: y}] = c*/
	return nil
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
	/*entityID, err := pk.UnpackVarInt(reader)
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
	}*/
	// TODO: Check if entity is player

	// To update player velocity, we need to a loop and wait 1 tick each time and update the velocity both on the client and server
	// If we don't wait for the server to send the packet to the client, the player will be teleported back to the original position or the player will rubberband

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
	cX, _ := pk.UnpackInt32(r)
	cY, _ := pk.UnpackInt32(r)
	v2 := Vector2{
		X: float32(cX),
		Y: float32(cY),
	}
	rCount, _ := pk.UnpackVarInt(r)
	for i := 0; i < int(rCount); i++ {
		relX, _ := r.ReadByte()
		relY, _ := r.ReadByte()
		relZ, _ := r.ReadByte()
		v3 := Vector3{
			X: float64(relX),
			Y: float64(relY),
			Z: float64(relZ),
		}
		blockID, _ := pk.UnpackVarInt(r)
		g.World.UpdateBlock(v2, v3, blockID)
	}
	// TODO
	/*c := g.World.Chunks[ChunkLoc{X: int(cX), Y: int(cY)}]
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
	}*/

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
	_, _ = r.ReadByte() // TODO: Flags
	TeleportID, _ := pk.UnpackVarInt(r)
	pos := Vector3{X: x, Y: y, Z: z}
	rot := Vector2{X: yaw, Y: pitch}
	g.SetPosition(pos)
	g.SetRotation(rot)
	SendTeleportConfirmPacket(g, TeleportID)
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
			X: float32(yaw) / 256 * 360,
			Y: float32(pitch) / 256 * 360,
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
	})
	yaw, err := pk.UnpackAngle(r)
	pitch, err := pk.UnpackAngle(r)
	np.SetRotation(Vector2{
		X: yaw,
		Y: pitch,
	})

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
		g.SetPositionAndRotation(g.Info.SpawnPosition, g.GetPlayer().Rotation)
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
			X: yaw0 + (yaw-yaw0)*float32(elapsed)/float32(t),
			Y: pitch0 + (pitch-pitch0)*float32(elapsed)/float32(t),
		})
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
		})
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func similar(a, b float64) bool {
	return a-b < 1 && b-a < 1
}

// TweenJump simulate player jump make no headway
func (g *Game) TweenJump() {
	go func() {
		p := g.GetPlayer()
		v3 := p.GetPosition()
		v3.Y = math.Floor(v3.Y) + 0.5
		p.SetPosition(v3)
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
			})
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// TweenJumpTo simulate player jump up a block
func (g *Game) TweenJumpTo(x, z int) {
	go func() {
		p := g.GetPlayer()
		v3 := p.GetPosition()
		v3.Y = math.Floor(v3.Y) + 0.5
		p.SetPosition(v3)
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
			})
			SendPlayerPositionPacket(g)
			time.Sleep(10 * time.Millisecond)
		}
		v3.X, v3.Z = float64(x)+0.5, float64(z)+0.5
		p.SetPosition(v3)
		SendPlayerPositionPacket(g)
	}()
}

func (g *Game) CalibratePos() {
	p := g.GetPlayer()
	v3 := p.GetBlockPos()
	v3under := p.GetBlockPosUnder()
	for NonSolid(g.GetBlock(v3under).String()) {
		v3under.Y--
		g.SetPosition(v3.Add(Vector3{X: 0.5, Y: 0.5, Z: 0.5}))
		SendPlayerPositionPacket(g)
		time.Sleep(time.Millisecond * 50)
	}
	g.Player.SetPosition(v3.Add(Vector3{X: 0.5, Y: 1, Z: 0.5}))
	SendPlayerPositionPacket(g)
}
