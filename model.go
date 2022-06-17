package main

import (
	"math/rand"
)

const (
	GridDimX = 120
	GridDimY = 120
)

type Species = string
type Grid = [GridDimX][GridDimY]Species
type Diff = map[int]map[int]Species

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
		ydiff := getOrMakeYDiff(diff, x)
		for y, v := range newYDiff {
			ydiff[y] = v
		}
	}
}

// neighbors returns the number of live cells and most populous species in the
// neighborhood of cell (x,y). If multiple species are tied for most populous,
// neighbors chooses one at random. The neighborhood of a cell is defined such
// that the left and right edges of the grid are stitched together, and the top
// and bottom edges are stitched together.
func neighbors(grid *Grid, x int, y int) (int, Species) {
	var left int
	if x == 0 {
		left = GridDimX - 1
	} else {
		left = x - 1
	}
	var right int
	if x == GridDimX-1 {
		right = 0
	} else {
		right = x + 1
	}
	var up int
	if y == 0 {
		up = GridDimY - 1
	} else {
		up = y - 1
	}
	var down int
	if y == GridDimY-1 {
		down = 0
	} else {
		down = y + 1
	}
	sCount := make(map[Species]int)
	var s Species
	if s = grid[left][up]; s != "" {
		sCount[s]++
	}
	if s = grid[x][up]; s != "" {
		sCount[s]++
	}
	if s = grid[right][up]; s != "" {
		sCount[s]++
	}
	if s = grid[left][y]; s != "" {
		sCount[s]++
	}
	if s = grid[right][y]; s != "" {
		sCount[s]++
	}
	if s = grid[left][down]; s != "" {
		sCount[s]++
	}
	if s = grid[x][down]; s != "" {
		sCount[s]++
	}
	if s = grid[right][down]; s != "" {
		sCount[s]++
	}
	var n int
	var sMax Species
	var sMaxCount int
	for k, v := range sCount {
		n += v
		if v > sMaxCount || (v == sMaxCount && rand.Intn(2) == 0) {
			sMax = k
			sMaxCount = v
		}
	}
	return n, sMax
}

// nextState computes the changes between a Grid's current state and next
// state, and writes the changes into a Diff.
// nextState implements the original rules of Conway's Game of Life, and
// additionally sets a live cell's species to the most populous neighboring
// species as determined by the neighbors function.
func nextState(grid *Grid, diff Diff) {
	for x := 0; x < GridDimX; x++ {
		for y := 0; y < GridDimY; y++ {
			n, sMax := neighbors(grid, x, y)
			current := grid[x][y]
			if current != "" {
				if n != 2 && n != 3 {
					getOrMakeYDiff(diff, x)[y] = ""
				} else if current != sMax {
					getOrMakeYDiff(diff, x)[y] = sMax
				}
			} else if n == 3 {
				getOrMakeYDiff(diff, x)[y] = sMax
			}
		}
	}
}

func getOrMakeYDiff(diff Diff, x int) map[int]Species {
	ydiff, ok := diff[x]
	if !ok {
		ydiff = make(map[int]Species)
		diff[x] = ydiff
	}
	return ydiff
}
