package PathFinding

type Node struct {
	X, Y, Z int
	Cost    int
}

type Path struct {
	Nodes []*Node
}
