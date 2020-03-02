package internal

import "fmt"

type Cell struct {
	row   int
	col   int
	value int
	moves Moves
}

func (c *Cell) Set(value int) error {
	if value < 1 || value > 9 {
		return fmt.Errorf("Cell value out of range: %s", value)
	}

	if c.value != 0 {
		return fmt.Errorf("Cell already set to %d", c.value)
	}

	c.value = value
	c.moves = empty

	return nil
}

func (c *Cell) CanPlay(value int) bool {
	return c.moves.Contains(value)
}

func (c *Cell) EliminateMove(value int) bool {
	if value < 1 || value > 9 {
		panic(fmt.Errorf("Value out of range: %s", value))
	}

	return c.moves.Remove(value)
}

func (c *Cell) Moves() []int {
	return c.moves.Slice()
}

type Cells []*Cell

func (c Cells) FindMove(value int) Cells {
	cells := make(Cells, 0)
	for _, cell := range c {
		if cell.CanPlay(value) {
			cells = append(cells, cell)
		}
	}
	return cells
}

func (c Cells) UnsetOnly() Cells {
	cells := make(Cells, 0)
	for _, cell := range c {
		if cell.value == 0 {
			cells = append(cells, cell)
		}
	}
	return cells
}

func (c Cells) Excluding(other Cells) Cells {
	difference := make(Cells, 0)

	for _, candidate := range c {
		excluded := false
		for _, other := range other {
			if candidate.row == other.row && candidate.col == other.col {
				excluded = true
				break
			}
		}

		if !excluded {
			difference = append(difference, candidate)
		}
	}

	return difference
}

func (c Cells) LocationString() string {
	s := ""
	for i, cell := range c {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("(%d,%d)", cell.row+1, cell.col+1)
	}
	return s
}

func (c Cells) PowerSet() []Cells {
	if len(c) == 0 {
		return []Cells{Cells{}}
	}

	subsets := c[1:].PowerSet()
	sets := make([]Cells, 0, 2<<len(c))

	sets = append(sets, subsets...)

	head := c[0:1]
	for _, tail := range subsets {
		subset := make(Cells, 0, 1+len(tail))
		subset = append(subset, head...)
		subset = append(subset, tail...)
		sets = append(sets, subset)
	}

	return sets
}

func (c Cells) RemainingMoves() *Moves {
	moves := empty
	for _, cell := range c {
		moves = moves | cell.moves
	}
	return &moves
}

func (c Cells) EliminateMove(value int) int {
	changes := 0
	for _, cell := range c {
		if cell.EliminateMove(value) {
			changes++
		}
	}
	return changes
}

func (c Cells) UniqueRows() []int {
	rowsPresent := [9]bool{}
	for _, cell := range c {
		rowsPresent[cell.row] = true
	}

	rows := make([]int, 0)
	for row := 0; row < 9; row++ {
		if rowsPresent[row] {
			rows = append(rows, row)
		}
	}

	return rows
}

func (c Cells) UniqueCols() []int {
	colsPresent := [9]bool{}
	for _, cell := range c {
		colsPresent[cell.col] = true
	}

	cols := make([]int, 0)
	for col := 0; col < 9; col++ {
		if colsPresent[col] {
			cols = append(cols, col)
		}
	}

	return cols
}
