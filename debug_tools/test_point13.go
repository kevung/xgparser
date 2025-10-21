package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
	"strings"
)

func main() {
	// Test case: Bar entry with no swapping
	input := `XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10
X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2`
	
	move, _, _ := xgparser.ParseXGIDFromReader(strings.NewReader(input))
	
	fmt.Println("Move notation: Bar/23(2) 13/11(2)")
	fmt.Println("Move array:", move.Analysis[0].Move)
	fmt.Println()
	
	fmt.Println("Initial position around point 13:")
	for i := 11; i <= 14; i++ {
		fmt.Printf("  Point %d: %d\n", i, move.Position.Checkers[i])
	}
	
	fmt.Println()
	fmt.Println("Final position around point 13:")
	for i := 11; i <= 14; i++ {
		fmt.Printf("  Point %d: %d\n", i, move.Analysis[0].Position.Checkers[i])
	}
	
	fmt.Println()
	fmt.Println("The move 13/11(2) means:")
	fmt.Println("  Move 2 checkers from point 13 to point 11")
	fmt.Println("  So point 13 should lose 2 checkers")
	fmt.Println("  And point 11 should gain 2 checkers")
	fmt.Println()
	fmt.Println("But we see:")
	fmt.Printf("  Point 13 changed from %d to %d (diff: %d)\n", 
		move.Position.Checkers[13], 
		move.Analysis[0].Position.Checkers[13],
		move.Analysis[0].Position.Checkers[13] - move.Position.Checkers[13])
	fmt.Printf("  Point 11 changed from %d to %d (diff: %d)\n",
		move.Position.Checkers[11],
		move.Analysis[0].Position.Checkers[11],
		move.Analysis[0].Position.Checkers[11] - move.Position.Checkers[11])
	
	// Decode XGID to see what's actually at point 13
	fmt.Println()
	fmt.Println("From XGID decoding:")
	positionID := "----BaC-B---aD--aa-bcbbBbB"
	fmt.Printf("  XGID index 11 (point 13): '%c'\n", positionID[11])
	fmt.Printf("  XGID index 12 (point 12): '%c'\n", positionID[12])
	fmt.Printf("  XGID index 13 (point 11): '%c' = 4 X checkers\n", positionID[13])
}
