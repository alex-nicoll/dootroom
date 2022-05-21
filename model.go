package main

import (
	"math/rand"
)

const (
	GridDimX = 100
	GridDimY = 100
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
// neighbors chooses one at random.
func neighbors(grid *Grid, x int, y int) (int, Species) {
	sCount := make(map[Species]int)
	var s Species
	if x > 0 {
		if y > 0 {
			s = grid[x-1][y-1]
			if s != "" {
				sCount[s]++
			}
		}
		s = grid[x-1][y]
		if s != "" {
			sCount[s]++
		}
		if y < GridDimY-1 {
			s = grid[x-1][y+1]
			if s != "" {
				sCount[s]++
			}
		}
	}
	if x < GridDimX-1 {
		if y > 0 {
			s = grid[x+1][y-1]
			if s != "" {
				sCount[s]++
			}
		}
		s = grid[x+1][y]
		if s != "" {
			sCount[s]++
		}
		if y < GridDimY-1 {
			s = grid[x+1][y+1]
			if s != "" {
				sCount[s]++
			}
		}
	}
	if y > 0 {
		s = grid[x][y-1]
		if s != "" {
			sCount[s]++
		}
	}
	if y < GridDimY-1 {
		s = grid[x][y+1]
		if s != "" {
			sCount[s]++
		}
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
