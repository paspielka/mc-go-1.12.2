package PathFinding

import (
	. "github.com/edouard127/mc-go-1.12.2/struct"
	"math"
)

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

// Compute finds the best path from start to end in Minecraft world using A* algorithm like in https://github.com/lambda-plugins/ElytraBot/blob/main/src/main/kotlin/AStar.kt
func Compute(IStar *AStar) *AStar {
	// Read the code at https://github.com/lambda-plugins/ElytraBot/blob/main/src/main/kotlin/AStar.kt
	// for a better understanding of the algorithm
	// then compare it to this code
	for len(IStar.OpenList) > 0 {
		// Get the node with the lowest cost
		var lowestCostNode *Node
		for _, node := range IStar.OpenList {
			if lowestCostNode == nil || node.Cost < lowestCostNode.Cost {
				lowestCostNode = node
			}
		}
		// Remove the node from the open list
		for i, node := range IStar.OpenList {
			if node.Position.X == lowestCostNode.Position.X && node.Position.Y == lowestCostNode.Position.Y && node.Position.Z == lowestCostNode.Position.Z {
				IStar.OpenList = append(IStar.OpenList[:i], IStar.OpenList[i+1:]...)
				break
			}
		}
		// Add the node to the closed list
		IStar.ClosedList = append(IStar.ClosedList, lowestCostNode)
		// Check if the node is the end node
		if lowestCostNode.Position.X == IStar.End.Position.X && lowestCostNode.Position.Y == IStar.End.Position.Y && lowestCostNode.Position.Z == IStar.End.Position.Z {
			IStar.Path.Nodes = append(IStar.Path.Nodes, lowestCostNode)
			IStar.Path.BackTrace()
			IStar.PathFound = true
			return IStar
		}
		// Get the neighbors of the node
		neighbors := lowestCostNode.GetNeighbors()
		for _, neighbor := range neighbors {
			// Check if the neighbor is in the closed list
			if IStar.Contains(IStar.ClosedList, neighbor) {
				continue
			}
			// Check if the neighbor is in the open list
			if IStar.Contains(IStar.OpenList, neighbor) {
				continue
			}
			// Add the neighbor to the open list
			neighbor.Cost = neighbor.GetCost(IStar.End)
			IStar.OpenList = append(IStar.OpenList, neighbor)
		}
		IStar.NodesEvaluated++
		if IStar.NodesEvaluated > IStar.MaxNodes {
			return IStar
		}
	}
	return IStar
}

func NewAStar(start, end *Node) *AStar {
	return &AStar{
		Start:          start,
		End:            end,
		OpenList:       []*Node{start},
		ClosedList:     []*Node{},
		Path:           &Path{},
		PathFound:      false,
		MaxNodes:       10000,
		NodesEvaluated: 0,
	}
}

func (n *Node) GetNeighbors() []*Node {
	// Get nearby nodes
	var neighbors []*Node
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			for z := -1; z <= 1; z++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}
				neighbors = append(neighbors, &Node{
					Position: Vector3{
						X: n.Position.X + float64(x),
						Y: n.Position.Y + float64(y),
						Z: n.Position.Z + float64(z),
					},
					Cost: n.Cost + 1,
				})
			}
		}
	}
	return neighbors
}

func (p *Path) BackTrace() {
	for i := len(p.Nodes) - 1; i >= 0; i-- {
		p.Nodes = append(p.Nodes, p.Nodes[i])
	}
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
