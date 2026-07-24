package engine

import (
	"crypto/sha256"
	"encoding/hex"
)

const CanonicalVersion = "canonical-v1"

var groupedPermutations = allGroupedPermutations()

func CanonicalFingerprint(clues Clues) string {
	var best [CellCount]byte
	hasBest := false
	for transpose := 0; transpose < 2; transpose++ {
		for _, rows := range groupedPermutations {
			for _, columns := range groupedPermutations {
				var candidate [CellCount]byte
				var mapping [10]byte
				next := byte(1)
				for target := 0; target < CellCount; target++ {
					row, column := target/9, target%9
					sourceRow, sourceColumn := rows[row], columns[column]
					if transpose == 1 {
						sourceRow, sourceColumn = columns[column], rows[row]
					}
					value := clues.cells[sourceRow*9+sourceColumn]
					if value != 0 {
						if mapping[value] == 0 {
							mapping[value], next = next, next+1
						}
						value = mapping[value]
					}
					candidate[target] = '0' + value
				}
				if !hasBest || lessGrid(candidate, best) {
					best, hasBest = candidate, true
				}
			}
		}
	}
	digest := sha256.Sum256(best[:])
	return hex.EncodeToString(digest[:])
}

func allGroupedPermutations() [][]int {
	base := permutations3()
	result := make([][]int, 0, 1296)
	for _, groups := range base {
		for _, first := range base {
			for _, second := range base {
				for _, third := range base {
					within := [][]int{first, second, third}
					permutation := make([]int, 0, 9)
					for _, group := range groups {
						for _, offset := range within[group] {
							permutation = append(permutation, group*3+offset)
						}
					}
					result = append(result, permutation)
				}
			}
		}
	}
	return result
}

func permutations3() [][]int {
	return [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
}

func lessGrid(first, second [CellCount]byte) bool {
	for index := 0; index < CellCount; index++ {
		if first[index] != second[index] {
			return first[index] < second[index]
		}
	}
	return false
}
