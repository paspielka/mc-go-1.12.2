package PathFinding

import "math"

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
		if current.X == IStar.End.X && current.Y == IStar.End.Y && current.Z == IStar.End.Z {
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
	if n.X > 0 {
		neighbors = append(neighbors, &Node{
			X: n.X - 1,
			Y: n.Y,
			Z: n.Z,
		})
	}
	if n.X < 255 {
		neighbors = append(neighbors, &Node{
			X: n.X + 1,
			Y: n.Y,
			Z: n.Z,
		})
	}
	if n.Y > 0 {
		neighbors = append(neighbors, &Node{
			X: n.X,
			Y: n.Y - 1,
			Z: n.Z,
		})
	}
	if n.Y < 255 {
		neighbors = append(neighbors, &Node{
			X: n.X,
			Y: n.Y + 1,
			Z: n.Z,
		})
	}
	if n.Z > 0 {
		neighbors = append(neighbors, &Node{
			X: n.X,
			Y: n.Y,
			Z: n.Z - 1,
		})
	}
	if n.Z < 255 {
		neighbors = append(neighbors, &Node{
			X: n.X,
			Y: n.Y,
			Z: n.Z + 1,
		})
	}
	return neighbors
}

func (a *AStar) Contains(nodes []*Node, node *Node) bool {
	for _, n := range nodes {
		if n.X == node.X && n.Y == node.Y && n.Z == node.Z {
			return true
		}
	}
	return false
}

func (n *Node) GetCost(end *Node) int {
	return int(math.Abs(float64(n.X-end.X)) + math.Abs(float64(n.Y-end.Y)) + math.Abs(float64(n.Z-end.Z)))
}
