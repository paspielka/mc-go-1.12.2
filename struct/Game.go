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
	case 0x18:
		HandlePluginPacket(g, reader)
	case 0x0D:
		err = HandleServerDifficultyPacket(g, reader)
	case 0x46:
		err = HandleSpawnPositionPacket(g, reader)
	case 0x2C:
		err = HandlePlayerAbilitiesPacket(g, reader)
		g.SendChan <- *g.Settings.Pack()
	case 0x3A:
		err = HandleHeldItemPacket(g, reader)
	/*case 0x20:
	err = handleChunkDataPacket(g, p)
	g.events <- BlockChangeEvent{}*/
	case 0x2F:
		err = HandlePlayerPositionAndLookPacket(g, reader)
	case 0x54:
		// handleDeclareRecipesPacket(g, reader)
	case 0x29:
		// err = handleEntityLookAndRelativeMove(g, reader)
	case 0x28:
		HandleEntityHeadLookPacket(g, reader)
	case 0x1F:
		err = HandleKeepAlivePacket(g, reader)
	/*case 0x26:
	handleEntityPacket(g, reader)*/
	case 0x05:
		err = HandleSpawnPlayerPacket(g, reader)
	case 0x15:
		err = HandleWindowItemsPacket(g, reader)
	case 0x44:
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
	case 0x4D:
		err = HandleSoundEffect(g, reader)
	case 0x3E: // Entity velocity
		err = HandleEntityVelocity(g, reader)
	case 0x48: // Title
		err = HandleTitle(g, reader)
	default:
		//fmt.Printf("unhandled packet 0x%X\n", p.ID)
	}
	return
}

func (g *Game) recvPacket() (*pk.Packet, error) {
	return pk.RecvPacket(g.Receiver, g.threshold > 0)
}

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
		time.Sleep(time.Millisecond * 50)
		if g.GetBlock(x, y, z).Id != b {
			break
		}
		g.SwingHand(true)
	}

	return nil
}

// PlaceBlock place a block in the position and wait
func (g *Game) PlaceBlock(x, y, z int, face Face) error {
	b := g.GetBlock(x, y, z).Id
	SendPlayerBlockPlacementPacket(g, x, y, z, face, 0, 0, 0, 0)

	for {
		time.Sleep(time.Millisecond * 50)
		if g.GetBlock(x, y, z).Id != b {
			break
		}
		g.SwingHand(true)
	}

	return nil
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
func (g *Game) GetPlayer() Player {
	return g.Player
}
func (g *Game) Attack(e LivingEntity) {
	SendUseEntityPacket(g, e.ID, 1, e.Position)
}
func (g *Game) SetPosition(v3 Vector3, onGround bool) {
	g.Motion <- func() {
		g.Player.X, g.Player.Y, g.Player.Z = v3.X, v3.Y, v3.Z
		g.Player.OnGround = onGround
		SendPlayerPositionPacket(g) // Update the location to the server
	}
}

// LookAt method turn player's hand and make it look at a point.
func (g *Game) LookAt(x, y, z float64) {
	x0, y0, z0 := g.Player.X, g.Player.Y, g.Player.Z
	x, y, z = x-x0, y-y0, z-z0

	r := math.Sqrt(x*x + y*y + z*z)
	yaw := -math.Atan2(x, z) / math.Pi * 180
	for yaw < 0 {
		yaw = 360 + yaw
	}
	pitch := -math.Asin(y/r) / math.Pi * 180

	g.LookYawPitch(float32(yaw), float32(pitch))
}

// LookYawPitch set player's hand to the direct by yaw and pitch.
// yaw can be [0, 360) and pitch can be (-180, 180).
// if |pitch|>90 the player's hand will be very strange.
func (g *Game) LookYawPitch(yaw, pitch float32) {
	g.Motion <- func() {
		g.Player.Yaw, g.Player.Pitch = yaw, pitch
		SendPlayerLookPacket(g) // Update the orientation to the server
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
		ID:   0x27,
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

func SendPlayerPositionAndLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackDouble(g.Player.X)...)
	data = append(data, pk.PackDouble(g.Player.Y)...)
	data = append(data, pk.PackDouble(g.Player.Z)...)
	data = append(data, pk.PackFloat(g.Player.Yaw)...)
	data = append(data, pk.PackFloat(g.Player.Pitch)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))
	//fmt.Printf("X:%f Y:%f Z:%f Yaw:%f Pitch:%f OnGround:%t\n", g.player.X, g.player.Y, g.player.Z, g.player.Yaw, g.player.Pitch, g.player.OnGround)

	g.SendChan <- pk.Packet{
		ID:   0x0E,
		Data: data,
	}
}
func SendPlayerLookPacket(g *Game) {
	var data []byte
	data = append(data, pk.PackFloat(g.Player.Yaw)...)
	data = append(data, pk.PackFloat(g.Player.Pitch)...)
	data = append(data, pk.PackBoolean(g.Player.OnGround))
	//fmt.Printf("Yaw:%f Pitch:%f OnGround:%t\n", g.player.Yaw, g.player.Pitch, g.player.OnGround)
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
		ID:   0x10,
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
		panic("UpdateVelocity: entity not found")
	}
	(*e).SetPosition((*e).Position.Add(velocity))
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
	c := g.World.Chunks[ChunkLoc{X: x >> 4, Y: z >> 4}]
	if c != nil {
		id, err := pk.UnpackVarInt(r)
		if err != nil {
			return err
		}
		c.Sections[y/16].Blocks[x&15][y&15][z&15] = Block{Id: uint(id)}
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
	sender, content := ExtractContent(cm.Text)
	timestamp := time.Now().UnixMilli()
	g.Events <- ChatMessageEvent{Content: content, Sender: sender, Timestamp: timestamp, Position: pos}

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
	g.Events <- EntityVelocityEvent{
		EntityID: entityID,
		Velocity: velocity,
	}
	//UpdateVelocity(g, entityID, velocity)
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

	flags, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch {
	case flags&0x01 == 0:
		g.Player.X = x
	case flags&0x01 == 1:
		g.Player.X += x
	case flags&0x02 == 0:
		g.Player.Y = y
	case flags&0x02 == 1:
		g.Player.Y += y
	case flags&0x04 == 0:
		g.Player.Z = z
	case flags&0x04 == 1:
		g.Player.Z += z
	case flags&0x08 == 0:
		g.Player.Yaw = yaw
	case flags&0x08 == 1:
		g.Player.Yaw += yaw
	case flags&0x10 == 0:
		g.Player.Pitch = pitch
	case flags&0x10 == 1:
		g.Player.Pitch += pitch
	default:
		panic("invalid flags")
	}

	TeleportID, _ := pk.UnpackVarInt(r)
	SendTeleportConfirmPacket(g, TeleportID)
	SendPlayerPositionAndLookPacket(g)
	return nil
}

func handleEntityLookAndRelativeMove(g *Game, r *bytes.Reader) error {
	ID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	E := g.World.Entities[ID]
	if E != nil {
		DeltaX, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaY, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaZ, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}

		yaw, err := r.ReadByte()
		if err != nil {
			return err
		}

		pitch, err := r.ReadByte()
		if err != nil {
			return err
		}
		var rotation = Vector2{
			X: float64(yaw) * 360 / 256,
			Y: float64(pitch) * 360 / 256,
		}
		E.SetRotation(rotation)

		og, err := r.ReadByte()
		if err != nil {
			return err
		}
		E.OnGround = og != 0x00
		E.SetPosition(Vector3{
			X: E.Position.X + float64(DeltaX)/4096,
			Y: E.Position.Y + float64(DeltaY)/4096,
			Z: E.Position.Z + float64(DeltaZ)/4096,
		})
	}
	return nil
}

func HandleEntityHeadLookPacket(g *Game, r *bytes.Reader) {
	ID, _ := pk.UnpackVarInt(r)
	E := g.World.Entities[ID]
	if E != nil {
		yaw, _ := r.ReadByte()
		pitch, _ := r.ReadByte()
		E.SetRotation(Vector2{
			X: float64(yaw) * 360 / 256,
			Y: float64(pitch) * 360 / 256,
		})
	}
}

func handleEntityRelativeMovePacket(g *Game, r *bytes.Reader) error {
	ID, err := pk.UnpackVarInt(r)
	if err != nil {
		return err
	}
	E := g.World.Entities[ID]
	if E != nil {
		DeltaX, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaY, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}
		DeltaZ, err := pk.UnpackInt16(r)
		if err != nil {
			return err
		}

		og, err := r.ReadByte()
		if err != nil {
			return err
		}
		E.OnGround = og != 0x00
		E.SetPosition(Vector3{
			X: E.Position.X + float64(DeltaX)/4096,
			Y: E.Position.Y + float64(DeltaY)/4096,
			Z: E.Position.Z + float64(DeltaZ)/4096,
		})
	}
	return nil
}

func HandlePluginPacket(g *Game, r *bytes.Reader) {
	// fmt.Println("Plugin Packet: ", p)
}

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
	return nil
}

func HandleSpawnPositionPacket(g *Game, r *bytes.Reader) (err error) {
	g.Info.SpawnPosition.X, g.Info.SpawnPosition.Y, g.Info.SpawnPosition.Z, err =
		pk.UnpackPosition(r)
	return
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
		SendPlayerPositionAndLookPacket(g)
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
		g.Player.SetPosition(Vector3{X: float64(x) + 0.5, Y: float64(y), Z: float64(z) + 0.5})
		time.Sleep(time.Millisecond * 50)
	}
	g.Player.SetPosition(Vector3{X: float64(x) + 0.5, Y: float64(y), Z: float64(z) + 0.5})
}
