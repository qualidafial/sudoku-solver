package main

import (
	"fmt"
	"github.com/sudoku-solver/internal"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]

	s, err := internal.NewSudokuFromFile(path)

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
