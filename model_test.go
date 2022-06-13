package main

import "testing"

func Test_flush(t *testing.T) {
	grid, diff := &Grid{}, make(Diff)
	diff[10] = map[int]Species{5: "a", 6: "b"}
	diff[11] = map[int]Species{7: "c"}

	flush(diff, grid)

	if len(diff) != 0 {
		t.Errorf("Expected diff to be empty but got %v", diff)
	}
	for x := 0; x < GridDimX; x++ {
		for y := 0; y < GridDimY; y++ {
			v := grid[x][y]
			if x == 10 && y == 5 {
				if v != "a" {
					t.Errorf("Expected (%v, %v) to be \"a\" but got %q", x, y, v)
				}
			} else if x == 10 && y == 6 {
				if v != "b" {
					t.Errorf("Expected (%v, %v) to be \"b\" but got %q", x, y, v)
				}
			} else if x == 11 && y == 7 {
				if v != "c" {
					t.Errorf("Expected (%v, %v) to be \"c\" but got %q", x, y, v)
				}
			} else if v != "" {
				t.Errorf("Expected (%v, %v) to be \"\" but got %q", x, y, v)
			}
		}
	}
}

func Test_merge(t *testing.T) {
	diff := make(Diff)
	diff[10] = map[int]Species{5: "a", 6: "b"}
	diff[11] = map[int]Species{7: "c"}
	newDiff := make(Diff)
	newDiff[10] = map[int]Species{6: "d", 9: "e"}
	newDiff[20] = map[int]Species{10: "f"}

	merge(newDiff, diff)

	if v := diff[10][5]; v != "a" {
		t.Errorf("Expected (10, 5) to be \"a\" but got %q", v)
	}
	if v := diff[10][6]; v != "d" {
		t.Errorf("Expected (10, 6) to be \"d\" but got %q", v)
	}
	if v := diff[10][9]; v != "e" {
		t.Errorf("Expected (10, 9) to be \"e\" but got %q", v)
	}
	if v := diff[11][7]; v != "c" {
		t.Errorf("Expected (11, 7) to be \"c\" but got %q", v)
	}
	if v := diff[20][10]; v != "f" {
		t.Errorf("Expected (20, 10) to be \"f\" but got %q", v)
	}
}

func Test_neighbors(t *testing.T) {
	grid := &Grid{}
	grid[10][10] = "a"
	grid[10][11] = "a"
	grid[11][11] = ""
	grid[11][12] = "a"
	grid[12][11] = "b"

	n, sMax := neighbors(grid, 11, 11)

	if n != 4 {
		t.Errorf("Expected number of neighbors be 4 but got %v", n)
	}
	if sMax != "a" {
		t.Errorf("Expected most populous species to be \"a\" but got %q", sMax)
	}
}

func Test_neighbors2(t *testing.T) {
	grid := &Grid{}
	grid[1][1] = "a"
	grid[GridDimX-1][GridDimY-1] = "b"
	grid[GridDimX-1][0] = "b"

	n, sMax := neighbors(grid, 0, 0)

	if n != 3 {
		t.Errorf("Expected number of neighbors to be 1 but got %v", n)
	}
	if sMax != "b" {
		t.Errorf("Expected most populous species to be \"b\" but got %q", sMax)
	}
}

func Test_nextState(t *testing.T) {
	grid, diff := &Grid{}, make(Diff)
	grid[10][10] = "a"
	grid[10][11] = "b"
	grid[11][11] = "a"
	grid[11][12] = "b"
	grid[12][11] = "c"

	nextState(grid, diff)

	if n := len(diff[10]); n < 2 || n > 3 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[10][10]; ok && v != "b" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[10][11]; !ok || v != "a" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[10][12]; !ok || v != "b" {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 10)

	if n := len(diff[11]); n < 1 || n > 2 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[11][11]; !ok || v != "" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[11][12]; ok && v != "a" && v != "c" {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 11)

	if n := len(diff[12]); n != 2 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[12][11]; !ok || (v != "a" && v != "b") {
		t.Errorf("Incorrect game state")
	}
	if v, ok := diff[12][12]; !ok || (v != "a" && v != "b" && v != "c") {
		t.Errorf("Incorrect game state")
	}
	delete(diff, 12)

	if len(diff) != 0 {
		t.Errorf("Incorrect game state")
	}
}
