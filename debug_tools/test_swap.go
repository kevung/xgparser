package main

import (
	"fmt"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	// Test XGID parsing for the problematic file
	filepath := "tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt"

	move, _, err := xgparser.ParseXGIDFile(filepath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Position checkers (first 10 points):")
	for i := 1; i <= 10; i++ {
		if move.Position.Checkers[i] != 0 {
			fmt.Printf("  Point %d: %d\n", i, move.Position.Checkers[i])
		}
	}

	fmt.Printf("\nActive player: %d\n", move.ActivePlayer)

	// Check first analysis
	if len(move.Analysis) > 3 {
		analysis := move.Analysis[3]
		fmt.Printf("\nAnalysis 4 move: %v\n", analysis.Move)
		fmt.Printf("  From point %d (has %d checkers)\n", analysis.Move[0], move.Position.Checkers[analysis.Move[0]])
		fmt.Printf("  Position changed: %v\n", move.Position.Checkers != analysis.Position.Checkers)
	}
}
