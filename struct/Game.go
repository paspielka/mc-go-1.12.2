package _struct

import (
	"bufio"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/packet/handler"
	. "github.com/edouard127/mc-go-1.12.2/packet/transaction"
	. "github.com/edouard127/mc-go-1.12.2/world"
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

// UseItem use the item in hand.
// if hand is true, swing the main hand
func (g *Game) UseItem(hand bool) {
	if hand {
		SendUseItemPacket(g, 0)
	} else {
		SendUseItemPacket(g, 1)
	}
}
