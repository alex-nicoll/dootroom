package main

import (
	"encoding/json"
	"log"
	"time"
)

const (
	GridDimX = 100
	GridDimY = 100
)

var grid = [GridDimX][GridDimY]bool{}

var diff = make(map[int]map[int]bool)

func flush() {
	for x, ydiff := range diff {
		for y, v := range ydiff {
			grid[x][y] = v
		}
		delete(diff, x)
	}
}

func merge(newDiff map[int]map[int]bool) {
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

func neighbors(x int, y int) int {
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

func nextState() {
	for x, ygrid := range grid {
		for y := range ygrid {
			n := neighbors(x, y)
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
	diff map[int]map[int]bool
}

type Init struct {
	li *Listener
}

type Tick struct{}

func model(in chan interface{}, hubChan chan<- interface{}) {
	for {
		m, ok := <-in
		if !ok {
			return
		}
		switch m := m.(type) {
		case *Merge:
			merge(m.diff)
		case *Init:
			// TODO: Merge and Init can run concurrently with each other, and Init with itself...
			message, _ := json.Marshal(&grid)
			hubChan <- &Forward{m.li, message}
		case *Tick:
			if len(diff) != 0 {
				log.Println(diff)
				message, _ := json.Marshal(diff)
				hubChan <- &Broadcast{message}
				flush()
			}
			nextState()
		}
	}
}

func clock(modelChan chan<- interface{}) {
	for {
		time.Sleep(150 * time.Millisecond)
		modelChan <- &Tick{}
	}
}
