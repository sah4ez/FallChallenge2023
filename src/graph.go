package main

import (
	"fmt"
	"os"
)

var moves2 = [][]int{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {0, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

var moves = [][]int{
	{-1, 0}, {0, -1}, {1, 0}, {0, 1},
}

type Vertex struct {
	ID     Point
	Node   *Node
	Parent *Vertex

	Vertices map[Point]*Vertex
}

func NewVertex(node *Node) *Vertex {
	return &Vertex{
		ID:       node.Point,
		Node:     node,
		Vertices: make(map[Point]*Vertex, 0),
	}
}

func FillLocation(i, j int, nodes [][]*Node, used map[Point]struct{}) []*Node {
	fmt.Fprintf(os.Stderr, "fill position %d %d\n", i, j)

	if i < 0 || j < 0 {
		return nil
	}
	if i >= len(nodes) || j >= len(nodes[i]) {
		return nil
	}

	if used == nil {
		used = make(map[Point]struct{})
	}
	border := make([]*Node, 0)
	// first
	for _, b := range nodes[0] {
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}
	// last
	for _, b := range nodes[len(nodes)-1] {
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}

	for j := 0; j < len(nodes); j++ {
		b := nodes[j][0]
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}
	for j := 0; j < len(nodes); j++ {
		b := nodes[j][len(nodes[j])-1]
		if _, ok := used[b.Point]; ok {
			continue
		}
		used[b.Point] = struct{}{}
		border = append(border, b)
	}

	var f func(x, y int, target *Node, parent *Node, mark map[Point]struct{}) *Node

	f = func(x, y int, target *Node, parent *Node, mark map[Point]struct{}) *Node {
		if mark == nil {
			mark = map[Point]struct{}{}
		}
		result := nodes[x][y]
		if parent.I != result.I && parent.J != result.J {
			result.Parent = parent
		}
		min := LocationDistance(Point{result.I, result.J}, Point{target.I, target.J})
		for _, move := range moves2 {
			i := x + move[0]
			j := y + move[1]
			if i < 0 || j < 0 {
				continue
			}
			if i >= len(nodes) || j >= len(nodes[i]) {
				continue
			}
			node := nodes[i][j]
			if _, ok := mark[node.Point]; ok {
				continue
			}
			node.Parent = parent
			newMin := LocationDistance(Point{node.I, node.J}, Point{target.I, target.J})
			if min > newMin {
				min = newMin
				result = node
			}
			mark[node.Point] = struct{}{}
		}

		if result.I == target.I && result.J == target.J {
			target.Parent = parent
			return target
		}
		result = f(result.I, result.J, target, result, mark)
		return result
	}

	// markScored := map[Point]struct{}{}
	for m, b := range border {
		parent := nodes[i][j]
		newB := f(i, j, b, parent, nil)
		if newB == nil {
			continue
		}
		border[m] = newB

		next := newB
		// score := next.Score
		// if next.Parent != nil {
		// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, next.Parent.I, next.Parent.J)
		// } else {
		// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, parent.I, parent.J)
		// }
		path := []*Node{next}
		for {
			parent := next.Parent
			if parent == nil {
				break
			}
			path = append(path, parent)
			// parent.Score += score
			// fmt.Fprintf(os.Stderr, "(%d:%d)(%d:%d)->(%d:%d)->", i, j, next.I, next.J, parent.I, parent.J)
			if parent.I == i && parent.J == j {
				break
			}
			next = parent
		}
		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}
		score := 0
		odd := 0
		if len(nodes)%2 == 0 {
			odd = 1
		}
		for i, p := range path {
			score += p.Score
			// fmt.Fprintf(os.Stderr, "(%d:%d:%d)\n", p.I, p.J, score)
			if i == len(path)-1 {
				if p.X == 0 || p.Y == 0 {
					p.Score = score - odd
					p.Steps = i + odd
				} else {
					p.Score = score
					p.Steps = i
				}
			}
		}
	}
	return border
}

func NewGraph(i, j int, nodes [][]*Node, used map[Point]struct{}, parent *Vertex) *Vertex {

	if i < 0 || j < 0 {
		return nil
	}
	if i >= len(nodes) || j >= len(nodes[i]) {
		return nil
	}

	start := nodes[i][j]
	if _, ok := used[start.Point]; ok {
		return nil
	}
	start.Used = true
	root := &Vertex{
		ID:     start.Point,
		Node:   start,
		Parent: parent,
	}

	if used == nil {
		used = map[Point]struct{}{}
	}
	used[root.ID] = struct{}{}
	// fmt.Fprintln(os.Stderr, i, j)
	vertices := []*Vertex{}
	for _, move := range moves {
		i := i + move[0]
		j := j + move[1]
		if i < 0 || j < 0 {
			continue
		}
		if i >= len(nodes) || j >= len(nodes[i]) {
			continue
		}
		node := nodes[i][j]
		if node.Used {
			continue
		}
		node.Used = true
		// fmt.Fprintf(os.Stderr, "%v:%d:%d|", move, i, j)
		v := NewVertex(node)
		if v == nil {
			continue
		}
		vertices = append(vertices, v)
	}

	for _, v := range vertices {
		v.Parent = root
		if root.Vertices == nil {
			root.Vertices = map[Point]*Vertex{}
		}

		v := NewGraph(v.Node.I, v.Node.J, nodes, used, root)
		if v != nil {
			root.Vertices[v.ID] = v
		}
		used[v.ID] = struct{}{}
	}

	return root
}

func DebugVertex(v *Vertex) {
	fmt.Fprintf(os.Stderr, "%d:%d:%d:%d\n", v.ID.X, v.ID.Y, v.Node.Score, v.Node.Distance)
	for _, k := range v.Vertices {
		fmt.Fprintf(os.Stderr, "(%d:%d:%d:%d)|", k.ID.X, k.ID.Y, k.Node.Score, v.Node.Distance)
	}
	fmt.Fprintln(os.Stderr)
	for _, vv := range v.Vertices {
		DebugVertex(vv)
	}
}

// create a node that holds the graphs vertex as data
type node struct {
	v    *Vertex
	next *node
}

// create a queue data structure
type queue struct {
	head *node
	tail *node
}

// enqueue adds a new node to the tail of the queue
func (q *queue) enqueue(v *Vertex) {
	n := &node{v: v}

	// if the queue is empty, set the head and tail as the node value
	if q.tail == nil {
		q.head = n
		q.tail = n
		return
	}

	q.tail.next = n
	q.tail = n
}

// dequeue removes the head from the queue and returns it
func (q *queue) dequeue() *Vertex {
	n := q.head
	// return nil, if head is empty
	if n == nil {
		return nil
	}

	q.head = q.head.next

	// if there wasn't any next node, that
	// means the queue is empty, and the tail
	// should be set to nil
	if q.head == nil {
		q.tail = nil
	}

	return n.v
}

func BFS(startVertex *Vertex, visitCb func(*Vertex)) {
	// initialize queue and visited vertices map
	vertexQueue := &queue{}
	visitedVertices := map[Point]struct{}{}

	currentVertex := startVertex
	// start a continuous loop
	for {
		// visit the current node
		visitCb(currentVertex)
		visitedVertices[currentVertex.ID] = struct{}{}

		// for each neighboring vertex, push it to the queue
		// if it isn't already visited
		for _, v := range currentVertex.Vertices {
			if _, ok := visitedVertices[v.ID]; !ok {
				vertexQueue.enqueue(v)
			}
		}

		// change the current vertex to the next one
		// in the queue
		currentVertex = vertexQueue.dequeue()
		// if the queue is empty, break out of the loop
		if currentVertex == nil {
			break
		}
	}
}
