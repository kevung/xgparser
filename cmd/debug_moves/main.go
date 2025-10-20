package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug_moves <xg_file>")
		os.Exit(1)
	}

	filename := os.Args[1]
	fmt.Printf("Processing file: %s\n\n", filename)

	// Parse using full parser to get raw data
	imp := xgparser.NewImport(filename)
	segments, err := imp.GetFileSegments()
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	fileVersion := int32(-1)
	for _, segment := range segments {
		if segment.Type == xgparser.SegmentXGGameFile {
			records, err := xgparser.ParseGameFile(segment.Data, fileVersion)
			if err != nil {
				log.Fatalf("Error parsing game file: %v", err)
			}

			gameNum := 0
			moveNum := 0

			for _, rec := range records {
				switch r := rec.(type) {
				case *xgparser.HeaderMatchEntry:
					fileVersion = r.Version
					fmt.Printf("Match: %s vs %s\n", r.Player1, r.Player2)
					fmt.Printf("Version: %d\n\n", fileVersion)

				case *xgparser.HeaderGameEntry:
					gameNum++
					moveNum = 0
					fmt.Printf("=== Game %d ===\n", gameNum)

				case *xgparser.MoveEntry:
					moveNum++
					fmt.Printf("\nMove %d:\n", moveNum)
					fmt.Printf("  ActivePlayer: %d\n", r.ActiveP)
					fmt.Printf("  Dice: %v\n", r.Dice)
					fmt.Printf("  Raw XG Moves: %v\n", r.Moves)

					// Show what the conversion produces
					var converted [8]int32
					for i := 0; i < 8; i++ {
						if r.Moves[i] == -1 {
							converted[i] = -1
						} else if r.Moves[i] == -2 {
							converted[i] = -2
						} else if r.Moves[i] == 0 {
							converted[i] = 25
						} else {
							converted[i] = r.Moves[i] + 1
						}
					}
					fmt.Printf("  Converted:    %v\n", converted)

					// Show analysis if available
					if r.DataMoves != nil && r.DataMoves.NMoves > 0 {
						fmt.Printf("  Analysis (%d moves):\n", r.DataMoves.NMoves)
						for i := 0; i < int(r.DataMoves.NMoves) && i < 3; i++ {
							fmt.Printf("    Option %d: Raw XG: %v\n", i+1, r.DataMoves.Moves[i])

							var convAnalysis [8]int8
							for j := 0; j < 8; j++ {
								if r.DataMoves.Moves[i][j] == -1 {
									convAnalysis[j] = -1
								} else if r.DataMoves.Moves[i][j] == -2 {
									convAnalysis[j] = -2
								} else if r.DataMoves.Moves[i][j] == 0 {
									convAnalysis[j] = 25
								} else {
									convAnalysis[j] = r.DataMoves.Moves[i][j] + 1
								}
							}
							fmt.Printf("              Converted: %v (Equity: %.4f)\n",
								convAnalysis, r.DataMoves.Eval[i][6])
						}
					}

					if moveNum >= 10 {
						fmt.Println("\n(Stopping after 10 moves...)")
						goto done
					}
				}
			}
		}
	}

done:
	fmt.Println("\nDone!")
}
