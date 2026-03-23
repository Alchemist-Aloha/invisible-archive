package vfs

import (
	"strconv"
	"unicode"
)

// NaturalCompare returns true if s1 < s2 in natural sort order.
func NaturalCompare(s1, s2 string) bool {
	i, j := 0, 0
	for i < len(s1) && j < len(s2) {
		c1, c2 := rune(s1[i]), rune(s2[j])
		if unicode.IsDigit(c1) && unicode.IsDigit(c2) {
			// Find numeric segments
			n1Start := i
			for i < len(s1) && unicode.IsDigit(rune(s1[i])) {
				i++
			}
			n1, _ := strconv.Atoi(s1[n1Start:i])

			n2Start := j
			for j < len(s2) && unicode.IsDigit(rune(s2[j])) {
				j++
			}
			n2, _ := strconv.Atoi(s2[n2Start:j])

			if n1 != n2 {
				return n1 < n2
			}
			continue
		}
		if c1 != c2 {
			return c1 < c2
		}
		i++
		j++
	}
	return len(s1) < len(s2)
}
