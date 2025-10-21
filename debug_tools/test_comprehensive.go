package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
	"strings"
)

func testCase(name string, input string, expectedDiff map[int]int) bool {
	fmt.Printf("\n=== %s ===\n", name)
	
	move, _, err := xgparser.ParseXGIDFromReader(strings.NewReader(input))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	
	if len(move.Analysis) == 0 {
		fmt.Println("No analysis found")
		return false
	}
	
	// Calculate differences
	actualDiff := make(map[int]int)
	for i := 0; i <= 25; i++ {
		diff := int(move.Analysis[0].Position.Checkers[i]) - int(move.Position.Checkers[i])
		if diff != 0 {
			actualDiff[i] = diff
		}
	}
	
	fmt.Println("Initial position:")
	for i := 0; i <= 25; i++ {
		if move.Position.Checkers[i] != 0 {
			fmt.Printf("  [%2d]: %2d\n", i, move.Position.Checkers[i])
		}
	}
	
	fmt.Printf("\nMove: %v\n", move.Analysis[0].Move)
	
	fmt.Println("\nExpected changes:")
	for point, diff := range expectedDiff {
		fmt.Printf("  [%2d]: %+d\n", point, diff)
	}
	
	fmt.Println("\nActual changes:")
	for point, diff := range actualDiff {
		fmt.Printf("  [%2d]: %+d\n", point, diff)
	}
	
	// Verify
	passed := true
	for point, expectedChange := range expectedDiff {
		if actualDiff[point] != expectedChange {
			fmt.Printf("❌ Mismatch at point %d: expected %+d, got %+d\n", point, expectedChange, actualDiff[point])
			passed = false
		}
	}
	
	// Check for unexpected changes
	for point, actualChange := range actualDiff {
		if _, exists := expectedDiff[point]; !exists {
			fmt.Printf("❌ Unexpected change at point %d: %+d\n", point, actualChange)
			passed = false
		}
	}
	
	if passed {
		fmt.Println("✓ PASSED")
	}
	
	return passed
}

func main() {
	allPassed := true
	
	// Test 1: Bear-off with position swapping (XGID from O's perspective, X to play)
	allPassed = testCase(
		"Bear-off with swapping",
		`XGID=-A--B-DCC----A------bbbcdA:1:1:-1:41:6:2:0:13:10
X:marcow777   O:postmanpat
Score is X:6 O:2 13 pt.(s) match.
Cube: 2, O own cube
X to play 41

    1. 3-ply       4/3 4/Off                    eq:+1.448
      Player:   96.80% (G:56.26% B:0.04%)
      Opponent: 3.20% (G:0.00% B:0.00%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2`,
		map[int]int{
			21: -2, // Remove 2 checkers from point 21
			22: +1, // Add 1 checker to point 22
			// 1 checker is borne off (no array change)
		},
	) && allPassed
	
	// Test 2: No swapping needed (XGID from X's perspective, X to play)
	allPassed = testCase(
		"Bar entry with no swapping",
		`XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10
X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2`,
		map[int]int{
			25: -2, // Remove 2 from bar
			23: +2, // Add 2 to point 23
			13: -2, // Remove 2 from point 13
			11: +2, // Add 2 to point 11
		},
	) && allPassed
	
	// Test 3: Another bear-off case with swapping
	allPassed = testCase(
		"Bear-off 61 with swapping",
		`XGID=----C-E-A----C-AB------cb-:0:0:-1:61:4:2:0:13:10
X:marcow777   O:postmanpat
Score is X:4 O:2 13 pt.(s) match.
Cube: 1
X to play 61

    1. 4-ply       2/1 2/Off                    eq:+2.009
      Player:   99.98% (G:99.98% B:0.00%)
      Opponent: 0.02% (G:0.00% B:0.00%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2`,
		map[int]int{
			23: -2, // Remove 2 from point 23 (swapped from point 2)
			24: +1, // Add 1 to point 24 (swapped from point 1)
			// 1 checker is borne off
		},
	) && allPassed
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	if allPassed {
		fmt.Println("✓ ALL TESTS PASSED!")
	} else {
		fmt.Println("❌ SOME TESTS FAILED")
	}
}
