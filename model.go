package main

import (
	"math/rand"
)

const (
	gridDimX = 120
	gridDimY = 120
)

type species = string
type grid = [gridDimX][gridDimY]species
type diff = map[int]map[int]species

// flush copies a diff into a grid and empties the diff.
func flush(df diff, g *grid) {
	for x, ydiff := range df {
		for y, v := range ydiff {
			g[x][y] = v
		}
		delete(df, x)
	}
}

// merge copies a new diff into an existing diff.
func merge(newDiff diff, df diff) {
	for x, newYDiff := range newDiff {
		ydiff := getOrMakeYDiff(df, x)
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
func neighbors(g *grid, x int, y int) (int, species) {
	var left int
	if x == 0 {
		left = gridDimX - 1
	} else {
		left = x - 1
	}
	var right int
	if x == gridDimX-1 {
		right = 0
	} else {
		right = x + 1
	}
	var up int
	if y == 0 {
		up = gridDimY - 1
	} else {
		up = y - 1
	}
	var down int
	if y == gridDimY-1 {
		down = 0
	} else {
		down = y + 1
	}
	sCount := make(map[species]int)
	var s species
	if s = g[left][up]; s != "" {
		sCount[s]++
	}
	if s = g[x][up]; s != "" {
		sCount[s]++
	}
	if s = g[right][up]; s != "" {
		sCount[s]++
	}
	if s = g[left][y]; s != "" {
		sCount[s]++
	}
	if s = g[right][y]; s != "" {
		sCount[s]++
	}
	if s = g[left][down]; s != "" {
		sCount[s]++
	}
	if s = g[x][down]; s != "" {
		sCount[s]++
	}
	if s = g[right][down]; s != "" {
		sCount[s]++
	}
	var n int
	var sMax species
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

// nextState computes the changes between a grid's current state and next
// state, and writes the changes into a diff.
// nextState implements the original rules of Conway's Game of Life, and
// additionally sets a live cell's species to the most populous neighboring
// species as determined by the neighbors function.
func nextState(g *grid, df diff) {
	for x := 0; x < gridDimX; x++ {
		for y := 0; y < gridDimY; y++ {
			n, sMax := neighbors(g, x, y)
			current := g[x][y]
			if current != "" {
				if n != 2 && n != 3 {
					getOrMakeYDiff(df, x)[y] = ""
				} else if current != sMax {
					getOrMakeYDiff(df, x)[y] = sMax
				}
			} else if n == 3 {
				getOrMakeYDiff(df, x)[y] = sMax
			}
		}
	}
}

func getOrMakeYDiff(df diff, x int) map[int]species {
	ydiff, ok := df[x]
	if !ok {
		ydiff = make(map[int]species)
		df[x] = ydiff
	}
	return ydiff
}
