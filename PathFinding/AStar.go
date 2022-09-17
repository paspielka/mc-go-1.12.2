package PathFinding

import (
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"math"
)

var IStar *AStar

type AStar struct {
	// Start is the start node
	Start *Node
	// End is the end node
	End *Node
	// OpenList is the list of nodes to be evaluated
	OpenList []*Node
	// ClosedList is the list of nodes already evaluated
	ClosedList []*Node
	// Path is the path from start to end
	Path *Path
	// PathFound is true if a path is found
	PathFound bool
	// MaxNodes is the maximum number of nodes to be evaluated
	MaxNodes int
	// NodesEvaluated is the number of nodes evaluated
	NodesEvaluated int
}

func Compute(ch chan bool) {
	for !IStar.PathFound && IStar.NodesEvaluated < IStar.MaxNodes {
		if len(IStar.OpenList) == 0 {
			break
		}
		current := IStar.OpenList[0]
		for _, node := range IStar.OpenList {
			if node.Cost < current.Cost {
				current = node
			}
		}
		if current.Position.DistanceTo(IStar.End.Position) < 1 {
			IStar.PathFound = true
			break
		}
		IStar.OpenList = append(IStar.OpenList[:0], IStar.OpenList[1:]...)
		IStar.ClosedList = append(IStar.ClosedList, current)
		for _, neighbor := range current.GetNeighbors() {
			if IStar.Contains(IStar.ClosedList, neighbor) {
				continue
			}
			if !IStar.Contains(IStar.OpenList, neighbor) {
				neighbor.Cost = neighbor.GetCost(IStar.End)
				IStar.OpenList = append(IStar.OpenList, neighbor)
			}
		}
		IStar.NodesEvaluated++
	}
	ch <- true
}

func NewAStar(start, end *Node) *AStar {
	IStar = &AStar{
		Start:          start,
		End:            end,
		OpenList:       []*Node{start},
		ClosedList:     []*Node{},
		Path:           &Path{},
		PathFound:      false,
		MaxNodes:       10000,
		NodesEvaluated: 0,
	}
	return IStar
}

func (n *Node) GetNeighbors() []*Node {
	var neighbors []*Node
	for _, axis := range []float64{n.Position.X, n.Position.Y, n.Position.Z} {
		if axis > 0 {
			neighbors = append(neighbors, &Node{Position: Vector3{axis - 1, n.Position.Y, n.Position.Z}})
		}
		if axis < 100 {
			neighbors = append(neighbors, &Node{Position: Vector3{axis + 1, n.Position.Y, n.Position.Z}})
		}
	}
	return neighbors
}

func (a *AStar) Contains(nodes []*Node, node *Node) bool {
	for _, n := range nodes {
		if n.Position.X == node.Position.X && n.Position.Y == node.Position.Y && n.Position.Z == node.Position.Z {
			return true
		}
	}
	return false
}

func (n *Node) GetCost(end *Node) int {
	return int(math.Abs(n.Position.X-end.Position.X) + math.Abs(n.Position.Y-end.Position.Y) + math.Abs(n.Position.Z-end.Position.Z))
}
