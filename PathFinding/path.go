package PathFinding

import (
	. "github.com/edouard127/mc-go-1.12.2/maths"
)

type Node struct {
	Position Vector3
	Cost     int
}

type Path struct {
	Nodes []*Node
}
