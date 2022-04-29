package inject

import (
	"fmt"
	"sort"
)

type EdgeNode struct {
	index    int
	priority int
}

func (e EdgeNode) String() string {
	return fmt.Sprintf("%d(%d)", e.index, e.priority)
}

// Edge represents a pair of vertexes.  Each vertex is an opaque type.
type Edge [2]EdgeNode

func (e Edge) String() string {
	return fmt.Sprintf("%d-%d", e[0], e[1])
}

// Toposort performs a topological sort of the DAG defined by given edges.
//
// Takes a slice of Edge, where each element is a vertex pair representing an
// edge in the graph.  Each pair can also be considered a dependency
// relationship where Edge[0] must happen before Edge[1].
//
// To include a node that is not connected to the rest of the graph, include a
// node with one nil vertex.  It can appear anywhere in the sorted output.
//
// Returns an ordered list of vertexes where each vertex occurs before any of
// its destination vertexes.  An error is returned if a cycle is detected.
func Toposort(edges []Edge, allNodes []EdgeNode) ([]EdgeNode, []EdgeNode, error) {
	g, sortNodes, err := makeGraph(edges, allNodes)
	if err != nil {
		return nil, nil, err
	}
	sorted := make([]EdgeNode, 0, len(g))

	// Create map of vertexes to incoming edge count, and set counts to 0
	inDegree := make(map[EdgeNode]int, len(g))
	for _, n := range sortNodes {
		inDegree[n] = 0
	}

	// For each vertex u, get adjacent list
	for _, adjacent := range g {
		// For each vertex v adjacent to u
		for _, v := range adjacent {
			// Increment inDegree[v]
			inDegree[v]++
		}
	}

	// Make a list next consisting of all vertexes u such that inDegree[u] = 0
	var next []EdgeNode
	for u, deg := range inDegree {
		if deg == 0 {
			next = append(next, u)
		}
	}

	// While next is not empty...
	for len(next) > 0 {
		sort.Slice(next, sortEdgeNodes(next))
		// Pop a vertex from next and call it vertex u
		u := next[0]
		next = next[1:]
		// Add u to the end sorted list
		sorted = append(sorted, u)
		// For each vertex v adjacent to sorted vertex u
		for _, v := range g[u] {
			// Decrement count of incoming edges
			inDegree[v]--
			// Enqueue nodes with no incoming edges
			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	// Check for cycle
	if len(sorted) < len(g) {
		var cycleNodes []EdgeNode
		for u, deg := range inDegree {
			if deg != 0 {
				cycleNodes = append(cycleNodes, u)
			}
		}
		sort.Slice(cycleNodes, sortEdgeNodes(cycleNodes))
		return nil, cycleNodes, fmt.Errorf("graph contains cycle in nodes %s", cycleNodes)
	}

	// Return the sorted vertex list
	return sorted, nil, nil
}

// makeGraph creates a map of source node to destination nodes.  An edge with
// only one vertex is added to the graph, if it is not already in the graph.
func makeGraph(edges []Edge, allNodes []EdgeNode) (map[EdgeNode][]EdgeNode, []EdgeNode, error) {
	graph := make(map[EdgeNode][]EdgeNode, len(edges)+1)
	var startNodes []EdgeNode
	edgeNodeSet := make(map[EdgeNode]bool)
	for i := range edges {
		s, e := edges[i][0], edges[i][1]
		graph[s] = append(graph[s], e)
		startNodes = append(startNodes, s)
		edgeNodeSet[s] = true
		edgeNodeSet[e] = true
	}
	for _, n := range allNodes {
		if edgeNodeSet[n] {
			continue
		}
		graph[n] = nil
		startNodes = append(startNodes, n)
	}
	for _, nodes := range graph {
		sort.Slice(nodes, sortEdgeNodes(nodes))
	}
	sort.Slice(startNodes, sortEdgeNodes(startNodes))
	return graph, startNodes, nil
}

func sortEdgeNodes(nodes []EdgeNode) func(i int, j int) bool {
	return func(i, j int) bool {
		if nodes[i].priority == nodes[j].priority {
			return nodes[i].index < nodes[j].index
		}
		return nodes[i].priority > nodes[j].priority
	}
}
