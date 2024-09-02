package graph

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type DiGraph struct {
	graph     map[interface{}]map[interface{}]int
	inDegree  map[interface{}]int
	queue     []interface{}
	isInQueue map[interface{}]bool
	buffer    int
	mutex     sync.Mutex
}

func NewDirectedGraph(buffer int) *DiGraph {
	return &DiGraph{
		graph:     make(map[interface{}]map[interface{}]int),
		inDegree:  make(map[interface{}]int),
		isInQueue: make(map[interface{}]bool),
		buffer:    buffer,
	}
}

func (g *DiGraph) QueuedCount() int {
	return len(g.inDegree)
}

func (g *DiGraph) AddVertex(v interface{}) {
	log.Debug("AddVertex", "v", v)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.graph[v] == nil {
		g.graph[v] = make(map[interface{}]int)
	}
	if _, ok := g.inDegree[v]; !ok {
		g.inDegree[v] = 0
	}
}

func (g *DiGraph) AddEdge(from, to interface{}) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.graph[from] == nil {
		g.graph[from] = make(map[interface{}]int)
	}

	if from != to && g.graph[from][to] == 0 {
		g.graph[from][to] = 1
		if _, ok := g.inDegree[from]; !ok {
			g.inDegree[from] = 0
		}
		g.inDegree[to]++
	}

	log.Debug("AddEdge", "from", from, "to", to, "fromInDegree", g.inDegree[from], "toInDegree", g.inDegree[to])
}

func (g *DiGraph) DelEdge(from, to interface{}) {
	log.Debug("DelEdge", "from", from, "to", to)
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.delEdge(from, to)
}

func (g *DiGraph) delEdge(from, to interface{}) {
	if g.graph[from][to] != 0 {
		g.inDegree[to]--
		delete(g.graph[from], to)
	}
}

// vertex will be deleted after consuming
func (g *DiGraph) Pipeline() <-chan interface{} {
	output := make(chan interface{}, g.buffer)

	go func() {
		for {
			// log.Debug("read from inDegree starts")
			g.mutex.Lock()
			for vertex, inDegree := range g.inDegree {
				// log.Debug("read from inDegree handling", "vertex", vertex, "indegree", inDegree)
				if inDegree == 0 {
					// log.Debug("inqueue2")
					g.inqueue(vertex)
				}
			}
			g.mutex.Unlock()
			// log.Debug("read from inDegree ends")

			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			g.mutex.Lock()

			if len(g.queue) == 0 {
				g.mutex.Unlock()

				time.Sleep(time.Second)
				continue
			}

			// log.Debug("handle queue starts")

			vertex := g.outqueue()
			log.Debug("DiGraph pop from queue", "vertex", vertex)
			output <- vertex
			log.Debug("DiGraph pop from queue2", "vertex", vertex)

			for neighbour := range g.graph[vertex] {
				log.Debug("DiGraph walk through neighbours", "vertex", vertex, "neighbour", neighbour,
					"neighbourInDegree", g.inDegree[neighbour])
				g.delEdge(vertex, neighbour)
				if g.inDegree[neighbour] == 0 {
					// log.Debug("inqueue3")
					g.inqueue(neighbour)
				}
			}

			g.mutex.Unlock()

			// log.Debug("handle queue ends")
		}
	}()

	return output
}

func (g *DiGraph) inqueue(vertex interface{}) {
	if !g.isInQueue[vertex] {
		log.Debug("DiGraph inqueue", "vertex", vertex)
		g.queue = append(g.queue, vertex)
		g.isInQueue[vertex] = true
		delete(g.inDegree, vertex)
	}
}

func (g *DiGraph) outqueue() interface{} {
	vertex := g.queue[0]
	g.queue = g.queue[1:]
	g.isInQueue[vertex] = false
	log.Debug("DiGraph outqueue", "vertex", vertex)

	return vertex
}
