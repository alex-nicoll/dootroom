package main

import (
	"testing"
)

func reset() {
	grid = [GridDimX][GridDimY]bool{}
	diff = make(map[int]map[int]bool)
}

func Test_flush(t *testing.T) {
	reset()
	diff[10] = map[int]bool{5: true, 6: true}
	diff[11] = map[int]bool{7: true}

	flush()

	if len(diff) != 0 {
		t.Errorf("Expected diff to be empty but got %v", diff)
	}
	for x := 0; x < GridDimX; x++ {
		for y := 0; y < GridDimY; y++ {
			if x == 10 && (y == 5 || y == 6) {
				if !grid[x][y] {
					t.Errorf("Expected (%v, %v) to be true but got false", x, y)
				}
			} else if x == 11 && y == 7 {
				if !grid[x][y] {
					t.Errorf("Expected (%v, %v) to be true but got false", x, y)
				}
			} else {
				if grid[x][y] {
					t.Errorf("Expected (%v, %v) to be false but got true", x, y)
				}
			}
		}
	}
}

func Test_merge(t *testing.T) {
	reset()
	diff[10] = map[int]bool{5: true, 6: true}
	diff[11] = map[int]bool{7: true}
	newDiff := make(map[int]map[int]bool)
	newDiff[10] = map[int]bool{6: false, 9: true}
	newDiff[20] = map[int]bool{10: false}

	merge(newDiff)

	if !diff[10][5] {
		t.Errorf("Expected (10, 5) to be true but got false")
	}
	if diff[10][6] {
		t.Errorf("Expected (10, 6) to be false but got true")
	}
	if !diff[10][9] {
		t.Errorf("Expected (10, 9) to be true but got false")
	}
	if !diff[11][7] {
		t.Errorf("Expected (11, 7) to be true but got false")
	}
	if diff[20][10] {
		t.Errorf("Expected (20, 10) to be false but got true")
	}
}

func Test_neighbors(t *testing.T) {
	reset()
	grid[10][10] = true
	grid[10][11] = true
	grid[11][11] = true
	grid[11][12] = true
	grid[12][11] = true

	n := neighbors(11, 11)

	if n != 4 {
		t.Errorf("Expected 4 but got %v", n)
	}
}

func Test_neighbors2(t *testing.T) {
	reset()
	grid[1][1] = true

	n := neighbors(0, 0)

	if n != 1 {
		t.Errorf("Expected 1 but got %v", n)
	}
}

func Test_nextState(t *testing.T) {
	reset()
	grid[10][10] = true
	grid[10][11] = true
	grid[11][11] = true
	grid[11][12] = true
	grid[12][11] = true

	nextState()

	if len(diff[10]) != 1 {
		t.Errorf("Incorrect game state")
	}
	v, ok := diff[10][12]
	if !ok {
		t.Errorf("Incorrect game state")
	}
	if !v {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 10)
	if len(diff[11]) != 1 {
		t.Errorf("Incorrect game state")
	}
	v, ok = diff[11][11]
	if !ok {
		t.Errorf("Incorrect game state")
	}
	if v {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 11)
	if len(diff[12]) != 1 {
		t.Errorf("Incorrect game state")
	}
	v, ok = diff[12][12]
	if !ok {
		t.Errorf("Incorrect game state")
	}
	if !v {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 12)
	if len(diff) != 0 {
		t.Errorf("Incorrect game state")
	}
}

func Test_model(t *testing.T) {
	reset()
	modelChan := make(chan interface{})
	hubChan := make(chan interface{})
	mIn := &Merge{
		diff: map[int]map[int]bool{
			30: map[int]bool{
				30: true,
				31: true,
			},
			31: map[int]bool{
				31: true,
				32: true,
			},
			32: map[int]bool{
				31: true,
			},
		},
	}

	go model(modelChan, hubChan)
	modelChan <- mIn
	modelChan <- &Tick{}
	mOut := (<-hubChan).(*Broadcast)

	json := string(mOut.message)
	expectedJson := "{\"30\":{\"30\":true,\"31\":true},\"31\":{\"31\":true,\"32\":true},\"32\":{\"31\":true}}"
	if json != expectedJson {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	modelChan <- &Tick{}
	mOut = (<-hubChan).(*Broadcast)

	json = string(mOut.message)
	expectedJson = "{\"30\":{\"32\":true},\"31\":{\"31\":false},\"32\":{\"32\":true}}"
	if json != expectedJson {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	close(modelChan)
}
