package main

import (
	"errors"
)

func validateDiff(diff Diff) error {
	if len(diff) == 0 {
		return errors.New("Diff is empty.")
	}
	for x := range diff {
		if x >= GridDimX {
			return errors.New("Diff exceeds grid's X dimension.")
		}
		if len(diff[x]) == 0 {
			return errors.New("Diff includes an X coordinate with no Y coordinate.")
		}
		for y := range diff[x] {
			if y >= GridDimY {
				return errors.New("Diff exceeds grid's Y dimension.")
			}
		}
	}
	return nil
}
