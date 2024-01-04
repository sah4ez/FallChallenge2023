package main

import (
	"fmt"
	"os"
)

var moves = [][]int{
	{-1, 0}, {0, -1}, {1, 0}, {0, 1},
}

type Vertex struct {
	ID   Point
	Node *Node

	Vertices map[Point]*Vertex
}

func NewVertex(node *Node) *Vertex {
	return &Vertex{
		ID:       node.Point,
		Node:     node,
		Vertices: make(map[Point]*Vertex, 0),
	}
}

func NewGraph(i, j int, nodes [][]*Node, used map[Point]struct{}) *Vertex {

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
		ID:   start.Point,
		Node: start,
	}

	if used == nil {
		used = map[Point]struct{}{}
	}
	used[root.ID] = struct{}{}
	// fmt.Fprintln(os.Stderr, i, j)
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
		// это гавно не работае. надо как-то починить построениение гарафа из середины массива
		node.Used = true
		// fmt.Fprintf(os.Stderr, "%v:%d:%d|", move, i, j)
		v := NewGraph(i, j, nodes, used)
		if v == nil {
			continue
		}
		if root.Vertices == nil {
			root.Vertices = map[Point]*Vertex{}
		}

		used[node.Point] = struct{}{}
		root.Vertices[node.Point] = v
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
