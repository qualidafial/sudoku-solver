package main

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func main() {
	s, err := NewSudokuFromString(`
		 6 | 8 |  1
		79 |  3|  6
		4  |62 |9  
		---+---+---
		   |8  |69 
		   |541|   
		 34|  9|   
		---+---+---
		  5| 94|  7
		1  |3  | 69
		3  | 7 | 1 
	`)
	if err != nil {
		fmt.Println(err)
		s.PrintBoard()
		return
	}

	if err = s.Solve(); err != nil {
		fmt.Println(err)
		s.PrintBoard()
	}
}

type Set map[int]Nothing

type Nothing struct{}

var (
	nothing   = Nothing{}
	nondigits = regexp.MustCompile(`[^0-9 ]`)
)

type Sudoku struct {
	board   [9][9]int
	moves   [9][9]Set
	cols    [9]Set
	rows    [9]Set
	squares [3][3]Set
}

func NewSudokuFromString(board string) (*Sudoku, error) {
	var b = [9][9]int{}
	row := 0
	scanner := bufio.NewScanner(strings.NewReader(board))

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

	return NewSudoku(b)
}

func NewSudoku(board [9][9]int) (*Sudoku, error) {
	s := &Sudoku{}

	allNumbers := func() map[int]Nothing {
		return Set{
			1: nothing,
			2: nothing,
			3: nothing,
			4: nothing,
			5: nothing,
			6: nothing,
			7: nothing,
			8: nothing,
			9: nothing,
		}
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			s.moves[row][col] = allNumbers()
		}
	}

	for i := 0; i < 9; i++ {
		s.rows[i] = allNumbers()
		s.cols[i] = allNumbers()
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			s.squares[i][j] = allNumbers()
		}
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			value := board[row][col]
			if value != 0 {
				if err := s.playMove(row, col, value); err != nil {
					return s, err
				}
			}
		}
	}

	fmt.Println("Initialized Board")
	s.PrintBoard()

	return s, nil
}

func (s *Sudoku) playMove(row int, col int, value int) error {
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
	if !contains(s.rows[row], value) {
		return fmt.Errorf("Row %d already contains %d", row+1, value)
	}
	if !contains(s.cols[col], value) {
		return fmt.Errorf("Col %d already contains %d", col+1, value)
	}
	squareRow, squareCol := row/3, col/3
	if !contains(s.squares[squareRow][squareCol], value) {
		return fmt.Errorf("Square %d,%d already contains %d", squareRow+1, squareCol+1, value)
	}
	if !contains(s.moves[row][col], value) {
		return fmt.Errorf("Square %d,%d is not a valid spot for %d", row+1, col+1, value)
	}

	s.board[row][col] = value
	s.eliminateMoves(row, 0, row, 8, value)
	s.eliminateMoves(0, col, 8, col, value)
	s.eliminateMoves(squareRow*3, squareCol*3, squareRow*3+2, squareCol*3+2, value)
	delete(s.rows[row], value)
	delete(s.cols[col], value)
	delete(s.squares[squareRow][squareCol], value)

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

func (s *Sudoku) Solve() error {
	moves := 0

	// Cells where only a single move is possible
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if s.board[row][col] == 0 {
				if len(s.moves[row][col]) == 1 {
					for value := range s.moves[row][col] {
						if err := s.playMove(row, col, value); err != nil {
							return err
						}
						moves++
						fmt.Println(fmt.Sprintf("Only %d fits at %d,%d", value, row+1, col+1))
					}
				}
			}
		}
	}

	// Squares where a number only fits in one cell
	for squareRow := 0; squareRow < 3; squareRow++ {
		for squareCol := 0; squareCol < 3; squareCol++ {
			for value := range s.squares[squareRow][squareCol] {
				var fittingRows []int
				var fittingCols []int
				for row := squareRow * 3; row < squareRow*3+3; row++ {
					for col := squareCol * 3; col < squareCol*3+3; col++ {
						if contains(s.moves[row][col], value) {
							fittingRows = append(fittingRows, row)
							fittingCols = append(fittingCols, col)
						}
					}
				}
				if len(fittingRows) == 1 && len(fittingCols) == 1 {
					row := fittingRows[0]
					col := fittingCols[0]
					fmt.Println(fmt.Sprintf("In square %d,%d, the number %d only fits at %d,%d", squareRow+1, squareCol+1, value, col%3+1, row%3+1))
					if err := s.playMove(row, col, value); err != nil {
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
			var fittingCols []int
			for col := 0; col < 9; col++ {
				if contains(s.moves[row][col], value) {
					fittingCols = append(fittingCols, col)
				}
			}
			if len(fittingCols) == 1 {
				col := fittingCols[0]
				fmt.Println(fmt.Sprintf("In row %d, %d only fits at column %d", row+1, value, col+1))
				if err := s.playMove(row, col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// Columns where a number only fits in one cell
	for col := 0; col < 9; col++ {
		for value := range s.cols[col] {
			var fittingRows []int
			for row := 0; row < 9; row++ {
				if contains(s.moves[row][col], value) {
					fittingRows = append(fittingRows, col)
				}
			}
			if len(fittingRows) == 1 {
				row := fittingRows[0]
				fmt.Println(fmt.Sprintf("In column %d, %d only fits at row %d", col+1, value, row+1))
				if err := s.playMove(row, col, value); err != nil {
					return err
				}
				moves++
			}
		}
	}

	// If a number can only be in one row in a square, the other squares in that row cannot have that number in that row
	for squareRow := 0; squareRow < 3; squareRow++ {
		for squareCol := 0; squareCol < 3; squareCol++ {
			var fittingRows []int
			for value := range s.squares[squareRow][squareCol] {
				for row := squareRow * 3; row < squareRow*3+3; row++ {
					found := false
					for col := squareCol * 3; col < squareCol*3+3; col++ {
						found = found || contains(s.moves[row][col], value)
					}
					if found {
						fittingRows = append(fittingRows, row)
					}
				}
				if len(fittingRows) == 1 {
					row := fittingRows[0]
					changed := s.eliminateMoves(row, 0, row, squareCol*3-1, value)
					changed += s.eliminateMoves(row, squareCol*3+3, row, 8, value)
					if changed > 0 {
						fmt.Println(fmt.Sprintf("In square %d,%d, the number %d only fits in row %d", squareRow+1, squareCol+1, value, row+1))
						moves++
					}
				}
			}

		}
	}

	// If a number can only be in one column in a square, the other squares in that column cannot have that number in that column
	for squareRow := 0; squareRow < 3; squareRow++ {
		for squareCol := 0; squareCol < 3; squareCol++ {
			var fittingCols []int
			for value := range s.squares[squareRow][squareCol] {
				for col := squareCol * 3; col < squareCol*3+3; col++ {
					found := false
					for row := squareRow * 3; row < squareRow*3+3; row++ {
						found = found || contains(s.moves[row][col], value)
					}
					if found {
						fittingCols = append(fittingCols, col)
					}
				}
				if len(fittingCols) == 1 {
					col := fittingCols[0]
					changed := s.eliminateMoves(0, col, squareRow*3-1, col, value)
					changed += s.eliminateMoves(squareRow*3+3, col, 8, col, value)
					if changed > 0 {
						fmt.Println(fmt.Sprintf("In square %d,%d, the number %d only fits in column %d", squareRow+1, squareCol+1, value, col+1))
						moves++
					}
				}
			}

		}
	}

	if s.emptySpaces(0, 0, 8, 8) == 0 {
		fmt.Println("Solved")
		s.PrintBoard()
		return nil
	}

	if moves == 0 {
		return errors.New("No moves found")
	}

	s.PrintBoard()

	return s.Solve()
}

func (s *Sudoku) eliminateMove(row, col, value int) bool {
	if ! contains(s.moves[row][col], value) {
		return false
	}
	delete(s.moves[row][col], value)
	return true
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

func contains(group Set, value int) bool {
	_, ok := group[value]
	return ok
}
