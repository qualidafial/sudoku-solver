package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	rowPositionNames = []string{"top", "center", "bottom"}
	colPositionNames = []string{"left", "center", "right"}
	positionNames    = [][]string{
		{"top left", "top center", "top right"},
		{"center left", "center", "center right"},
		{"bottom left", "bottom center", "bottom right"},
	}

	nondigits = regexp.MustCompile(`[^1-9 ]`)
)

type Sudoku struct {
	board [9][9]Cell
}

func NewSudokuFromReader(reader io.Reader) (*Sudoku, error) {
	var b = [9][9]int{}
	row := 0
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() && row < 9 {
		line := nondigits.ReplaceAllString(scanner.Text(), "")
		if len(line) > 0 {
			for col := 0; col < 9 && col < len(line); col++ {
				if line[col] != ' ' {
					b[row][col] = int(line[col]) - '0'
				}
			}
			row++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	s, err := NewSudoku(b)

	fmt.Println("Initialized Board")
	s.PrintBoard()

	return s, err
}

func NewSudokuFromString(board string) (*Sudoku, error) {
	return NewSudokuFromReader(strings.NewReader(board))
}

func NewSudokuFromFile(path string) (*Sudoku, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewSudokuFromReader(f)
}

func NewSudoku(board [9][9]int) (*Sudoku, error) {
	s := &Sudoku{}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			s.board[row][col] = Cell{
				row:   row,
				col:   col,
				value: 0,
				moves: full,
			}
		}
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			value := board[row][col]
			if value != 0 {
				if err := s.PlayMove(row, col, value); err != nil {
					return s, err
				}
			}
		}
	}

	return s, nil
}

func (s *Sudoku) Cells() Cells {
	return s.Range(0, 0, 8, 8)
}

func (s *Sudoku) Row(row int) Cells {
	return s.Range(row, 0, row, 8)
}

func (s *Sudoku) Col(col int) Cells {
	return s.Range(0, col, 8, col)
}

func (s *Sudoku) Cell(row, col int) *Cell {
	return &s.board[row][col]
}

func (s *Sudoku) Square(row, col int) Cells {
	return s.Range(row*3, col*3, row*3+2, col*3+2)
}

func (s *Sudoku) Range(top, left, bottom, right int) Cells {
	cells := make(Cells, 0, 9)
	for row := top; row <= bottom; row++ {
		for col := left; col <= right; col++ {
			cells = append(cells, &s.board[row][col])
		}
	}
	return cells
}

func (s *Sudoku) Rows() []Cells {
	rows := make([]Cells, 0, 9)
	for i := 0; i < 9; i++ {
		rows = append(rows, s.Row(i))
	}
	return rows
}

func (s *Sudoku) Cols() []Cells {
	cols := make([]Cells, 0, 9)
	for i := 0; i < 9; i++ {
		cols = append(cols, s.Col(i))
	}
	return cols
}

func (s *Sudoku) Squares() []Cells {
	squares := make([]Cells, 0, 9)
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			squares = append(squares, s.Square(row, col))
		}
	}
	return squares
}

func (s *Sudoku) Groups() []Cells {
	return append(append(s.Rows(), s.Cols()...), s.Squares()...)
}

func (s *Sudoku) Clone() *Sudoku {
	return &Sudoku{
		board: s.board,
	}
}

func (s *Sudoku) PlayMove(row int, col int, value int) error {
	if row < 0 || row >= 9 {
		return fmt.Errorf("Row %d out of bounds", row+1)
	}
	if col < 0 || col >= 9 {
		return fmt.Errorf("Col %d out of bounds", col+1)
	}
	if value < 1 || value > 9 {
		return fmt.Errorf("Value %d out of bounds", col+1)
	}

	if s.Cell(row, col).value != 0 {
		return fmt.Errorf("Cell %d,%d already contains %d", row+1, col+1, s.board[row][col])
	}

	if !s.Row(row).RemainingMoves().Contains(value) {
		return fmt.Errorf("Row %d already contains %d", row+1, value)
	}
	if !s.Col(col).RemainingMoves().Contains(value) {
		return fmt.Errorf("Col %d already contains %d", col+1, value)
	}
	squareRow, squareCol := row/3, col/3
	if !s.Square(squareRow, squareCol).RemainingMoves().Contains(value) {
		return fmt.Errorf("The %s square already contains %d", positionNames[squareRow][squareCol], value)
	}
	if !s.Cell(row, col).CanPlay(value) {
		return fmt.Errorf("Cell %d,%d is not a valid spot for %d", row+1, col+1, value)
	}

	err := s.Cell(row, col).Set(value)
	if err != nil {
		return err
	}
	s.Row(row).EliminateMove(value)
	s.Col(col).EliminateMove(value)
	s.Square(row/3, col/3).EliminateMove(value)

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			cell := s.Cell(row, col)
			if cell.value == 0 && cell.moves == empty {
				return fmt.Errorf("No moves left at square %d,%d", row+1, col+1)
			}
		}
	}

	return nil
}

func (s *Sudoku) PrintBoard() {
	fmt.Println()
	for row := 0; row < 9; row++ {
		if row == 3 || row == 6 {
			fmt.Println("-----+-----+-----")
		}
		for col := 0; col < 9; col++ {
			if col == 3 || col == 6 {
				fmt.Print("|")
			} else if col > 0 {
				fmt.Print(" ")
			}
			value := s.board[row][col].value
			if value > 0 {
				fmt.Print(value)
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func (s *Sudoku) PrintMoves() {
	fmt.Println()
	for row := 0; row < 9; row++ {
		if row == 3 || row == 6 {
			fmt.Println("-----------+-----------+-----------")
		} else if row > 0 {
			fmt.Println("           |           |           ")
		}

		for moveRow := 0; moveRow < 3; moveRow++ {
			for col := 0; col < 9; col++ {
				if col == 3 || col == 6 {
					fmt.Print("|")
				} else if col > 0 {
					fmt.Print(" ")
				}

				for value := moveRow*3 + 1; value < moveRow*3+4; value++ {
					if s.Cell(row, col).CanPlay(value) {
						fmt.Print(value)
					} else {
						fmt.Print(" ")
					}
				}
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

func (s *Sudoku) Solve() error {
	moves := 0

	// Cells where only a single move is possible
	for _, cell := range s.Cells() {
		possibleMoves := cell.Moves()
		if len(possibleMoves) == 1 {
			value := possibleMoves[0]
			if err := s.PlayMove(cell.row, cell.col, value); err != nil {
				return err
			}
			moves++
			log("Only %d fits in row %d column %d\n", value, cell.row+1, cell.col+1)
		}
	}

	// Squares where a number only fits in one cell
	for _, square := range s.Squares() {
		for _, value := range square.RemainingMoves().Slice() {
			cells := square.FindMove(value)
			if len(cells) == 1 {
				cell := cells[0]
				log("In the %s square, the number %d only fits in the %s cell\n", positionNames[cell.row/3][cell.col/3], value, positionNames[cell.row%3][cell.col%3])
				if err := s.PlayMove(cell.row, cell.col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// Rows where a number only fits in one cell
	for _, row := range s.Rows() {
		for _, value := range row.RemainingMoves().Slice() {
			cells := row.FindMove(value)
			if len(cells) == 1 {
				cell := cells[0]
				log("The %d on row %d only fits in column %d\n", value, cell.row+1, cell.col+1)
				if err := s.PlayMove(cell.row, cell.col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// Columns where a number only fits in one cell
	for _, col := range s.Cols() {
		for _, value := range col.RemainingMoves().Slice() {
			cells := col.FindMove(value)
			if len(cells) == 1 {
				cell := cells[0]
				log("The %d in column %d only fits at row %d\n", value, cell.col+1, cell.row+1)
				if err := s.PlayMove(cell.row, cell.col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// If a number can only be in one row/col in a square, eliminate the number from that row/col in aligned squares
	for _, square := range s.Squares() {
		squareRow := square[0].row / 3
		squareCol := square[0].col / 3

		for _, value := range square.RemainingMoves().Slice() {
			cells := square.FindMove(value)

			rows := cells.UniqueRows()
			if len(rows) == 1 {
				row := rows[0]
				if s.Row(row).Excluding(square).EliminateMove(value) > 0 {
					log("In the %s square, the number %d only fits in the %s row\n", positionNames[squareRow][squareCol], value, rowPositionNames[row%3])
					moves++
				}
			}

			cols := cells.UniqueCols()
			if len(cols) == 1 {
				col := cols[0]
				if s.Col(col).Excluding(square).EliminateMove(value) > 0 {
					log("In the %s square, the number %d only fits in the %s column\n", positionNames[squareRow][squareCol], value, colPositionNames[col%3])
					moves++
				}
			}
		}
	}

	// If a number can only be played in a single square on a row, eliminate the number from the other rows in the square
	for _, row := range s.Rows() {
		for _, value := range row.RemainingMoves().Slice() {
			cells := row.FindMove(value)
			cols := cells.UniqueCols()
			squareCols := uniqueSquares(cols)

			if len(squareCols) == 1 {
				squareRow := row[0].row / 3
				squareCol := squareCols[0]
				if s.Square(squareRow, squareCol).Excluding(row).EliminateMove(value) > 0 {
					log("The %d in the %s square must be in the %s row\n", value, positionNames[squareRow][squareCol], rowPositionNames[row[0].row%3])
					moves++
				}
			}
		}
	}

	// If a number can only be played in a single square in a column, eliminate the number from the other columns in the square
	for _, col := range s.Cols() {
		for _, value := range col.RemainingMoves().Slice() {
			cells := col.FindMove(value)
			rows := cells.UniqueRows()
			squareRows := uniqueSquares(rows)

			if len(squareRows) == 1 {
				squareRow := squareRows[0]
				squareCol := col[0].col / 3
				if s.Square(squareRow, squareCol).Excluding(col).EliminateMove(value) > 0 {
					log("The %d in the %s square must be in the %s column\n", value, positionNames[squareRow][squareCol], colPositionNames[col[0].col%3])
					moves++
				}
			}
		}
	}

	if len(s.Cells().UnsetOnly()) == 0 {
		fmt.Println("Solved")
		s.PrintBoard()
		return nil
	}

	if moves > 0 {
		s.PrintBoard()
		return s.Solve()
	}

	//for _, cell := range s.Cells().UnsetOnly() {
	//	for _, value := range cell.Moves() {
	//		row := cell.row
	//		col := cell.col
	//		log("Guessing number %d in row %d column %d\n", value, row+1, col+1)
	//
	//		clone := s.Clone()
	//		if err := clone.PlayMove(row, col, value); err != nil {
	//			return err
	//		}
	//		if err := clone.Solve(); err != nil {
	//			log("Bad guess")
	//			continue
	//		}
	//		return nil
	//	}
	//}

	// If a group of numbers
	for _, group := range s.Groups() {
		group = group.UnsetOnly()
		for _, subset := range group.PowerSet() {
			if len(subset) < 2 || len(subset) == len(group) {
				continue
			}
			remainingMoves := subset.RemainingMoves().Slice()

			if len(subset) == len(remainingMoves) {
				otherCells := group.Excluding(subset)
				for _, value := range remainingMoves {
					excludable := otherCells.FindMove(value)
					if len(excludable) > 0 {
						log("The %d can be eliminated from cells %s since it can only be in symmetric cell group %s\n", value, excludable.LocationString(), subset.LocationString())
						excludable.EliminateMove(value)
						moves++
					}
				}
			}
		}
	}

	if moves > 0 {
		s.PrintBoard()
		return s.Solve()
	}

	return errors.New("No solution found")
}

func uniqueSquares(values []int) []int {
	squaresPresent := [3]bool{}
	for _, value := range values {
		squaresPresent[value / 3] = true
	}

	squares := make([]int, 0)
	for square := 0; square < 3; square++ {
		if squaresPresent[square] {
			squares = append(squares, square)
		}
	}

	return squares
}

func log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
