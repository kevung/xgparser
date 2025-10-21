package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
)

func main() {
	// Decode XGID position string manually
	positionID := "-A--B-DCC----A------bbbcdA"
	
	fmt.Println("XGID Position String: " + positionID)
	fmt.Println("\nDecoding (XGID format - index 0 = point 24, index 23 = point 1):")
	
	for i := 0; i < 26; i++ {
		char := positionID[i]
		var count int
		var player string
		
		if char == '-' {
			count = 0
			player = "none"
		} else if char >= 'A' && char <= 'Z' {
			count = int(char - 'A' + 1)
			player = "X"
		} else if char >= 'a' && char <= 'o' {
			count = int(char - 'a' + 1)
			player = "O"
		}
		
		if count > 0 {
			var location string
			if i < 24 {
				// Points in reverse order: index 0 = point 24, index 23 = point 1
				location = fmt.Sprintf("Point %d", 24-i)
			} else if i == 24 {
				location = "X's bar"
			} else if i == 25 {
				location = "O's bar"
			}
			fmt.Printf("  Index %2d ('%c') = %s: %d %s checker(s)\n", i, char, location, count, player)
		}
	}
	
	// Now see how XGIDToPosition converts it
	position := xgparser.XGIDToPosition(positionID)
	
	fmt.Println("\nAfter XGIDToPosition (internal array):")
	for i := 0; i <= 25; i++ {
		if position[i] != 0 {
			var location string
			if i == 0 {
				location = "Opponent Bar"
			} else if i == 25 {
				location = "Player Bar"
			} else {
				location = fmt.Sprintf("Point %d", i)
			}
			fmt.Printf("  Array[%2d] = %s: %2d\n", i, location, position[i])
		}
	}
	
	// Parse the full file to see what happens with swapping
	fmt.Println("\n=== Full file parse ===")
	move, _, _ := xgparser.ParseXGIDFile("tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt")
	
	fmt.Println("\nAfter swapping (because playerToMove=-1 but ActivePlayer=1):")
	for i := 0; i <= 25; i++ {
		if move.Position.Checkers[i] != 0 {
			var location string
			if i == 0 {
				location = "Opponent Bar"
			} else if i == 25 {
				location = "Player Bar"
			} else {
				location = fmt.Sprintf("Point %d", i)
			}
			fmt.Printf("  Array[%2d] = %s: %2d\n", i, location, move.Position.Checkers[i])
		}
	}
	
	fmt.Println("\n=== Understanding the board ===")
	fmt.Println("Looking at the ASCII board from the file:")
	fmt.Println("  Points 1-6 (X's home): O has checkers at points 2, 5")
	fmt.Println("  Points 7-12: O has checkers at points 7, 8, 9")
	fmt.Println("  O has 1 checker on the bar")
	fmt.Println("  Points 19-24 (O's home): X has checkers at points 21, 22, 23, 24")
	fmt.Println("  X also has 1 checker on O's bar (captured)")
	
	fmt.Println("\nMove '4/3 4/Off' means:")
	fmt.Println("  - From X's perspective")
	fmt.Println("  - X moves from point 4 to point 3")
	fmt.Println("  - X bears off from point 4")
	fmt.Println("  - But after swapping, X's checkers are in array indices 21-24")
	fmt.Println("  - So we need array index 4 to contain X checkers for this to work!")
}
