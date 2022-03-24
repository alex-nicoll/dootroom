package main

import (
	"encoding/json"
	"time"
)

const (
	GridDimX = 100
	GridDimY = 100
)

type Grid = [GridDimX][GridDimY]bool
type Diff = map[int]map[int]bool

// flush copies a Diff into a Grid and empties the Diff.
func flush(diff Diff, grid *Grid) {
	for x, ydiff := range diff {
		for y, v := range ydiff {
			grid[x][y] = v
		}
		delete(diff, x)
	}
}

// merge copies a new Diff into an existing Diff.
func merge(newDiff Diff, diff Diff) {
	for x, newYDiff := range newDiff {
		ydiff, ok := diff[x]
		if !ok {
			ydiff = make(map[int]bool)
			diff[x] = ydiff
		}
		for y, v := range newYDiff {
			ydiff[y] = v
		}
	}
}

// neighbors returns the number of live cells in the neighborhood of cell (x,y).
func neighbors(grid *Grid, x int, y int) int {
	var n int
	if x > 0 {
		if y > 0 && grid[x-1][y-1] {
			n++
		}
		if grid[x-1][y] {
			n++
		}
		if y < GridDimY-1 && grid[x-1][y+1] {
			n++
		}
	}
	if x < GridDimX-1 {
		if y > 0 && grid[x+1][y-1] {
			n++
		}
		if grid[x+1][y] {
			n++
		}
		if y < GridDimY-1 && grid[x+1][y+1] {
			n++
		}
	}
	if y > 0 && grid[x][y-1] {
		n++
	}
	if y < GridDimY-1 && grid[x][y+1] {
		n++
	}
	return n
}

// nextState computes the changes between a Grid's current state and next
// state, and writes the changes into a Diff.
func nextState(grid *Grid, diff Diff) {
	for x, ygrid := range grid {
		for y := range ygrid {
			n := neighbors(grid, x, y)
			if grid[x][y] && n != 2 && n != 3 {
				ydiff, ok := diff[x]
				if !ok {
					ydiff = make(map[int]bool)
					diff[x] = ydiff
				}
				ydiff[y] = false
			} else if !grid[x][y] && n == 3 {
				ydiff, ok := diff[x]
				if !ok {
					ydiff = make(map[int]bool)
					diff[x] = ydiff
				}
				ydiff[y] = true
			}
		}
	}
}

type Merge struct {
	diff Diff
}

type InitListener struct {
	li *Listener
}

type Tick struct{}

func model(in chan interface{}, hubChan chan<- interface{}) {
	grid, diff := &Grid{}, make(Diff)

	// We could handle one Merge message and an arbitrary number of
	// InitListener messages concurrently. But for simplicity of implementation
	// we'll have one goroutine handle all three message types.
	for {
		switch m := (<-in).(type) {
		case *Merge:
			merge(m.diff, diff)
		case *InitListener:
			message, _ := json.Marshal(grid)
			hubChan <- &Forward{m.li, message}
		case *Tick:
			if len(diff) != 0 {
				message, _ := json.Marshal(diff)
				hubChan <- &Broadcast{message}
				flush(diff, grid)
			}
			nextState(grid, diff)
		}
	}
}

func clock(modelChan chan<- interface{}) {
	for {
		time.Sleep(150 * time.Millisecond)
		modelChan <- &Tick{}
	}
}
