package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/packet/transaction"
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"time"
)

func handleUpdateHealthPacket(g *Game, r *bytes.Reader) (err error) {
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

	if g.Player.Health < 1 { //player is dead
		g.Events <- PlayerDeadEvent{} //Dead event
		SendPlayerPositionAndLookPacket(g)
		time.Sleep(time.Second * 2)  //wait for 2 sec make it more like a human
		SendClientStatusPacket(g, 0) //status 0 means perform respawn
	}
	return
}
