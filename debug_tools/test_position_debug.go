package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
)

func printPosition(label string, pos xgparser.Position) {
	fmt.Printf("\n%s:\n", label)
	fmt.Println("Checkers array:")
	for i := 0; i <= 25; i++ {
		if pos.Checkers[i] != 0 {
			var location string
			if i == 0 {
				location = "Opponent Bar"
			} else if i == 25 {
				location = "Player Bar"
			} else {
				location = fmt.Sprintf("Point %d", i)
			}
			fmt.Printf("  %s: %d\n", location, pos.Checkers[i])
		}
	}
}

func main() {
	// Test XGID parsing for a bear-off position
	filepath := "tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt"
	
	move, _, err := xgparser.ParseXGIDFile(filepath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println("=== XGID Analysis ===")
	fmt.Printf("XGID: -A--B-DCC----A------bbbcdA:1:1:-1:41\n")
	fmt.Printf("Position encoding breakdown:\n")
	fmt.Printf("  playerToMove (from XGID): -1 (O's perspective)\n")
	fmt.Printf("  ActivePlayer (from text):  %d (X to play)\n", move.ActivePlayer)
	fmt.Printf("  Cube owner: %d\n", move.Position.CubePos)
	fmt.Printf("  Dice: %v\n", move.Dice)
	
	// The XGID encodes from O's perspective but X is to play
	// So we need to understand if swapping happened
	printPosition("Initial Position (should be from X's perspective since X to play)", move.Position)
	
	if len(move.Analysis) > 0 {
		fmt.Printf("\nAnalyzing first move: %v\n", move.Analysis[0].Move)
		
		// Manually check if the move makes sense from the initial position
		firstMove := move.Analysis[0].Move
		fmt.Printf("\nMove analysis:\n")
		for i := 0; i < 8; i += 2 {
			if firstMove[i] == -1 {
				break
			}
			from := firstMove[i]
			to := firstMove[i+1]
			fmt.Printf("  Step: %d -> %d", from, to)
			if to == -2 {
				fmt.Printf(" (bear off)")
			}
			fmt.Printf(" | Checkers at source: %d\n", move.Position.Checkers[from])
		}
		
		printPosition("Position after move (should show changed checkers)", move.Analysis[0].Position)
		
		// Check if the position changed
		changed := false
		for i := 0; i < 26; i++ {
			if move.Position.Checkers[i] != move.Analysis[0].Position.Checkers[i] {
				changed = true
				break
			}
		}
		fmt.Printf("\nPosition changed after applying move: %v\n", changed)
	}
	
	// Test a second file with O to play
	fmt.Println("\n\n=== Testing O to play scenario ===")
	filepath2 := "tmp/xgid/en/XGID=----C-E-A----C-AB------cb-:0:0:-1:6.txt"
	move2, _, err2 := xgparser.ParseXGIDFile(filepath2)
	if err2 != nil {
		fmt.Printf("Error: %v\n", err2)
		return
	}
	
	fmt.Printf("XGID: ----C-E-A----C-AB------cb-:0:0:-1:6\n")
	fmt.Printf("  playerToMove (from XGID): -1\n")
	fmt.Printf("  ActivePlayer (from text):  %d\n", move2.ActivePlayer)
	
	printPosition("Initial Position", move2.Position)
	
	if len(move2.Analysis) > 0 {
		fmt.Printf("\nFirst move: %v\n", move2.Analysis[0].Move)
		printPosition("Position after move", move2.Analysis[0].Position)
	}
}
