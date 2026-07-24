package engine

import "math/bits"

const fullMask uint16 = 0x3fe

func Candidates(clues Clues, index int) uint16 {
	if index < 0 || index >= CellCount || clues.cells[index] != 0 {
		return 0
	}
	return candidatesFor(clues.cells, index)
}

func candidatesFor(cells [CellCount]uint8, index int) uint16 {
	row, column := index/9, index%9
	var used uint16
	for offset := 0; offset < 9; offset++ {
		used |= 1 << cells[row*9+offset]
		used |= 1 << cells[offset*9+column]
		boxIndex := (row/3*3+offset/3)*9 + column/3*3 + offset%3
		used |= 1 << cells[boxIndex]
	}
	return fullMask &^ used
}

func CountSolutions(clues Clues) (int, Solution) {
	cells := clues.cells
	count := 0
	var first Solution
	search(&cells, &count, &first)
	return count, first
}

func search(cells *[CellCount]uint8, count *int, first *Solution) {
	if *count >= 2 {
		return
	}
	bestIndex, bestMask, bestCount := -1, uint16(0), 10
	for index, value := range cells {
		if value != 0 {
			continue
		}
		mask := candidatesFor(*cells, index)
		size := bits.OnesCount16(mask)
		if size == 0 {
			return
		}
		if size < bestCount {
			bestIndex, bestMask, bestCount = index, mask, size
			if size == 1 {
				break
			}
		}
	}
	if bestIndex == -1 {
		*count++
		if *count == 1 {
			first.cells = *cells
		}
		return
	}
	for digit := uint8(1); digit <= 9 && *count < 2; digit++ {
		if bestMask&(1<<digit) == 0 {
			continue
		}
		cells[bestIndex] = digit
		search(cells, count, first)
		cells[bestIndex] = 0
	}
}
