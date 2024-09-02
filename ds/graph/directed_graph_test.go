package graph

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

func TestDiGraph(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	type testcase struct {
		dependencies [][]int
	}

	testcases := []testcase{
		{
			dependencies: [][]int{{1}, {2}, {2, 3}, {4, 3}},
		},
		{
			dependencies: [][]int{{2, 1}, {1, 4}, {4, 3}},
		},
		{
			dependencies: [][]int{{2}, {4}, {3, 1}, {1, 2}, {4, 3}},
		},
		{
			dependencies: [][]int{{1, 2}, {2, 3}, {3, 4}, {5, 6}, {4, 5}},
		},
		{
			dependencies: [][]int{{1, 2}, {2, 3}, {3, 4}, {5, 6}, {4, 5}, {6, 1}},
		},
	}

	for i, tt := range testcases {
		t.Logf("run case#%d", i)

		dag := NewDirectedGraph()

		for j, d := range tt.dependencies {
			if len(d) == 1 {
				dag.AddVertex(d[0])
			} else if len(d) == 2 {
				dag.AddEdge(d[0], d[1])
			} else {
				t.Fatalf("invalid testcase dependencies#%d", j)
			}
		}

		outChan := dag.Pipeline()
		timer := time.NewTimer(2 * time.Second)

		res := []int{}
	waitRes:
		for {

			select {
			case <-timer.C:
				break waitRes
			case item := <-outChan:
				res = append(res, item.(int))
			}
		}

		t.Logf("res: %v", res)
		assert.True(t, isValidPipeline(tt.dependencies, res))
	}
}

func isValidPipeline(dependencies [][]int, output []int) bool {
	graph := make(map[int]map[int]bool)

	for _, d := range dependencies {
		if len(d) == 2 {
			if graph[d[0]] == nil {
				graph[d[0]] = make(map[int]bool)
			}
			graph[d[0]][d[1]] = true
		}
	}

	for i, o1 := range output {
		for j := i + 1; j < len(output); j++ {
			o2 := output[j]
			if graph[o2][o1] {
				fmt.Println("countercase: ", o1, o2)
				return false
			}
		}
	}

	return true
}
