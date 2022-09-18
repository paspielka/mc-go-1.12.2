package data

import (
	"fmt"
	. "github.com/edouard127/mc-go-1.12.2/items"
)

// Slot indicates a slot in a window
type Slot struct {
	ID    int
	Count byte
}
type slotNBT struct {
}

func (s Slot) String() string {
	return fmt.Sprintf("Slot[%s %d]", ItemNameByID[s.ID], s.Count)
}
