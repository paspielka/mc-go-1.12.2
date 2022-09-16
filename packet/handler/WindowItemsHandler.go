package handler

import (
	"bytes"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func handleWindowItemsPacket(g *Game, r *bytes.Reader) (err error) {
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
