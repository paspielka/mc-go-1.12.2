package handler

import (
	"bytes"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func HandlePack(g *Game, p *pk.Packet) (err error) {
	//fmt.Printf("recv packet 0x%X\n", p.ID)
	reader := bytes.NewReader(p.Data)

	switch p.ID {
	case 0x23:
		err = handleJoinGamePacket(g, reader)
		g.Events <- JoinGameEvent{
			EntityID: g.Player.ID,
		}
	case 0x18:
		handlePluginPacket(g, reader)
	case 0x0D:
		err = handleServerDifficultyPacket(g, reader)
	case 0x46:
		err = handleSpawnPositionPacket(g, reader)
	case 0x2C:
		err = handlePlayerAbilitiesPacket(g, reader)
		g.SendChan <- *g.Settings.Pack()
	case 0x3A:
		err = handleHeldItemPacket(g, reader)
	/*case 0x20:
	err = handleChunkDataPacket(g, p)
	g.events <- BlockChangeEvent{}*/
	case 0x2F:
		err = handlePlayerPositionAndLookPacket(g, reader)
	case 0x54:
		// handleDeclareRecipesPacket(g, reader)
	case 0x29:
		// err = handleEntityLookAndRelativeMove(g, reader)
	case 0x28:
		handleEntityHeadLookPacket(g, reader)
	case 0x1F:
		err = handleKeepAlivePacket(g, reader)
	/*case 0x26:
	handleEntityPacket(g, reader)*/
	case 0x05:
		err = handleSpawnPlayerPacket(g, reader)
	case 0x15:
		err = handleWindowItemsPacket(g, reader)
	case 0x44:
		err = handleUpdateHealthPacket(g, reader)
	case 0x0F:
		err = handleChatMessagePacket(g, reader)
	case 0x0B:
		err = handleBlockChangePacket(g, reader)
	case 0x10:
		err = handleMultiBlockChangePacket(g, reader)
		g.Events <- BlockChangeEvent{}
	case 0x1A:
		// Should assume that the server has already closed the connection by the time the packet arrives.
		g.Events <- DisconnectEvent{Text: "disconnect"}
		err = fmt.Errorf("disconnect")
	case 0x17:
	// 	err = handleSetSlotPacket(g, reader)
	case 0x4D:
		err = handleSoundEffect(g, reader)
	case 0x3E: // Entity velocity
		err = handleEntityVelocity(g, reader)
	case 0x48: // Title
		err = handleTitle(g, reader)
	default:
		//fmt.Printf("unhandled packet 0x%X\n", p.ID)
	}
	return
}
