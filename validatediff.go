package main

import (
	"errors"
	"fmt"
	"regexp"
)

var hexColorCode *regexp.Regexp = regexp.MustCompile("\\A#[0-9a-f]{6}\\z")

func validateDiff(diff Diff) error {
	if len(diff) == 0 {
		return errors.New("Diff is empty.")
	}
	for x := range diff {
		if x >= GridDimX {
			return errors.New("Diff exceeds grid's X dimension.")
		}
		ydiff := diff[x]
		if len(ydiff) == 0 {
			return errors.New("Diff includes an X coordinate with no Y coordinate.")
		}
		for y, v := range ydiff {
			if y >= GridDimY {
				return errors.New("Diff exceeds grid's Y dimension.")
			}
			if !hexColorCode.MatchString(v) {
				return errors.New(fmt.Sprintf("Diff contains a cell value that is "+
					"not a hexadecimal color code. (%v)", v))
			}
		}
	}
	return nil
}
