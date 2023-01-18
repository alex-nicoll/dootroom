package main

import (
	"errors"
	"fmt"
	"regexp"
)

var hexColorCode = regexp.MustCompile(`\A#[0-9a-f]{6}\z`)

func validateDiff(df diff) error {
	if len(df) == 0 {
		return errors.New("diff is empty")
	}
	for x := range df {
		if x >= gridDimX {
			return errors.New("diff exceeds grid's X dimension")
		}
		ydiff := df[x]
		if len(ydiff) == 0 {
			return errors.New("diff includes an X coordinate with no Y coordinate")
		}
		for y, v := range ydiff {
			if y >= gridDimY {
				return errors.New("diff exceeds grid's Y dimension")
			}
			if !hexColorCode.MatchString(v) {
				return fmt.Errorf("diff contains a cell value that is not a "+
					"hexadecimal color code (%v)", v)
			}
		}
	}
	return nil
}
