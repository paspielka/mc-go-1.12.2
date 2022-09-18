package _struct

import (
	"bufio"
	"bytes"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/util"
	"io"
	"math"
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
			time.Sleep(50 * time.Millisecond)
			g.Events <- TickEvent{}
		}
	}()

	for {
		select {
		case err := <-errChan:
			close(g.SendChan)
			return err
		case pack, ok := <-g.recvChan:
			if !ok {
				break
			}
			err := HandlePack(g, pack)
			if err != nil {
				return fmt.Errorf("handle packet 0x%X error: %v", pack, err)
			}
		case motion := <-g.Motion:
			motion()
		}
	}
}
func HandlePack(g *Game, p *pk.Packet) (err error) {
	//fmt.Printf("recv packet 0x%X\n", p.ID)
	reader := bytes.NewReader(p.Data)

	switch p.ID {
	case 0x23:
		err = HandleJoinGamePacket(g, reader)
		g.Events <- JoinGameEvent{
			EntityID: g.Player.ID,
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
	case 0x27:
		err = HandleEntityLookAndRelativeMove(g, reader)
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
		// Should assume that the server has already closed the connection by the time the packet arrives.
		g.Events <- DisconnectEvent{Text: "disconnect"}
		err = fmt.Errorf("disconnect")
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
		//fmt.Printf("unhandled packet 0x%1X\n", p.ID)
	}
	return
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
	worldAge, err := pk.UnpackInt64(reader)
	if err != nil {
		return err
	}
	timeOfDay, err := pk.UnpackInt64(reader)
	if err != nil {
		return err
	}
	time := WorldTime{
		WorldAge:  worldAge,
		TimeOfDay: timeOfDay,
	}
	g.World.SetTime(time)
	g.Events <- TimeUpdateEvent{Time: time}
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
	entityID, err := pk.UnpackVarInt(reader)
	if err != nil {
		return err
	}
	x, err := pk.UnpackDouble(reader)
	if err != nil {
		return err
	}
	y, err := pk.UnpackDouble(reader)
	if err != nil {
		return err
	}
	z, err := pk.UnpackDouble(reader)
	if err != nil {
		return err
	}
	yaw, err := pk.UnpackFloat(reader)
	if err != nil {
		return err
	}
	pitch, err := pk.UnpackFloat(reader)
	if err != nil {
		return err
	}
	onGround, err := pk.UnpackBoolean(reader)
	if err != nil {
		return err
	}
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
	entity.SetRotation(rotation)
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
func (g *Game) Dig(x, y, z int) error {
	b := g.GetBlock(x, y, z).Id
	SendPlayerDiggingPacket(g, 0, x, y, z, Top) //start
	SendPlayerDiggingPacket(g, 2, x, y, z, Top) //end

	for {
		select {
		case e := <-g.Events:
			switch e.(type) {
			case TickEvent:
				if g.GetBlock(x, y, z).Id != b {
					break
				}
				g.SwingHand(true)
			}
		}
	}
}

// PlaceBlock place a block in the position and wait
func (g *Game) PlaceBlock(x, y, z int, face Face) error {
	b := g.GetBlock(x, y, z).Id
	SendPlayerBlockPlacementPacket(g, x, y, z, face, 0, 0, 0, 0)

	for {
		select {
		case e := <-g.Events:
			switch e.(type) {
			case TickEvent:
				if g.GetBlock(x, y, z).Id != b {
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
func (g *Game) GetBlock(x, y, z int) Block {
	bc := make(chan Block)

	g.Motion <- func() {
		bc <- g.World.GetBlock(x, y, z)
	}

	return <-bc
}

// GetPlayer return the player
func (g *Game) GetPlayer() *Player {
	return &g.Player
}
func (g *Game) Attack(e *LivingEntity) {
	g.LookAt(e.X, e.Y, e.Z)
	//SendUseEntityPacket(g, e.ID, 1, e.Position)
}
func (g *Game) SetPosition(v3 Vector3, onGround bool) {
	g.Motion <- func() {
		g.GetPlayer().Position = v3
		g.GetPlayer().X, g.GetPlayer().Y, g.GetPlayer().Z = v3.X, v3.Y, v3.Z
		g.GetPlayer().OnGround = onGround
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

func (g *Game) ClosestEntity(r float64) *LivingEntity {
	return g.World.ClosestEntity(g.GetPlayer().Position, r)
}

func (g *Game) WalkTo(x, y, z float64) {
	g.Motion <- func() {
		g.Player.X, g.Player.Y, g.Player.Z = x, y, z
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

func (g *Game) WalkStraight(dist float64) {
	g.Motion <- func() {
		g.Player.X += dist * math.Sin(float64(g.Player.Yaw))
		g.Player.Z += dist * math.Cos(float64(g.Player.Yaw))
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

// LookAt method turn player's hand and make it look at a point.
func (g *Game) LookAt(x, y, z float64) {
	g.Motion <- func() {
		dx := x - g.Player.X
		dy := y - g.Player.Y
		dz := z - g.Player.Z
		r := math.Sqrt(dx*dx + dy*dy + dz*dz)
		yaw := -math.Atan2(dx, dz) / math.Pi * 180
		if yaw < 0 {
			yaw = 360 + yaw
		}
		pitch := -math.Asin(dy/r) / math.Pi * 180
		g.GetPlayer().Yaw, g.GetPlayer().Pitch = float32(yaw), float32(pitch)
		g.GetPlayer().SetRotation(Vector2{
			X: float64(yaw),
			Y: float64(pitch),
		})

		SendPlayerLookPacket(g, float32(yaw), float32(pitch), true) // Update the location to the server
	}
}

// LookYawPitch set player's hand to the direct by yaw and pitch.
// yaw can be [0, 360) and pitch can be (-180, 180).
// if |pitch|>90 the player's hand will be very strange.
func (g *Game) LookYawPitch(yaw, pitch float32) {
	g.Motion <- func() {
		g.GetPlayer().Yaw, g.GetPlayer().Pitch = yaw, pitch
		SendPlayerLookPacket(g, yaw, pitch, true) // Update the orientation to the server
	}
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

func SendPlayerBlockPlacementPacket(g *Game, x int, y int, z int, face Face, i int, i2 int, i3 int, i4 int) {
	var data []byte
	data = append(data, pk.PackPosition(x, y, z)...)
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

func SendPlayerDiggingPacket(g *Game, status int32, x, y, z int, face Face) {
	data := pk.PackVarInt(status)
	data = append(data, pk.PackPosition(x, y, z)...)
	data = append(data, byte(face))

	g.SendChan <- pk.Packet{
		ID:   0x14,
		Data: data,
	}
}

func SendPlayerPositionAndLookPacket(g *Game, x, y, z float64, yaw, pitch float32, onGround bool) {
	var data []byte
	data = append(data, pk.PackDouble(x)...)
	data = append(data, pk.PackDouble(y)...)
	data = append(data, pk.PackDouble(z)...)
	data = append(data, pk.PackFloat(yaw)...)
	data = append(data, pk.PackFloat(pitch)...)
	data = append(data, pk.PackBoolean(onGround))

	g.SendChan <- pk.Packet{
		ID:   0x0E,
		Data: data,
	}
}
func SendPlayerLookPacket(g *Game, yaw, pitch float32, onGround bool) {
	var data []byte
	data = append(data, pk.PackFloat(yaw)...)
	data = append(data, pk.PackFloat(pitch)...)
	data = append(data, pk.PackBoolean(onGround))
	g.SendChan <- pk.Packet{
		ID:   0x0F,
		Data: data,
	}
}

func SendPlayerPositionPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackDouble(g.Player.X)...)
	data = append(data, pk.PackDouble(g.Player.Y)...)
	data = append(data, pk.PackDouble(g.Player.Z)...)
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
		e.SetPosition(e.Position.Add(velocity), true)
	}
}

func SendUseEntityPacket(g *Game, TargetEntityID int32, Type int32, Pos Vector3) {
	data := pk.PackVarInt(TargetEntityID)
	data = append(data, pk.PackVarInt(Type)...)
	if Type == 2 {
		data = append(data, PackVector3(Pos)...)
	}
	g.SendChan <- pk.Packet{
		ID:   0x02,
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
	x, y, z, err := pk.UnpackPosition(r)
	if err != nil {
		return err
	}
	c := g.World.Chunks[ChunkLoc{X: int(x) >> 4, Y: int(z) >> 4}]
	if c != nil {
		id, err := pk.UnpackVarInt(r)
		if err != nil {
			return err
		}
		c.Sections[int(y)/16].Blocks[int(x)&15][int(y)&15][int(z)&15] = Block{Id: uint(id)}
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
	var raw string
	if sender != "" {
		raw = fmt.Sprintf("<%s> %s", sender, content)
	} else {
		raw = content
	}
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

func HandleEntity(g *Game, r *bytes.Reader) error {
	entityID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	if !g.World.HasEntity(entityID) {
		g.World.CreateEntity(entityID)
	}
	return nil
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
		e.Yaw = float32(yaw)
	} else {
		g.World.CreateEntity(entityID)
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
			X: (float64(deltaX) / 4096),
			Y: (float64(deltaY) / 4096),
			Z: (float64(deltaZ) / 4096),
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
	g.Abilities.FieldofViewModifier, err = pk.UnpackFloat(r)
	return err
}

func HandlePlayerPositionAndLookPacket(g *Game, r *bytes.Reader) error {
	x, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	y, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	z, err := pk.UnpackDouble(r)
	if err != nil {
		return err
	}
	yaw, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	pitch, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}

	TeleportID, _ := pk.UnpackVarInt(r)
	SendTeleportConfirmPacket(g, TeleportID)
	SendPlayerPositionAndLookPacket(g, x, y, z, yaw, pitch, true)
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
		if err != nil {
			return err
		}

		pitch, err := r.ReadByte()
		if err != nil {
			return err
		}
		var rotation = Vector2{
			X: float64(yaw),
			Y: float64(pitch),
		}
		e.SetRotation(rotation)

		err = HandleEntityRelativeMove(g, r)
		if err != nil {
			return err
		}
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
	slotData, err := UnpackSlot(r)
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
	SoundID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	SoundCategory, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}

	x, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	y, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	z, err := pk.UnpackInt32(r)
	if err != nil {
		return err
	}
	Volume, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	Pitch, err := pk.UnpackFloat(r)
	if err != nil {
		return err
	}
	g.Events <- SoundEffectEvent{Sound: SoundID, Category: SoundCategory, X: float64(x) / 8, Y: float64(y) / 8, Z: float64(z) / 8, Volume: Volume, Pitch: Pitch}

	return nil
}

func HandleSpawnPlayerPacket(g *Game, r *bytes.Reader) (err error) {
	np := new(Player)
	np.ID, err = pk.UnpackVarInt(r)
	if err != nil {
		return
	}
	np.UUID[0], err = pk.UnpackInt64(r)
	if err != nil {
		return
	}
	np.UUID[1], err = pk.UnpackInt64(r)
	if err != nil {
		return
	}
	np.X, err = pk.UnpackDouble(r)
	if err != nil {
		return
	}
	np.Y, err = pk.UnpackDouble(r)
	if err != nil {
		return
	}
	np.Z, err = pk.UnpackDouble(r)
	if err != nil {
		return
	}

	yaw, err := r.ReadByte()
	if err != nil {
		return err
	}

	pitch, err := r.ReadByte()
	if err != nil {
		return err
	}

	np.Yaw = float32(yaw) * (1.0 / 256)
	np.Pitch = float32(pitch) * (1.0 / 256)

	g.World.Entities[np.ID] = &np.LivingEntity // Add the player to the world entities
	g.World.Entities[np.ID].SetPosition(Vector3{np.X, np.Y, np.Z}, true)
	return nil
}

func HandleSpawnPositionPacket(g *Game, r *bytes.Reader) error {
	x, y, z, err := pk.UnpackPosition(r)
	if err != nil {
		return err
	}
	g.Info.SpawnPosition = Vector3{float64(x), float64(y), float64(z)}
	g.GetPlayer().SetPosition(g.Info.SpawnPosition, true)
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
	if err != nil {
		return
	}
	g.Player.Food, err = pk.UnpackVarInt(r)
	if err != nil {
		return
	}
	g.Player.FoodSaturation, err = pk.UnpackFloat(r)
	if err != nil {
		return
	}

	if g.Player.Health < 1 { // Player is dead
		g.Events <- PlayerDeadEvent{} // Dead event
		SendPlayerPositionAndLookPacket(g, g.Player.X, g.Player.Y, g.Player.Z, g.Player.Yaw, g.Player.Pitch, true)
		time.Sleep(time.Second * 2)  // Wait for 2 sec make it more like a human
		SendClientStatusPacket(g, 0) // Status 0 means perform respawn
	}
	return
}

func HandleWindowItemsPacket(g *Game, r *bytes.Reader) (err error) {
	WindowID, err := r.ReadByte()
	if err != nil {
		return
	}

	Count, err := pk.UnpackInt16(r)
	if err != nil {
		return
	}

	slots := make([]Slot, Count)
	for i := int16(0); i < Count; i++ {
		slots[i], err = UnpackSlot(r)
		if err != nil {
			return
		}
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
	yaw0, pitch0 := p.Yaw, p.Pitch
	ofstY, ofstP := yaw-yaw0, pitch-pitch0
	var scale float32
	for scale < 1 {
		scale = float32(time.Since(start)) / float32(t)
		g.LookYawPitch(yaw0+ofstY*scale, pitch0+ofstP*scale)
		time.Sleep(time.Millisecond * 50)
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
	var scale float64
	for scale < 1 {
		scale = float64(time.Since(start)) / float64(t)
		g.SetPosition(Vector3{
			X: v3.X + ofstX*scale,
			Y: v3.Y,
			Z: v3.Z + ofstZ*scale,
		}, g.Player.OnGround)
		time.Sleep(time.Millisecond * 50)
	}

	p = g.GetPlayer()
	if !similar(p.X, x) || !similar(p.Z, z) {
		return fmt.Errorf("wrongly move")
	}
	return nil
}

func similar(a, b float64) bool {
	return a-b < 1 && b-a < 1
}

// TweenJump simulate player jump make no headway
func TweenJump(g *Game) {
	p := g.GetPlayer()
	y := math.Floor(p.Y)
	for tick := 0; tick < 11; tick++ {
		h := -1.7251e-8 + 0.4591*float64(tick) - 0.0417*float64(tick)*float64(tick)

		g.SetPosition(Vector3{
			X: p.X,
			Y: y + h,
			Z: p.Z,
		}, false)
		time.Sleep(time.Millisecond * 50)
	}
	g.SetPosition(Vector3{
		X: p.X,
		Y: y,
		Z: p.Z,
	}, true)
}

// TweenJumpTo simulate player jump up a block
func TweenJumpTo(g *Game, x, z int) {
	p := g.GetPlayer()
	y := math.Floor(p.Y)
	for tick := 0; tick < 7; tick++ {
		h := -1.7251e-8 + 0.4591*float64(tick) - 0.0417*float64(tick)*float64(tick)

		g.SetPosition(Vector3{
			X: p.X,
			Y: y + h,
			Z: p.Z,
		}, false)
		time.Sleep(time.Millisecond * 50)
	}
	err := TweenLineMove(g, float64(x)+0.5, float64(z)+0.5)
	if err != nil {
		return
	}
	CalibratePos(g)
}

func CalibratePos(g *Game) {
	p := g.GetPlayer()
	x, y, z := p.GetBlockPos()
	for NonSolid(g.GetBlock(x, y-1, z).String()) {
		y--
		g.SetPosition(Vector3{X: float64(x) + 0.5, Y: float64(y), Z: float64(z) + 0.5}, true) // TODO: Find a better way to do this
		time.Sleep(time.Millisecond * 50)
	}
	g.Player.SetPosition(Vector3{X: float64(x) + 0.5, Y: float64(y), Z: float64(z) + 0.5}, true)
}
