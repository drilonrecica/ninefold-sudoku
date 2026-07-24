package engine

import (
	"errors"
	"fmt"
	"strings"
)

const CellCount = 81

type Clues struct{ cells [CellCount]uint8 }
type Solution struct{ cells [CellCount]uint8 }

func ParseClues(encoded string) (Clues, error) {
	var clues Clues
	if len(encoded) != CellCount {
		return clues, fmt.Errorf("clues must contain exactly %d ASCII cells", CellCount)
	}
	for index, char := range []byte(encoded) {
		switch {
		case char == '.' || char == '0':
			clues.cells[index] = 0
		case char >= '1' && char <= '9':
			clues.cells[index] = char - '0'
		default:
			return Clues{}, fmt.Errorf("clues contain invalid cell at index %d", index)
		}
	}
	if !validPartial(clues.cells) {
		return Clues{}, errors.New("clues contain a row, column, or box conflict")
	}
	return clues, nil
}

func ParseSolution(encoded string) (Solution, error) {
	var solution Solution
	if len(encoded) != CellCount {
		return solution, fmt.Errorf("solution must contain exactly %d ASCII digits", CellCount)
	}
	for index, char := range []byte(encoded) {
		if char < '1' || char > '9' {
			return Solution{}, fmt.Errorf("solution contains invalid digit at index %d", index)
		}
		solution.cells[index] = char - '0'
	}
	if !validComplete(solution.cells) {
		return Solution{}, errors.New("solution is not a complete valid Sudoku grid")
	}
	return solution, nil
}

func NewPuzzle(clues Clues, solution Solution) error {
	for index, clue := range clues.cells {
		if clue != 0 && clue != solution.cells[index] {
			return fmt.Errorf("clue disagrees with solution at index %d", index)
		}
	}
	return nil
}

func (clues Clues) String() string {
	var builder strings.Builder
	builder.Grow(CellCount)
	for _, value := range clues.cells {
		builder.WriteByte('0' + value)
	}
	return builder.String()
}

func (solution Solution) String() string {
	var builder strings.Builder
	builder.Grow(CellCount)
	for _, value := range solution.cells {
		builder.WriteByte('0' + value)
	}
	return builder.String()
}

func (clues Clues) Value(index int) (uint8, bool) {
	if index < 0 || index >= CellCount {
		return 0, false
	}
	return clues.cells[index], clues.cells[index] != 0
}

func (solution Solution) Value(index int) (uint8, bool) {
	if index < 0 || index >= CellCount {
		return 0, false
	}
	return solution.cells[index], true
}

func validPartial(cells [CellCount]uint8) bool {
	var rows, columns, boxes [9]uint16
	for index, value := range cells {
		if value == 0 {
			continue
		}
		if value > 9 {
			return false
		}
		bit := uint16(1 << value)
		row, column := index/9, index%9
		box := (row/3)*3 + column/3
		if rows[row]&bit != 0 || columns[column]&bit != 0 || boxes[box]&bit != 0 {
			return false
		}
		rows[row] |= bit
		columns[column] |= bit
		boxes[box] |= bit
	}
	return true
}

func validComplete(cells [CellCount]uint8) bool {
	if !validPartial(cells) {
		return false
	}
	for _, value := range cells {
		if value == 0 {
			return false
		}
	}
	return true
}
