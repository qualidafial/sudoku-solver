package internal

import "fmt"

const (
	empty Moves = 0
	full  Moves = 0b111111111
)

type Moves int

func (m *Moves) Contains(value int) bool {
	return *m & mask(value) != 0
}

func (m *Moves) Add(value int) bool {
	if m.Contains(value) {
		return false
	}

	*m = *m | mask(value)

	return true
}

func (m *Moves) Remove(value int) bool {
	if !m.Contains(value) {
		return false
	}

	*m = *m & ^mask(value)

	return true
}

func (m *Moves) Slice() []int {
	moves := make([]int, 0)
	for value := 1; value <= 9; value++ {
		if m.Contains(value) {
			moves = append(moves, value)
		}
	}
	return moves
}

func mask(value int) Moves {
	if value < 1 || value > 9 {
		panic(fmt.Errorf("value out of range: %d", value))
	}
	return Moves(1 << (value - 1))
}
