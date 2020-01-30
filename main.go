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
		   |5  | 9 
		   | 79|   
		8  |2 4|5  
		---+---+---
		  3|  8|45 
		  2| 9 |3  
		 57|3  |6  
		---+---+---
		  8|6 2|  1
		   |81 |   
		 6 |  3|   
	`)
	if err != nil {
		fmt.Println(err)
		s.PrintBoard()
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

var (
	full  = Set{false, true, true, true, true, true, true, true, true, true}
	empty = Set{false, false, false, false, false, false, false, false, false, false}

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

	fmt.Println("Initialized Board")
	s.PrintBoard()

	return s, nil
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
		return fmt.Errorf("Square %d,%d already contains %d", squareRow+1, squareCol+1, value)
	}
	if !s.moves[row][col][value] {
		return fmt.Errorf("Square %d,%d is not a valid spot for %d", row+1, col+1, value)
	}

	s.board[row][col] = value
	s.moves[row][col] = empty
	s.eliminateMoves(row, 0, row, 8, value)
	s.eliminateMoves(0, col, 8, col, value)
	s.eliminateMoves(squareRow*3, squareCol*3, squareRow*3+2, squareCol*3+2, value)
	s.rows[row][value] = false
	s.cols[col][value] = false
	s.squares[squareRow][squareCol][value] = false

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
			if len(s.moves[row][col]) == 1 {
				for value := range s.moves[row][col] {
					if err := s.PlayMove(row, col, value); err != nil {
						return err
					}
					moves++
					fmt.Println(fmt.Sprintf("Only %d fits in cell %d,%d", value, row+1, col+1))
				}
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
					fmt.Println(fmt.Sprintf("In square %d,%d, the number %d only fits in cell %d,%d", squareRow+1, squareCol+1, value, row+1, col+1))
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
				fmt.Println(fmt.Sprintf("In row %d, %d only fits at column %d", row+1, value, col+1))
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
				fmt.Println(fmt.Sprintf("In column %d, %d only fits at row %d", col+1, value, row+1))
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

				rows := uniqueRows(cells)
				if len(rows) == 1 {
					row := rows[0]
					changed := s.eliminateMoves(row, 0, row, squareCol*3-1, value)
					changed += s.eliminateMoves(row, squareCol*3+3, row, 8, value)
					if changed > 0 {
						fmt.Println(fmt.Sprintf("In square %d,%d, the number %d only fits in row %d", squareRow+1, squareCol+1, value, row+1))
						moves++
					}
				}

				cols := uniqueCols(cells)
				if len(cols) == 1 {
					col := cols[0]
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
	if !s.moves[row][col][value] {
		return false
	}
	s.moves[row][col][value] = false
	return true
}

func (s *Sudoku) findMoves(top, left, bottom, right, value int) []Cell {
	var result []Cell
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

func uniqueRows(cells []Cell) []int {
	var rows = empty
	for _, cell := range cells {
		rows[cell.row] = true
	}

	return rows.slice()
}

func uniqueCols(cells []Cell) []int {
	var cols = empty
	for _, cell := range cells {
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
