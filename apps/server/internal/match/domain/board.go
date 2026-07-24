package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

// cellPeers returns the indices of all cells in the same row, column, and 3x3 box as i.
func cellPeers(i shared.CellIndex) []shared.CellIndex {
	row := int(i) / 9
	col := int(i) % 9
	boxRow := row / 3
	boxCol := col / 3
	var peers []shared.CellIndex
	for c := 0; c < 9; c++ {
		if c != col {
			peers = append(peers, shared.CellIndex(row*9+c))
		}
	}
	for r := 0; r < 9; r++ {
		if r != row {
			peers = append(peers, shared.CellIndex(r*9+col))
		}
	}
	for r := boxRow * 3; r < boxRow*3+3; r++ {
		for c := boxCol * 3; c < boxCol*3+3; c++ {
			idx := r*9 + c
			if idx != int(i) {
				peers = append(peers, shared.CellIndex(idx))
			}
		}
	}
	return peers
}

// hasDirectConflict reports whether placing value at index would conflict with an existing value.
func hasDirectConflict(values map[shared.CellIndex]shared.Digit, index shared.CellIndex, value shared.Digit) bool {
	for _, peer := range cellPeers(index) {
		if v, ok := values[peer]; ok && v == value {
			return true
		}
	}
	return false
}

// boardComplete reports whether all 81 cells have a value.
func boardComplete(values map[shared.CellIndex]shared.Digit, clues []byte) bool {
	for i := 0; i < 81; i++ {
		idx := shared.CellIndex(i)
		if clues[i] != 0 {
			continue
		}
		if _, ok := values[idx]; !ok {
			return false
		}
	}
	return true
}

// boardCorrect reports whether every placed value matches the solution.
func boardCorrect(values map[shared.CellIndex]shared.Digit, clues, solution []byte) bool {
	for i := 0; i < 81; i++ {
		idx := shared.CellIndex(i)
		if clues[i] != 0 {
			continue
		}
		v, ok := values[idx]
		if !ok {
			return false
		}
		if byte(v) != solution[i] {
			return false
		}
	}
	return true
}

// emptyCells returns the indices of all non-clue cells without a value.
func emptyCells(values map[shared.CellIndex]shared.Digit, clues []byte) []shared.CellIndex {
	var out []shared.CellIndex
	for i := 0; i < 81; i++ {
		idx := shared.CellIndex(i)
		if clues[i] == 0 {
			if _, ok := values[idx]; !ok {
				out = append(out, idx)
			}
		}
	}
	return out
}
