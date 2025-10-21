package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: xgid_cube_parser <xgid-cube-file-or-directory>")
		fmt.Println("Example: xgid_cube_parser tmp/xgid/en/XGID=-aa-B-D-C---bD---c-d--AbbA:0:0:-1:0.txt")
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

	cubeMove, metadata, err := xgparser.ParseXGIDCubeFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		return
	}

	// Print summary
	fmt.Printf("Players: %s (X) vs %s (O)\n", metadata.Player1Name, metadata.Player2Name)
	fmt.Printf("Score: %d-%d (Match to %d)\n", cubeMove.Position.Score[0], cubeMove.Position.Score[1], metadata.MatchLength)
	fmt.Printf("Cube: %d", cubeMove.Position.Cube)
	if cubeMove.Position.CubePos == 1 {
		fmt.Printf(" (owned by X)")
	} else if cubeMove.Position.CubePos == -1 {
		fmt.Printf(" (owned by O)")
	}
	fmt.Println()

	playerName := "X"
	if cubeMove.ActivePlayer == -1 {
		playerName = "O"
	}
	fmt.Printf("On roll: %s (cube action)\n", playerName)
	fmt.Printf("XG Version: %s, MET: %s\n", metadata.ProductVersion, metadata.MET)

	// Print analysis
	if cubeMove.Analysis != nil {
		fmt.Printf("\nAnalysis (Depth: %d-ply):\n", cubeMove.Analysis.AnalysisDepth)
		fmt.Printf("  Player Winning Chances: %.2f%% (G: %.2f%%, BG: %.2f%%)\n",
			cubeMove.Analysis.Player1WinRate*100,
			cubeMove.Analysis.Player1GammonRate*100,
			cubeMove.Analysis.Player1BgRate*100)
		fmt.Printf("  Opponent Winning Chances: %.2f%% (G: %.2f%%, BG: %.2f%%)\n",
			(1.0-cubeMove.Analysis.Player1WinRate)*100,
			cubeMove.Analysis.Player2GammonRate*100,
			cubeMove.Analysis.Player2BgRate*100)

		fmt.Printf("\n  Cubeless Equities:\n")
		fmt.Printf("    No Double: %+.3f\n", cubeMove.Analysis.CubelessNoDouble)
		fmt.Printf("    Double:    %+.3f\n", cubeMove.Analysis.CubelessDouble)

		fmt.Printf("\n  Cubeful Equities:\n")
		fmt.Printf("    No double:     %+.3f\n", cubeMove.Analysis.CubefulNoDouble)
		fmt.Printf("    Double/Take:   %+.3f\n", cubeMove.Analysis.CubefulDoubleTake)
		fmt.Printf("    Double/Pass:   %+.3f\n", cubeMove.Analysis.CubefulDoublePass)

		fmt.Printf("\n  Best Action: ")
		switch cubeMove.CubeAction {
		case 0:
			fmt.Println("No Double")
		case 1:
			fmt.Println("Double/Redouble")
		case 2:
			fmt.Println("Take")
		case 3:
			fmt.Println("Pass")
		default:
			fmt.Println("Unknown")
		}

		if cubeMove.Analysis.WrongPassTakePercent != 0 {
			fmt.Printf("  Wrong Pass/Take: %.1f%%\n", cubeMove.Analysis.WrongPassTakePercent)
		}
	}

	// Optionally output JSON
	if os.Getenv("JSON_OUTPUT") == "1" {
		fmt.Println("\n=== JSON Output ===")
		type OutputData struct {
			CubeMove *xgparser.CubeMove      `json:"cube_move"`
			Metadata *xgparser.MatchMetadata `json:"metadata"`
		}
		output := OutputData{CubeMove: cubeMove, Metadata: metadata}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	}
}
