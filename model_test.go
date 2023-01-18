package main

import "testing"

func Test_flush(t *testing.T) {
	g, df := &grid{}, make(diff)
	df[10] = map[int]species{5: "a", 6: "b"}
	df[11] = map[int]species{7: "c"}

	flush(df, g)

	if len(df) != 0 {
		t.Errorf("Expected diff to be empty but got %v", df)
	}
	for x := 0; x < gridDimX; x++ {
		for y := 0; y < gridDimY; y++ {
			v := g[x][y]
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
	df := make(diff)
	df[10] = map[int]species{5: "a", 6: "b"}
	df[11] = map[int]species{7: "c"}
	newDiff := make(diff)
	newDiff[10] = map[int]species{6: "d", 9: "e"}
	newDiff[20] = map[int]species{10: "f"}

	merge(newDiff, df)

	if v := df[10][5]; v != "a" {
		t.Errorf("Expected (10, 5) to be \"a\" but got %q", v)
	}
	if v := df[10][6]; v != "d" {
		t.Errorf("Expected (10, 6) to be \"d\" but got %q", v)
	}
	if v := df[10][9]; v != "e" {
		t.Errorf("Expected (10, 9) to be \"e\" but got %q", v)
	}
	if v := df[11][7]; v != "c" {
		t.Errorf("Expected (11, 7) to be \"c\" but got %q", v)
	}
	if v := df[20][10]; v != "f" {
		t.Errorf("Expected (20, 10) to be \"f\" but got %q", v)
	}
}

func Test_neighbors(t *testing.T) {
	g := &grid{}
	g[10][10] = "a"
	g[10][11] = "a"
	g[11][11] = ""
	g[11][12] = "a"
	g[12][11] = "b"

	n, sMax := neighbors(g, 11, 11)

	if n != 4 {
		t.Errorf("Expected number of neighbors be 4 but got %v", n)
	}
	if sMax != "a" {
		t.Errorf("Expected most populous species to be \"a\" but got %q", sMax)
	}
}

func Test_neighbors2(t *testing.T) {
	g := &grid{}
	g[1][1] = "a"
	g[gridDimX-1][gridDimY-1] = "b"
	g[gridDimX-1][0] = "b"

	n, sMax := neighbors(g, 0, 0)

	if n != 3 {
		t.Errorf("Expected number of neighbors to be 1 but got %v", n)
	}
	if sMax != "b" {
		t.Errorf("Expected most populous species to be \"b\" but got %q", sMax)
	}
}

func Test_nextState(t *testing.T) {
	g, df := &grid{}, make(diff)
	g[10][10] = "a"
	g[10][11] = "b"
	g[11][11] = "a"
	g[11][12] = "b"
	g[12][11] = "c"

	nextState(g, df)

	if n := len(df[10]); n < 2 || n > 3 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[10][10]; ok && v != "b" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[10][11]; !ok || v != "a" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[10][12]; !ok || v != "b" {
		t.Errorf("Incorrect game state")
	}
	delete(df, 10)

	if n := len(df[11]); n < 1 || n > 2 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[11][11]; !ok || v != "" {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[11][12]; ok && v != "a" && v != "c" {
		t.Errorf("Incorrect game state")
	}
	delete(df, 11)

	if n := len(df[12]); n != 2 {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[12][11]; !ok || (v != "a" && v != "b") {
		t.Errorf("Incorrect game state")
	}
	if v, ok := df[12][12]; !ok || (v != "a" && v != "b" && v != "c") {
		t.Errorf("Incorrect game state")
	}
	delete(df, 12)

	if len(df) != 0 {
		t.Errorf("Incorrect game state")
	}
}
