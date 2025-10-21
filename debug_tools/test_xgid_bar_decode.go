package main

import (
	"fmt"
)

func main() {
	// XGID: ----BaC-B---aD--aa-bcbbBbB:0:0:1:22
	// PlayerToMove: 1 (X)
	
	positionID := "----BaC-B---aD--aa-bcbbBbB"
	
	fmt.Println("XGID Position Decoding:")
	fmt.Println("Position string: " + positionID)
	fmt.Println("PlayerToMove: 1 (X)")
	fmt.Println()
	
	fmt.Println("Character-by-character:")
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
			location = fmt.Sprintf("Point %d", 24-i)
		} else if i == 24 {
			location = "Index 24 (bar position 1)"
		} else {
			location = "Index 25 (bar position 2)"
		}
		
		fmt.Printf("  %s ['%c']: %d %s checker(s)\n", location, char, count, player)
	}
	
	fmt.Println("\nFrom the ASCII board:")
	fmt.Println("  X has checkers on the bar (shown in BAR column)")
	fmt.Println("  O has checkers on points")
	fmt.Println()
	fmt.Println("Key observation:")
	fmt.Println("  Index 24 = 'b' = 2 O checkers")
	fmt.Println("  Index 25 = 'B' = 2 X checkers")
	fmt.Println()
	fmt.Println("Hypothesis 1: XGID is always from X's absolute perspective")
	fmt.Println("  Index 24 = O's bar (from X's view)")
	fmt.Println("  Index 25 = X's bar (from X's view)")
	fmt.Println("  This matches! X has 2 on bar, O has 0 visible on bar in the diagram")
	fmt.Println()
	fmt.Println("Wait, let me check the board again...")
	fmt.Println("Actually, looking at the BAR, I see X pieces")
	fmt.Println("So X definitely has checkers on the bar")
	fmt.Println()
	fmt.Println("Conclusion: XGID uses absolute perspective")
	fmt.Println("  Index 24 = opponent's bar point (where opponent's hit checkers go)")
	fmt.Println("  Index 25 = player's bar point (where player's hit checkers go)")
	fmt.Println("  From X's view: index 24 = O's bar, index 25 = X's bar")
	fmt.Println("  'b' at index 24 means O checkers, but wait - that would be O checkers ON O's bar?")
	fmt.Println()
	fmt.Println("Let me reconsider... In backgammon:")
	fmt.Println("  When X is hit, X goes to 'the bar'")
	fmt.Println("  When O is hit, O goes to 'the bar'")
	fmt.Println("  There's only ONE bar, but checkers enter from different sides")
	fmt.Println()
	fmt.Println("XGID must encode: X's checkers on bar, O's checkers on bar")
	fmt.Println("  Index 24 'b' = 2... but lowercase = O notation")
	fmt.Println("  Index 25 'B' = 2... uppercase = X notation")
	fmt.Println()
	fmt.Println("ERROR in my understanding! Let me check if bars are swapped in definition")
}
