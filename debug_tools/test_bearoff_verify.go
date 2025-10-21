package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
)

func main() {
	// Parse the bear-off file
	move, _, err := xgparser.ParseXGIDFile("tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println("=== Verifying the bear-off move ===")
	fmt.Println("\nFrom the file:")
	fmt.Println("  Move: 4/3 4/Off")
	fmt.Println("  Dice: 41")
	fmt.Println("  This means: use the 4 to move from point 4 to point 3,")
	fmt.Println("             and use the 1 to bear off from point 4")
	fmt.Println()
	fmt.Println("Initial position (X's perspective after swapping):")
	fmt.Println("  Point 21: 2 checkers")
	fmt.Println("  Point 22: 2 checkers")
	fmt.Println("  Point 23: 2 checkers")
	fmt.Println("  Point 24: 3 checkers")
	fmt.Println()
	fmt.Println("After swapping, move 4/3 becomes 21/22:")
	fmt.Println("  Move from point 4 -> point 3  becomes  point 21 -> point 22")
	fmt.Println("  (25-4=21, 25-3=22)")
	fmt.Println()
	fmt.Println("After swapping, move 4/Off stays 21/Off:")
	fmt.Println("  Move from point 4 -> Off  becomes  point 21 -> Off")
	fmt.Println()
	
	fmt.Printf("Swapped move array: %v\n", move.Analysis[0].Move)
	fmt.Println()
	
	fmt.Println("Expected result:")
	fmt.Println("  Point 21: 0 checkers (2 removed: 1 to 22, 1 borne off)")
	fmt.Println("  Point 22: 3 checkers (2 + 1 from 21)")
	fmt.Println("  Point 23: 2 checkers (unchanged)")
	fmt.Println("  Point 24: 3 checkers (unchanged)")
	fmt.Println()
	
	fmt.Println("Actual result:")
	for i := 21; i <= 24; i++ {
		fmt.Printf("  Point %d: %d checkers\n", i, move.Analysis[0].Position.Checkers[i])
	}
	
	// Verify correctness
	correct := true
	expected := map[int]int8{21: 0, 22: 3, 23: 2, 24: 3}
	for point, expectedCount := range expected {
		actual := move.Analysis[0].Position.Checkers[point]
		if actual != expectedCount {
			fmt.Printf("\n❌ MISMATCH at point %d: expected %d, got %d\n", point, expectedCount, actual)
			correct = false
		}
	}
	
	if correct {
		fmt.Println("\n✓ All positions match expected values!")
	}
}
