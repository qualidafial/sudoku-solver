package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]

	s, err := NewSudokuFromFile(path)

	if err != nil {
		fmt.Println(err)
		if s != nil {
			s.PrintBoard()
		}
		return
	}

	if err = s.Solve(); err != nil {
		fmt.Println(err)
		s.PrintBoard()
		s.PrintMoves()
	}
}

type Set [10]bool

type Cell struct {
	row int
	col int
}

type Cells []Cell

var (
	empty = Set{false, false, false, false, false, false, false, false, false, false}
	full  = Set{false, true, true, true, true, true, true, true, true, true}

	rowPositionNames = []string{"top", "center", "bottom"}
	colPositionNames = []string{"left", "center", "right"}
	positionNames    = [][]string{
		{"top left", "top center", "top right"},
		{"center left", "center", "center right"},
		{"bottom left", "bottom center", "bottom right"},
	}

	nondigits = regexp.MustCompile(`[^0-9 ]`)
)

type Sudoku struct {
	board   [9][9]int
	moves   [9][9]Set
	cols    [9]Set
	rows    [9]Set
	squares [3][3]Set
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

func NewSudokuFromFile(path string) (*Sudoku, error){
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
			s.moves[row][col] = full
		}
	}

	for i := 0; i < 9; i++ {
		s.rows[i] = full
		s.cols[i] = full
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			s.squares[i][j] = full
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

func (s *Sudoku) Clone() *Sudoku {
	return &Sudoku{
		board:   s.board,
		moves:   s.moves,
		cols:    s.cols,
		rows:    s.rows,
		squares: s.squares,
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

	if s.board[row][col] != 0 {
		return fmt.Errorf("Cell %d,%d already contains %d", row+1, col+1, s.board[row][col])
	}
	if !s.rows[row][value] {
		return fmt.Errorf("Row %d already contains %d", row+1, value)
	}
	if !s.cols[col][value] {
		return fmt.Errorf("Col %d already contains %d", col+1, value)
	}
	squareRow, squareCol := row/3, col/3
	if !s.squares[squareRow][squareCol][value] {
		return fmt.Errorf("The %s square already contains %d", positionNames[squareRow][squareCol], value)
	}
	if !s.moves[row][col][value] {
		return fmt.Errorf("Cell %d,%d is not a valid spot for %d", row+1, col+1, value)
	}

	s.board[row][col] = value
	s.moves[row][col] = empty
	s.eliminateMoves(row, 0, row, 8, value)
	s.eliminateMoves(0, col, 8, col, value)
	s.eliminateMoves(squareRow*3, squareCol*3, squareRow*3+2, squareCol*3+2, value)
	s.rows[row][value] = false
	s.cols[col][value] = false
	s.squares[squareRow][squareCol][value] = false

	for row := 0; row < 9; row++ {
		for cell := 0; cell < 9; cell++ {
			if s.board[row][cell] == 0 && s.moves[row][cell] == empty {
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
			value := s.board[row][col]
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
					if s.moves[row][col][value] {
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
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			values := s.moves[row][col].slice()
			if len(values) == 1 {
				value := values[0]
				if err := s.PlayMove(row, col, value); err != nil {
					return err
				}
				moves++
				log("Only %d fits in row %d column %d\n", value, row+1, col+1)
			}
		}
	}

	// Squares where a number only fits in one cell
	for squareRow := 0; squareRow < 3; squareRow++ {
		for squareCol := 0; squareCol < 3; squareCol++ {
			for value := range s.squares[squareRow][squareCol] {
				cells := s.findMoves(squareRow*3, squareCol*3, squareRow*3+2, squareCol*3+2, value)

				if len(cells) == 1 {
					row, col := cells[0].row, cells[0].col
					log("In the %s square, the number %d only fits in the %s cell\n", positionNames[squareRow][squareCol], value, positionNames[row%3][col%3])
					if err := s.PlayMove(row, col, value); err != nil {
						return err
					}
					moves++
				}
			}
		}
	}

	// Rows where a number only fits in one cell
	for row := 0; row < 9; row++ {
		for value := range s.rows[row] {
			cells := s.findMoves(row, 0, row, 8, value)

			if len(cells) == 1 {
				col := cells[0].col
				log("The %d on row %d only fits in column %d\n", value, row+1, col+1)
				if err := s.PlayMove(row, col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// Columns where a number only fits in one cell
	for col := 0; col < 9; col++ {
		for value := range s.cols[col] {
			cells := s.findMoves(0, col, 8, col, value)

			if len(cells) == 1 {
				row := cells[0].row
				log("The %d in column %d only fits at row %d\n", value, col+1, row+1)
				if err := s.PlayMove(row, col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// If a number can only be in one row/col in a square, eliminate the number from that row/col in aligned squares
	for squareRow := 0; squareRow < 3; squareRow++ {
		for squareCol := 0; squareCol < 3; squareCol++ {
			for value := range s.squares[squareRow][squareCol] {
				cells := s.findMoves(squareRow*3, squareCol*3, squareRow*3+2, squareCol*3+2, value)

				rows := cells.uniqueRows()
				if len(rows) == 1 {
					row := rows[0]
					changed := s.eliminateMoves(row, 0, row, squareCol*3-1, value)
					changed += s.eliminateMoves(row, squareCol*3+3, row, 8, value)
					if changed > 0 {
						log("In the %s square, the number %d only fits in the %s row\n", positionNames[squareRow][squareCol], value, rowPositionNames[row%3])
						moves++
					}
				}

				cols := cells.uniqueCols()
				if len(cols) == 1 {
					col := cols[0]
					changed := s.eliminateMoves(0, col, squareRow*3-1, col, value)
					changed += s.eliminateMoves(squareRow*3+3, col, 8, col, value)
					if changed > 0 {
						log("In the %s square, the number %d only fits in the %s column\n", positionNames[squareRow][squareCol], value, colPositionNames[col%3])
						moves++
					}
				}
			}
		}
	}

	// If a number can only be played in a single square on a row, eliminate the number from the other rows in the square
	for row := 0; row < 9; row++ {
		squareRow := row / 3
		top, bottom := squareRow*3, squareRow*3+2

		for value := range s.rows[row] {
			cols := s.findMoves(row, 0, row, 8, value).uniqueCols()
			squareCols := uniqueSquares(cols)

			if len(squareCols) == 1 {
				squareCol := squareCols[0]
				left, right := squareCol*3, squareCol*3+2
				changed := s.eliminateMoves(top, left, row-1, right, value)
				changed += s.eliminateMoves(row+1, left, bottom, right, value)
				if changed > 0 {
					log("The %d in the %s square must be in the %s row\n", value, positionNames[squareRow][squareCol], rowPositionNames[row%3])
					moves++
				}
			}
		}
	}

	// If a number can only be played in a single square in a column, eliminate the number from the other columns in the square
	for col := 0; col < 9; col++ {
		squareCol := col / 3
		left, right := squareCol*3, squareCol*3+2

		for value := range s.cols[col] {
			rows := s.findMoves(0, col, 8, col, value).uniqueRows()
			squareRows := uniqueSquares(rows)

			if len(squareRows) == 1 {
				squareRow := squareRows[0]
				top, bottom := squareRow*3, squareRow*3+2
				changed := s.eliminateMoves(top, left, bottom, col-1, value)
				changed += s.eliminateMoves(top, col+1, bottom, right, value)
				if changed > 0 {
					log("The %d in the %s square must be in the %s column\n", value, positionNames[squareRow][squareCol], colPositionNames[col%3])
					moves++
				}
			}
		}
	}

	if s.emptySpaces(0, 0, 8, 8) == 0 {
		fmt.Println("Solved")
		s.PrintBoard()
		return nil
	}

	if moves > 0 {
		s.PrintBoard()
		return s.Solve()
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			for _, value := range s.moves[row][col].slice() {
				log("Guessing number %d in row %d column %d\n", value, row+1, col+1)

				clone := s.Clone()
				if err := clone.PlayMove(row, col, value); err != nil {
					return err
				}
				if err := clone.Solve(); err != nil {
					log("Bad guess")
					continue
				}
				return nil
			}
		}
	}

	return errors.New("No solution found")
}

func (s *Sudoku) eliminateMove(row, col, value int) bool {
	if !s.moves[row][col][value] {
		return false
	}
	s.moves[row][col][value] = false
	return true
}

func (s *Sudoku) findMoves(top, left, bottom, right, value int) Cells {
	var result Cells
	for row := top; row <= bottom; row++ {
		for col := left; col <= right; col++ {
			if s.moves[row][col][value] {
				result = append(result, Cell{row, col})
			}
		}
	}
	return result
}

func (s *Sudoku) eliminateMoves(top, left, bottom, right, value int) int {
	count := 0
	for row := top; row <= bottom; row++ {
		for col := left; col <= right; col++ {
			if s.eliminateMove(row, col, value) {
				count++
			}
		}
	}
	return count
}

func (s *Sudoku) emptySpaces(top, left, bottom, right int) int {
	count := 0
	for row := top; row <= bottom; row++ {
		for col := left; col <= right; col++ {
			if s.board[row][col] == 0 {
				count++
			}
		}
	}
	return count
}

func (c Cells) uniqueRows() []int {
	var rows = empty

	for _, cell := range c {
		rows[cell.row] = true
	}

	return rows.slice()
}

func (c Cells) uniqueCols() []int {
	var cols = empty

	for _, cell := range c {
		cols[cell.col] = true
	}

	return cols.slice()
}

func (s Set) slice() []int {
	var values []int
	for val, present := range s {
		if present {
			values = append(values, val)
		}
	}
	return values
}

func uniqueSquares(values []int) []int {
	var squares Set
	for _, value := range values {
		squares[value/3] = true
	}
	return squares.slice()
}

func log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}