package main

import (
	"fmt"
	"strings"
)

// Test function to decode XGID correctly
func decodeXGID(positionID string) {
	fmt.Println("XGID: " + positionID)
	fmt.Println()
	
	// Count total checkers for each player
	xCount := 0
	oCount := 0
	
	for i := 0; i < 26; i++ {
		char := positionID[i]
		if char >= 'A' && char <= 'Z' {
			xCount += int(char - 'A' + 1)
		} else if char >= 'a' && char <= 'o' {
			oCount += int(char - 'a' + 1)
		}
	}
	
	fmt.Printf("Total X checkers: %d\n", xCount)
	fmt.Printf("Total O checkers: %d\n", oCount)
	fmt.Println()
	
	// Decode with different hypotheses
	fmt.Println("Decoding positions (points 1-24 + 2 bar positions):")
	for i := 0; i < 26; i++ {
		char := positionID[i]
		if char == '-' {
			continue
		}
		
		var count int
		var player string
		if char >= 'A' && char <= 'Z' {
			count = int(char - 'A' + 1)
			player = "X"
		} else if char >= 'a' && char <= 'o' {
			count = int(char - 'a' + 1)
			player = "O"
		}
		
		location := ""
		if i < 24 {
			location = fmt.Sprintf("Point %2d", 24-i)
		} else if i == 24 {
			location = "Bar Pos 1"
		} else {
			location = "Bar Pos 2"
		}
		
		fmt.Printf("  XGID[%2d] ('%c') -> %s: %d %s\n", i, char, location, count, player)
	}
}

func main() {
	fmt.Println("=== Test Case 1: X to play with X on bar ===")
	fmt.Println("From file: X has 2 checkers on bar (visible in diagram)")
	fmt.Println()
	decodeXGID("----BaC-B---aD--aa-bcbbBbB")
	
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	
	fmt.Println("=== Test Case 2: Another position ===")
	decodeXGID("-A--B-DCC----A------bbbcdA")
}
