package PathFinding

import (
	"fmt"
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
func Compute(IStar *AStar, g *Game) *AStar {
	for loc, _ := range g.World.Chunks {
		for x := 0; x < 16; x++ {
			for y := 0; y < 256; y++ {
				for z := 0; z < 16; z++ {
					if !g.GetBlock(x, y, z).IsAir() {
						fmt.Println(g.GetBlock(x, y, z))
						IStar.ClosedList = append(IStar.ClosedList, &Node{
							Position: Vector3{
								X: float64(loc.X*16 + x),
								Y: float64(y),
								Z: float64(loc.Y*16 + z),
							},
							Cost: 0,
						})
					}
				}
			}
		}
	}
	for len(IStar.OpenList) > 0 {
		// Get the node with the lowest cost
		var currentNode *Node
		var currentNodeIndex int
		for i, node := range IStar.OpenList {
			if currentNode == nil || node.Cost < currentNode.Cost {
				currentNode = node
				currentNodeIndex = i
			}
		}
		// Remove the node from the open list
		IStar.OpenList = append(IStar.OpenList[:currentNodeIndex], IStar.OpenList[currentNodeIndex+1:]...)
		// Add the node to the closed list
		IStar.ClosedList = append(IStar.ClosedList, currentNode)
		// Check if the node is the end node
		if currentNode.Position.X == IStar.End.Position.X && currentNode.Position.Y == IStar.End.Position.Y && currentNode.Position.Z == IStar.End.Position.Z {
			IStar.PathFound = true
			break
		}
		// Get the neighbors of the node
		neighbors := currentNode.GetNeighbors()
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
			neighbor.Cost = currentNode.Cost + 1 + neighbor.GetCost(IStar.End)
			IStar.OpenList = append(IStar.OpenList, neighbor)
		}
		IStar.NodesEvaluated++
		if IStar.NodesEvaluated > IStar.MaxNodes {
			break
		}
	}
	if IStar.PathFound {
		var currentNode *Node
		for _, node := range IStar.ClosedList {
			if node.Position.X == IStar.End.Position.X && node.Position.Y == IStar.End.Position.Y && node.Position.Z == IStar.End.Position.Z {
				currentNode = node
				break
			}
		}
		for currentNode != nil {
			IStar.Path.Nodes = append(IStar.Path.Nodes, currentNode)
		}
		IStar.Path.BackTrace()
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
