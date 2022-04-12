package main

import (
	"errors"
)

func validateDiff(diff Diff) error {
	if len(diff) == 0 {
		return errors.New("Unmarshalled diff is empty.")
	}
	for x := range diff {
		if x >= GridDimX {
			return errors.New("Unmarshalled diff exceeds grid's X dimension.")
		}
		if len(diff[x]) == 0 {
			return errors.New("Unmarshalled diff includes an X coordinate with no Y coordinate.")
		}
		for y := range diff[x] {
			if y >= GridDimY {
				return errors.New("Unmarshalled diff exceeds grid's Y dimension.")
			}
		}
	}
	return nil
}
