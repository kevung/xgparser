package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
)

func main() {
	// Decode the XGID position
	positionID := "----BaC-B---aD--aa-bcbbBbB"
	position := xgparser.XGIDToPosition(positionID)
	
	fmt.Println("XGID: ----BaC-B---aD--aa-bcbbBbB:0:0:1:22")
	fmt.Println("PlayerToMove: 1 (X)")
	fmt.Println()
	fmt.Println("Decoded position:")
	for i := 0; i <= 25; i++ {
		if position[i] != 0 {
			location := ""
			if i == 0 {
				location = "Opponent bar (O's bar from X's perspective)"
			} else if i == 25 {
				location = "Player bar (X's bar from X's perspective)"
			} else {
				location = fmt.Sprintf("Point %d", i)
			}
			fmt.Printf("  [%2d] %s: %2d\n", i, location, position[i])
		}
	}
	
	fmt.Println()
	fmt.Println("Looking at the ASCII board in the file:")
	fmt.Println("  BAR shows: | X | with 2 X pieces")
	fmt.Println("  So X has 2 checkers on the bar")
	fmt.Println()
	fmt.Println("In our internal representation:")
	fmt.Println("  X is the active player (playerToMove=1)")
	fmt.Println("  X's bar should be at index 25")
	fmt.Println(fmt.Sprintf("  But position[25] = %d (negative = O checkers!)", position[25]))
	fmt.Println()
	fmt.Println("This suggests an issue with XGIDToPosition mapping!")
}
