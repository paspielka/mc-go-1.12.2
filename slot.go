package gomcbot

import (
	"bytes"
	"fmt"
	pk "github.com/edouard127/mc-go-1.12.2/packet"
)

// Slot indicates a slot in a window
type Slot struct {
	ID    int
	Count byte
}

type slotNBT struct {
}

func unpackSolt(b *bytes.Reader) (Slot, error) {
	index := 0
	p, err := b.ReadByte()
	if err != nil {
		return Slot{}, err
	}
	Present := p != 0x00
	index++
	if Present {
		itemID, err := pk.UnpackVarInt(b)
		if err != nil {
			return Slot{}, err
		}
		count, err := b.ReadByte()
		if err != nil {
			return Slot{}, err
		}
		index++

		//nbt.Unmarshal(nbt.Uncompressed)

		return Slot{
			ID:    int(itemID),
			Count: count,
		}, nil
	}
	return Slot{}, nil
}

func (s Slot) String() string {
	return fmt.Sprintf("solt[%s %d]", itemNameByID[s.ID], s.Count)
}
