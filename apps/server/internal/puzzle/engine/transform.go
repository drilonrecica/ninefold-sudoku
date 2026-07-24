package engine

import "errors"

const TransformationVersion = "splitmix64-v1"

type splitMix64 struct{ state uint64 }

func (random *splitMix64) next() uint64 {
	random.state += 0x9e3779b97f4a7c15
	value := random.state
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func Transform(clues Clues, solution Solution, seed uint64) (Clues, Solution, error) {
	if err := NewPuzzle(clues, solution); err != nil {
		return Clues{}, Solution{}, err
	}
	random := splitMix64{state: seed}
	digits := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	shuffle(digits, &random)
	rows := groupedPermutation(&random)
	columns := groupedPermutation(&random)
	symmetry := int(random.next() % 8)

	var transformedClues Clues
	var transformedSolution Solution
	for target := 0; target < CellCount; target++ {
		row, column := inverseSymmetry(target/9, target%9, symmetry)
		source := rows[row]*9 + columns[column]
		if clues.cells[source] != 0 {
			transformedClues.cells[target] = digits[clues.cells[source]-1]
		}
		transformedSolution.cells[target] = digits[solution.cells[source]-1]
	}
	if !validPartial(transformedClues.cells) || !validComplete(transformedSolution.cells) {
		return Clues{}, Solution{}, errors.New("transformation violated Sudoku invariants")
	}
	return transformedClues, transformedSolution, NewPuzzle(transformedClues, transformedSolution)
}

func groupedPermutation(random *splitMix64) []int {
	groups := []int{0, 1, 2}
	shuffle(groups, random)
	result := make([]int, 0, 9)
	for _, group := range groups {
		within := []int{0, 1, 2}
		shuffle(within, random)
		for _, offset := range within {
			result = append(result, group*3+offset)
		}
	}
	return result
}

func shuffle[T any](values []T, random *splitMix64) {
	for index := len(values) - 1; index > 0; index-- {
		swap := int(random.next() % uint64(index+1))
		values[index], values[swap] = values[swap], values[index]
	}
}

func inverseSymmetry(row, column, symmetry int) (int, int) {
	switch symmetry {
	case 0:
		return row, column
	case 1:
		return 8 - column, row
	case 2:
		return 8 - row, 8 - column
	case 3:
		return column, 8 - row
	case 4:
		return row, 8 - column
	case 5:
		return 8 - row, column
	case 6:
		return column, row
	default:
		return 8 - column, 8 - row
	}
}
