package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
	"strings"
)

func main() {
	// Test case 1: XGID playerToMove=-1, but X to play (needs swapping)
	fmt.Println("=== Case 1: Position needs swapping ===")
	input1 := `XGID=-A--B-DCC----A------bbbcdA:1:1:-1:41:6:2:0:13:10

X:marcow777   O:postmanpat
Score is X:6 O:2 13 pt.(s) match.
Cube: 2, O own cube
X to play 41

    1. 3-ply       4/3 4/Off                    eq:+1.448
      Player:   96.80% (G:56.26% B:0.04%)
      Opponent: 3.20% (G:0.00% B:0.00%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2
`
	
	move1, _, _ := xgparser.ParseXGIDFromReader(strings.NewReader(input1))
	fmt.Printf("playerToMove from XGID: -1 (O)\n")
	fmt.Printf("ActivePlayer from text: %d (X)\n", move1.ActivePlayer)
	fmt.Printf("Position swapped: YES (because they differ)\n\n")
	
	fmt.Println("Initial position (after swap):")
	for i := 1; i <= 24; i++ {
		if move1.Position.Checkers[i] != 0 {
			fmt.Printf("  Point %2d: %2d\n", i, move1.Position.Checkers[i])
		}
	}
	
	fmt.Printf("\nMove notation: 4/3 4/Off\n")
	fmt.Printf("Parsed move array: %v\n", move1.Analysis[0].Move)
	fmt.Println("This means:")
	fmt.Println("  - Move from point 4 to point 3")
	fmt.Println("  - Bear off from point 4")
	fmt.Println("BUT: X's checkers are at points 21-24, NOT at point 4!")
	
	// Test case 2: No swapping needed
	fmt.Println("\n\n=== Case 2: No swapping needed ===")
	input2 := `XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10

X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2
`
	
	move2, _, _ := xgparser.ParseXGIDFromReader(strings.NewReader(input2))
	fmt.Printf("playerToMove from XGID: 1 (X)\n")
	fmt.Printf("ActivePlayer from text: %d (X)\n", move2.ActivePlayer)
	fmt.Printf("Position swapped: NO (because they match)\n\n")
	
	fmt.Println("Initial position (no swap):")
	for i := 1; i <= 24; i++ {
		if move2.Position.Checkers[i] != 0 {
			fmt.Printf("  Point %2d: %2d\n", i, move2.Position.Checkers[i])
		}
	}
	
	fmt.Printf("\nMove notation: Bar/23(2) 13/11(2)\n")
	fmt.Printf("Parsed move array: %v\n", move2.Analysis[0].Move)
	fmt.Println("This works correctly because:")
	fmt.Println("  - Bar is index 25")
	fmt.Println("  - X has 2 checkers on bar (25)")
	fmt.Println("  - Move them to point 23")
}
