package main

import (
	"fmt"
)

func main() {
	fmt.Println("Analyzing XGID position mapping")
	fmt.Println()
	
	// From the file, X has checkers at these points (counting from ASCII):
	// Point 13: 4 checkers (visible as X X X X in column 13)
	// Point 11: appears to have checkers
	// Point 20, 16, 18: have checkers
	// Point 1: has checkers
	// And on the bar: 2 checkers
	
	positionID := "----BaC-B---aD--aa-bcbbBbB"
	
	fmt.Println("XGID: " + positionID)
	fmt.Println()
	fmt.Println("Let's try different mapping hypothesis:")
	fmt.Println()
	
	fmt.Println("Hypothesis: XGID index i = point (i+1) [NOT reversed]")
	for i := 0; i < 24; i++ {
		char := positionID[i]
		if char != '-' {
			var count int
			var player string
			if char >= 'A' && char <= 'Z' {
				count = int(char - 'A' + 1)
				player = "X"
			} else if char >= 'a' && char <= 'o' {
				count = int(char - 'a' + 1)
				player = "O"
			}
			fmt.Printf("  Index %2d ('%c') -> Point %2d: %d %s\n", i, char, i+1, count, player)
		}
	}
	
	fmt.Println()
	fmt.Println("From the ASCII board, X checkers are at:")
	fmt.Println("  Point 13: 4 checkers")
	fmt.Println("  Point 11: some checkers")
	fmt.Println("  Point 20: 2 checkers")
	fmt.Println("  Point 16: 2 checkers")
	fmt.Println("  Point 18: 3 checkers")
	fmt.Println("  Point 1: 2 checkers")
	fmt.Println()
	fmt.Println("With this mapping:")
	fmt.Println("  Index 12 ('a') -> Point 13: 1 O ❌ (should be 4 X)")
	fmt.Println()
	fmt.Println("This doesn't match either!")
}
