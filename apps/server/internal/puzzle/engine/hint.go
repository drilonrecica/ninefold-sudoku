package engine

import "errors"

type NudgeHint struct {
	Technique     Technique
	UnitKind      string
	UnitIndex     int
	AffectedCells []int
}

type RevealHint struct {
	Cell  int
	Value uint8
}

func Nudge(clues Clues) (NudgeHint, error) {
	state, err := newLogicalState(clues)
	if err != nil {
		return NudgeHint{}, err
	}
	step, ok := nextStep(&state)
	if !ok {
		return NudgeHint{}, errors.New("no deterministic logical hint is available")
	}
	cells := append([]int(nil), step.AffectedCells...)
	if len(cells) == 0 && step.Cell >= 0 {
		cells = []int{step.Cell}
	}
	return NudgeHint{
		Technique:     step.Technique,
		UnitKind:      step.UnitKind,
		UnitIndex:     step.UnitIndex,
		AffectedCells: cells,
	}, nil
}

func Reveal(clues Clues, solution Solution) (RevealHint, error) {
	if err := NewPuzzle(clues, solution); err != nil {
		return RevealHint{}, err
	}
	state, err := newLogicalState(clues)
	if err != nil {
		return RevealHint{}, err
	}
	for !complete(state.values) {
		step, ok := nextStep(&state)
		if !ok {
			return RevealHint{}, errors.New("no deterministic logical reveal is available")
		}
		applyStep(&state, step)
		if step.Kind == Placement {
			value := solution.cells[step.Cell]
			if clues.cells[step.Cell] != 0 || value == 0 || value != step.Digit {
				return RevealHint{}, errors.New("logical reveal disagrees with the verified solution")
			}
			return RevealHint{Cell: step.Cell, Value: value}, nil
		}
	}
	return RevealHint{}, errors.New("puzzle is already complete")
}
