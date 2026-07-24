package engine

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
)

type Technique uint8

const (
	NakedSingle Technique = iota + 1
	HiddenSingle
	LockedCandidates
	NakedPair
	HiddenPair
	NakedTriple
	HiddenTriple
	XWing
	Swordfish
	XYWing
	SimpleColoring
)

func (technique Technique) String() string {
	names := [...]string{"", "naked_single", "hidden_single", "locked_candidates", "naked_pair", "hidden_pair", "naked_triple", "hidden_triple", "x_wing", "swordfish", "xy_wing", "simple_coloring"}
	if int(technique) >= len(names) {
		return "unknown"
	}
	return names[technique]
}

type Grade string

const (
	Easy   Grade = "Easy"
	Medium Grade = "Medium"
	Hard   Grade = "Hard"
	Expert Grade = "Expert"
)

type StepKind string

const (
	Placement   StepKind = "placement"
	Elimination StepKind = "elimination"
)

type CandidateRemoval struct {
	Cell  int
	Digit uint8
}

type Step struct {
	Kind          StepKind
	Technique     Technique
	Cell          int
	Digit         uint8
	Eliminations  []CandidateRemoval
	UnitKind      string
	UnitIndex     int
	AffectedCells []int
}

type SolveResult struct {
	Solution         Solution
	Steps            []Step
	HardestTechnique Technique
	Grade            Grade
}

type logicalState struct {
	values     [CellCount]uint8
	candidates [CellCount]uint16
}

var units = buildUnits()
var peers = buildPeers()

func SolveLogical(clues Clues) (SolveResult, error) {
	state, err := newLogicalState(clues)
	if err != nil {
		return SolveResult{}, err
	}
	var steps []Step
	var hardest Technique
	for !complete(state.values) {
		step, ok := nextStep(&state)
		if !ok {
			return SolveResult{}, errors.New("logical solver stalled after simple coloring")
		}
		if step.Technique > hardest {
			hardest = step.Technique
		}
		applyStep(&state, step)
		steps = append(steps, step)
	}
	if !validComplete(state.values) {
		return SolveResult{}, errors.New("logical solver produced an invalid grid")
	}
	var solution Solution
	solution.cells = state.values
	return SolveResult{Solution: solution, Steps: steps, HardestTechnique: hardest, Grade: gradeFor(hardest)}, nil
}

func newLogicalState(clues Clues) (logicalState, error) {
	state := logicalState{values: clues.cells}
	for index, value := range state.values {
		if value == 0 {
			state.candidates[index] = candidatesFor(state.values, index)
			if state.candidates[index] == 0 {
				return logicalState{}, fmt.Errorf("cell %d has no candidate", index)
			}
		}
	}
	return state, nil
}

func nextStep(state *logicalState) (Step, bool) {
	finders := []func(*logicalState) (Step, bool){
		findNakedSingle,
		findHiddenSingle,
		findLockedCandidates,
		func(state *logicalState) (Step, bool) { return findNakedSubset(state, 2, NakedPair) },
		func(state *logicalState) (Step, bool) { return findHiddenSubset(state, 2, HiddenPair) },
		func(state *logicalState) (Step, bool) { return findNakedSubset(state, 3, NakedTriple) },
		func(state *logicalState) (Step, bool) { return findHiddenSubset(state, 3, HiddenTriple) },
		func(state *logicalState) (Step, bool) { return findFish(state, 2, XWing) },
		func(state *logicalState) (Step, bool) { return findFish(state, 3, Swordfish) },
		findXYWing,
		findSimpleColoring,
	}
	for _, finder := range finders {
		if step, ok := finder(state); ok {
			return step, true
		}
	}
	return Step{}, false
}

func findNakedSingle(state *logicalState) (Step, bool) {
	for cell, mask := range state.candidates {
		if state.values[cell] == 0 && bits.OnesCount16(mask) == 1 {
			return Step{Kind: Placement, Technique: NakedSingle, Cell: cell, Digit: firstDigit(mask), AffectedCells: []int{cell}}, true
		}
	}
	return Step{}, false
}

func findHiddenSingle(state *logicalState) (Step, bool) {
	for unitIndex, unit := range units {
		for digit := uint8(1); digit <= 9; digit++ {
			cell := -1
			count := 0
			for _, candidate := range unit.cells {
				if state.values[candidate] == 0 && state.candidates[candidate]&(1<<digit) != 0 {
					cell, count = candidate, count+1
				}
			}
			if count == 1 {
				return Step{Kind: Placement, Technique: HiddenSingle, Cell: cell, Digit: digit, UnitKind: unit.kind, UnitIndex: unit.index, AffectedCells: []int{cell}}, true
			}
		}
		_ = unitIndex
	}
	return Step{}, false
}

func findLockedCandidates(state *logicalState) (Step, bool) {
	for box := 0; box < 9; box++ {
		unit := units[18+box]
		for digit := uint8(1); digit <= 9; digit++ {
			var cells []int
			for _, cell := range unit.cells {
				if state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
					cells = append(cells, cell)
				}
			}
			if len(cells) < 2 {
				continue
			}
			if sameRow(cells) {
				if removals := removeFromUnitOutside(state, units[cells[0]/9].cells, cells, digit); len(removals) > 0 {
					return eliminationStep(LockedCandidates, removals, "box", box, cells), true
				}
			}
			if sameColumn(cells) {
				if removals := removeFromUnitOutside(state, units[9+cells[0]%9].cells, cells, digit); len(removals) > 0 {
					return eliminationStep(LockedCandidates, removals, "box", box, cells), true
				}
			}
		}
	}
	for line := 0; line < 18; line++ {
		unit := units[line]
		for digit := uint8(1); digit <= 9; digit++ {
			var cells []int
			for _, cell := range unit.cells {
				if state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
					cells = append(cells, cell)
				}
			}
			if len(cells) >= 2 && sameBox(cells) {
				box := (cells[0]/9/3)*3 + cells[0]%9/3
				if removals := removeFromUnitOutside(state, units[18+box].cells, cells, digit); len(removals) > 0 {
					return eliminationStep(LockedCandidates, removals, unit.kind, unit.index, cells), true
				}
			}
		}
	}
	return Step{}, false
}

func findNakedSubset(state *logicalState, size int, technique Technique) (Step, bool) {
	for _, unit := range units {
		var eligible []int
		for _, cell := range unit.cells {
			count := bits.OnesCount16(state.candidates[cell])
			if state.values[cell] == 0 && count >= 2 && count <= size {
				eligible = append(eligible, cell)
			}
		}
		for _, combo := range combinations(eligible, size) {
			var union uint16
			for _, cell := range combo {
				union |= state.candidates[cell]
			}
			if bits.OnesCount16(union) != size {
				continue
			}
			var removals []CandidateRemoval
			for _, cell := range unit.cells {
				if containsInt(combo, cell) || state.values[cell] != 0 {
					continue
				}
				for digit := uint8(1); digit <= 9; digit++ {
					if union&(1<<digit) != 0 && state.candidates[cell]&(1<<digit) != 0 {
						removals = append(removals, CandidateRemoval{cell, digit})
					}
				}
			}
			if len(removals) > 0 {
				return eliminationStep(technique, removals, unit.kind, unit.index, combo), true
			}
		}
	}
	return Step{}, false
}

func findHiddenSubset(state *logicalState, size int, technique Technique) (Step, bool) {
	digits := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, unit := range units {
		for _, digitCombo := range combinations(digits, size) {
			var selectedMask uint16
			var cells []int
			for _, digit := range digitCombo {
				selectedMask |= 1 << digit
			}
			for _, cell := range unit.cells {
				if state.values[cell] == 0 && state.candidates[cell]&selectedMask != 0 {
					cells = append(cells, cell)
				}
			}
			if len(cells) != size {
				continue
			}
			allPresent := true
			for _, digit := range digitCombo {
				present := false
				for _, cell := range cells {
					present = present || state.candidates[cell]&(1<<digit) != 0
				}
				allPresent = allPresent && present
			}
			if !allPresent {
				continue
			}
			var removals []CandidateRemoval
			for _, cell := range cells {
				for digit := uint8(1); digit <= 9; digit++ {
					if selectedMask&(1<<digit) == 0 && state.candidates[cell]&(1<<digit) != 0 {
						removals = append(removals, CandidateRemoval{cell, digit})
					}
				}
			}
			if len(removals) > 0 {
				return eliminationStep(technique, removals, unit.kind, unit.index, cells), true
			}
		}
	}
	return Step{}, false
}

func findFish(state *logicalState, size int, technique Technique) (Step, bool) {
	for orientation := 0; orientation < 2; orientation++ {
		for digit := uint8(1); digit <= 9; digit++ {
			var bases []int
			masks := [9]uint16{}
			for base := 0; base < 9; base++ {
				for cover := 0; cover < 9; cover++ {
					cell := base*9 + cover
					if orientation == 1 {
						cell = cover*9 + base
					}
					if state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
						masks[base] |= 1 << cover
					}
				}
				count := bits.OnesCount16(masks[base])
				if count >= 2 && count <= size {
					bases = append(bases, base)
				}
			}
			for _, combo := range combinations(bases, size) {
				var covers uint16
				for _, base := range combo {
					covers |= masks[base]
				}
				if bits.OnesCount16(covers) != size {
					continue
				}
				var removals []CandidateRemoval
				for base := 0; base < 9; base++ {
					if containsInt(combo, base) {
						continue
					}
					for cover := 0; cover < 9; cover++ {
						if covers&(1<<cover) == 0 {
							continue
						}
						cell := base*9 + cover
						if orientation == 1 {
							cell = cover*9 + base
						}
						if state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
							removals = append(removals, CandidateRemoval{cell, digit})
						}
					}
				}
				if len(removals) > 0 {
					kind := "rows"
					if orientation == 1 {
						kind = "columns"
					}
					return eliminationStep(technique, removals, kind, -1, combo), true
				}
			}
		}
	}
	return Step{}, false
}

func findXYWing(state *logicalState) (Step, bool) {
	for pivot := 0; pivot < CellCount; pivot++ {
		pivotMask := state.candidates[pivot]
		if state.values[pivot] != 0 || bits.OnesCount16(pivotMask) != 2 {
			continue
		}
		for _, first := range peers[pivot] {
			firstMask := state.candidates[first]
			if state.values[first] != 0 || bits.OnesCount16(firstMask) != 2 || bits.OnesCount16(firstMask&pivotMask) != 1 {
				continue
			}
			zMask := firstMask &^ pivotMask
			for _, second := range peers[pivot] {
				if second <= first {
					continue
				}
				secondMask := state.candidates[second]
				if state.values[second] != 0 || bits.OnesCount16(secondMask) != 2 ||
					bits.OnesCount16(secondMask&pivotMask) != 1 || secondMask&pivotMask == firstMask&pivotMask ||
					secondMask&^pivotMask != zMask {
					continue
				}
				z := firstDigit(zMask)
				var removals []CandidateRemoval
				for cell := 0; cell < CellCount; cell++ {
					if cell != pivot && cell != first && cell != second && isPeer(cell, first) && isPeer(cell, second) &&
						state.values[cell] == 0 && state.candidates[cell]&(1<<z) != 0 {
						removals = append(removals, CandidateRemoval{cell, z})
					}
				}
				if len(removals) > 0 {
					return eliminationStep(XYWing, removals, "", -1, []int{pivot, first, second}), true
				}
			}
		}
	}
	return Step{}, false
}

func findSimpleColoring(state *logicalState) (Step, bool) {
	for digit := uint8(1); digit <= 9; digit++ {
		graph := make([][]int, CellCount)
		for _, unit := range units {
			var pair []int
			for _, cell := range unit.cells {
				if state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
					pair = append(pair, cell)
				}
			}
			if len(pair) == 2 {
				graph[pair[0]] = appendUnique(graph[pair[0]], pair[1])
				graph[pair[1]] = appendUnique(graph[pair[1]], pair[0])
			}
		}
		colors := [CellCount]int8{}
		for start := 0; start < CellCount; start++ {
			if len(graph[start]) == 0 || colors[start] != 0 {
				continue
			}
			colors[start] = 1
			queue := []int{start}
			var component []int
			for len(queue) > 0 {
				cell := queue[0]
				queue = queue[1:]
				component = append(component, cell)
				sort.Ints(graph[cell])
				for _, neighbor := range graph[cell] {
					if colors[neighbor] == 0 {
						colors[neighbor] = -colors[cell]
						queue = append(queue, neighbor)
					}
				}
			}
			for _, color := range []int8{1, -1} {
				conflict := false
				for i, first := range component {
					if colors[first] != color {
						continue
					}
					for _, second := range component[i+1:] {
						if colors[second] == color && isPeer(first, second) {
							conflict = true
						}
					}
				}
				if conflict {
					var removals []CandidateRemoval
					for _, cell := range component {
						if colors[cell] == color {
							removals = append(removals, CandidateRemoval{cell, digit})
						}
					}
					return eliminationStep(SimpleColoring, removals, "", -1, component), true
				}
			}
			for cell := 0; cell < CellCount; cell++ {
				if colors[cell] != 0 || state.values[cell] != 0 || state.candidates[cell]&(1<<digit) == 0 {
					continue
				}
				seesPositive, seesNegative := false, false
				for _, colored := range component {
					if isPeer(cell, colored) {
						seesPositive = seesPositive || colors[colored] == 1
						seesNegative = seesNegative || colors[colored] == -1
					}
				}
				if seesPositive && seesNegative {
					return eliminationStep(SimpleColoring, []CandidateRemoval{{cell, digit}}, "", -1, component), true
				}
			}
		}
	}
	return Step{}, false
}

func applyStep(state *logicalState, step Step) {
	if step.Kind == Placement {
		state.values[step.Cell] = step.Digit
		state.candidates[step.Cell] = 0
		for _, peer := range peers[step.Cell] {
			state.candidates[peer] &^= 1 << step.Digit
		}
		return
	}
	for _, removal := range step.Eliminations {
		state.candidates[removal.Cell] &^= 1 << removal.Digit
	}
}

func gradeFor(technique Technique) Grade {
	switch {
	case technique <= HiddenSingle:
		return Easy
	case technique <= HiddenPair:
		return Medium
	case technique <= XWing:
		return Hard
	default:
		return Expert
	}
}

func complete(values [CellCount]uint8) bool {
	for _, value := range values {
		if value == 0 {
			return false
		}
	}
	return true
}

type unit struct {
	kind  string
	index int
	cells []int
}

func buildUnits() []unit {
	result := make([]unit, 0, 27)
	for row := 0; row < 9; row++ {
		cells := make([]int, 9)
		for column := 0; column < 9; column++ {
			cells[column] = row*9 + column
		}
		result = append(result, unit{"row", row, cells})
	}
	for column := 0; column < 9; column++ {
		cells := make([]int, 9)
		for row := 0; row < 9; row++ {
			cells[row] = row*9 + column
		}
		result = append(result, unit{"column", column, cells})
	}
	for box := 0; box < 9; box++ {
		var cells []int
		for offset := 0; offset < 9; offset++ {
			cells = append(cells, (box/3*3+offset/3)*9+box%3*3+offset%3)
		}
		result = append(result, unit{"box", box, cells})
	}
	return result
}

func buildPeers() [CellCount][]int {
	var result [CellCount][]int
	for cell := 0; cell < CellCount; cell++ {
		for other := 0; other < CellCount; other++ {
			if cell != other && (cell/9 == other/9 || cell%9 == other%9 || (cell/9/3 == other/9/3 && cell%9/3 == other%9/3)) {
				result[cell] = append(result[cell], other)
			}
		}
	}
	return result
}

func eliminationStep(technique Technique, removals []CandidateRemoval, unitKind string, unitIndex int, affected []int) Step {
	sort.Slice(removals, func(i, j int) bool {
		if removals[i].Cell == removals[j].Cell {
			return removals[i].Digit < removals[j].Digit
		}
		return removals[i].Cell < removals[j].Cell
	})
	return Step{Kind: Elimination, Technique: technique, Eliminations: removals, UnitKind: unitKind, UnitIndex: unitIndex, AffectedCells: append([]int(nil), affected...)}
}

func combinations(values []int, size int) [][]int {
	var result [][]int
	var visit func(int, []int)
	visit = func(start int, selected []int) {
		if len(selected) == size {
			result = append(result, append([]int(nil), selected...))
			return
		}
		for index := start; index <= len(values)-(size-len(selected)); index++ {
			visit(index+1, append(selected, values[index]))
		}
	}
	visit(0, nil)
	return result
}

func firstDigit(mask uint16) uint8 { return uint8(bits.TrailingZeros16(mask)) }
func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func sameRow(cells []int) bool {
	for _, cell := range cells[1:] {
		if cell/9 != cells[0]/9 {
			return false
		}
	}
	return true
}
func sameColumn(cells []int) bool {
	for _, cell := range cells[1:] {
		if cell%9 != cells[0]%9 {
			return false
		}
	}
	return true
}
func sameBox(cells []int) bool {
	box := cells[0]/9/3*3 + cells[0]%9/3
	for _, cell := range cells[1:] {
		if cell/9/3*3+cell%9/3 != box {
			return false
		}
	}
	return true
}
func removeFromUnitOutside(state *logicalState, unitCells, excluded []int, digit uint8) []CandidateRemoval {
	var removals []CandidateRemoval
	for _, cell := range unitCells {
		if !containsInt(excluded, cell) && state.values[cell] == 0 && state.candidates[cell]&(1<<digit) != 0 {
			removals = append(removals, CandidateRemoval{cell, digit})
		}
	}
	return removals
}
func isPeer(first, second int) bool {
	return first != second && (first/9 == second/9 || first%9 == second%9 || (first/9/3 == second/9/3 && first%9/3 == second%9/3))
}
func appendUnique(values []int, value int) []int {
	if containsInt(values, value) {
		return values
	}
	return append(values, value)
}
