//package xgidparser

//   xgid_parser - Example program to parse XGID position files
//   Copyright (C) 2025 Kevin Unger
//

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: xgid_parser <xgid-file-or-directory>")
		fmt.Println("Example: xgid_parser tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt")
		fmt.Println("Example: xgid_parser tmp/xgid/en")
		os.Exit(1)
	}

	path := os.Args[1]
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		// Parse all files in directory
		parseDirectory(path)
	} else {
		// Parse single file
		parseFile(path)
	}
}

func parseDirectory(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsing %d files from %s\n\n", len(files), dir)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(dir, file.Name())
		parseFile(filePath)
		fmt.Println()
	}
}

func parseFile(filename string) {
	fmt.Printf("=== Parsing: %s ===\n", filename)

	move, metadata, err := xgparser.ParseXGIDFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		return
	}

	// Print summary
	fmt.Printf("Players: %s (X) vs %s (O)\n", metadata.Player1Name, metadata.Player2Name)
	fmt.Printf("Score: %d-%d (Match to %d)\n", move.Position.Score[0], move.Position.Score[1], metadata.MatchLength)
	fmt.Printf("Cube: %d", move.Position.Cube)
	if move.Position.CubePos == 1 {
		fmt.Printf(" (owned by X)")
	} else if move.Position.CubePos == -1 {
		fmt.Printf(" (owned by O)")
	}
	fmt.Println()

	playerName := "X"
	if move.ActivePlayer == -1 {
		playerName = "O"
	}
	fmt.Printf("To play: %s rolls %d%d\n", playerName, move.Dice[0], move.Dice[1])
	fmt.Printf("XG Version: %s, MET: %s\n", metadata.ProductVersion, metadata.MET)

	// Print analysis
	if len(move.Analysis) > 0 {
		fmt.Printf("\nAnalysis (%d moves):\n", len(move.Analysis))
		for i, analysis := range move.Analysis {
			rank := i + 1

			fmt.Printf("  %d. ", rank)
			if analysis.AnalysisDepth == 0 {
				fmt.Printf("[Book] ")
			} else {
				fmt.Printf("[%d-ply] ", analysis.AnalysisDepth)
			}

			// Display move from Move array
			fmt.Printf("%s ", formatMove(analysis.Move))

			fmt.Printf("eq: %+.3f", analysis.Equity)

			// Calculate equity difference from best move
			if i > 0 {
				equityDiff := analysis.Equity - move.Analysis[0].Equity
				fmt.Printf(" (%+.3f)", equityDiff)
			}
			fmt.Println()
			fmt.Printf("     Player: %.2f%% (G: %.2f%%, BG: %.2f%%)\n",
				analysis.Player1WinRate*100, analysis.Player1GammonRate*100, analysis.Player1BgRate*100)
			fmt.Printf("     Opponent: %.2f%% (G: %.2f%%, BG: %.2f%%)\n",
				(1.0-analysis.Player1WinRate)*100, analysis.Player2GammonRate*100, analysis.Player2BgRate*100)

			// Only print first 3 moves in summary, unless verbose mode
			if i >= 2 && len(move.Analysis) > 3 {
				fmt.Printf("  ... and %d more moves\n", len(move.Analysis)-3)
				break
			}
		}
	}

	// Optionally output JSON for the first file
	if os.Getenv("JSON_OUTPUT") == "1" {
		fmt.Println("\n=== JSON Output ===")
		type OutputData struct {
			Move     *xgparser.CheckerMove   `json:"move"`
			Metadata *xgparser.MatchMetadata `json:"metadata"`
		}
		output := OutputData{Move: move, Metadata: metadata}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	}

	// Display CheckerMove structure details
	fmt.Println("\n=== CheckerMove Structure ===")
	fmt.Printf("Active Player: %d\n", move.ActivePlayer)
	fmt.Printf("Dice: [%d, %d]\n", move.Dice[0], move.Dice[1])
	fmt.Printf("Position.Cube: %d\n", move.Position.Cube)
	fmt.Printf("Position.CubePos: %d\n", move.Position.CubePos)
	fmt.Printf("Position.Score: [%d, %d]\n", move.Position.Score[0], move.Position.Score[1])
	fmt.Printf("Analysis entries: %d\n", len(move.Analysis))
}

// formatMove converts a Move array back to human-readable notation
func formatMove(moveArray [8]int8) string {
	var parts []string

	for i := 0; i < 8; i += 2 {
		if moveArray[i] == -1 {
			break
		}

		from := formatPoint(moveArray[i])
		to := formatPoint(moveArray[i+1])
		parts = append(parts, fmt.Sprintf("%s/%s", from, to))
	}

	if len(parts) == 0 {
		return "[no move]"
	}

	return strings.Join(parts, " ")
}

// formatPoint converts a point number to string representation
func formatPoint(point int8) string {
	switch point {
	case 25:
		return "Bar"
	case -2:
		return "Off"
	case -1:
		return ""
	default:
		return fmt.Sprintf("%d", point)
	}
}
