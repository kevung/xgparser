package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
	"strings"
)

func main() {
	// Test the bar entry case
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
	
	fmt.Println("Initial position bar:")
	fmt.Printf("  Player bar (25): %d\n", move.Position.Checkers[25])
	fmt.Printf("  Opponent bar (0): %d\n", move.Position.Checkers[0])
	
	fmt.Println("\nInitial position points 11-13, 23:")
	fmt.Printf("  Point 11: %d\n", move.Position.Checkers[11])
	fmt.Printf("  Point 13: %d\n", move.Position.Checkers[13])
	fmt.Printf("  Point 23: %d\n", move.Position.Checkers[23])
	
	fmt.Printf("\nMove array: %v\n", move.Analysis[0].Move)
	
	fmt.Println("\nFinal position bar:")
	fmt.Printf("  Player bar (25): %d\n", move.Analysis[0].Position.Checkers[25])
	fmt.Printf("  Opponent bar (0): %d\n", move.Analysis[0].Position.Checkers[0])
	
	fmt.Println("\nFinal position points 11-13, 23:")
	fmt.Printf("  Point 11: %d\n", move.Analysis[0].Position.Checkers[11])
	fmt.Printf("  Point 13: %d\n", move.Analysis[0].Position.Checkers[13])
	fmt.Printf("  Point 23: %d\n", move.Analysis[0].Position.Checkers[23])
	
	// Manually check what the XGID says
	fmt.Println("\n=== XGID Analysis ===")
	components, _ := xgparser.ParseXGID("----BaC-B---aD--aa-bcbbBbB:0:0:1:22")
	fmt.Printf("PlayerToMove from XGID: %d\n", components.PlayerToMove)
	fmt.Printf("ActivePlayer from text: %d\n", move.ActivePlayer)
	fmt.Printf("Position was swapped: %v\n", components.PlayerToMove != move.ActivePlayer)
}
