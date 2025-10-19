//
//   main.go - XG lightweight parser command-line tool
//   Copyright (C) 2025 Kevin Unger
//
//   This program demonstrates parsing XG files into a lightweight
//   structure suitable for database integration.
//

package main

import (
	"fmt"
	"os"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <xgfile>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nThis tool parses an XG file and outputs a lightweight JSON representation\n")
		fmt.Fprintf(os.Stderr, "suitable for database integration.\n")
		os.Exit(1)
	}

	xgFilename := os.Args[1]
	fmt.Fprintf(os.Stderr, "Processing file: %s\n\n", xgFilename)

	// Parse the file
	match, err := xgparser.ParseXGLight(xgFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Convert to JSON
	jsonData, err := match.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	// Output JSON to stdout
	fmt.Println(string(jsonData))

	// Print summary to stderr
	fmt.Fprintf(os.Stderr, "\n=== Match Summary ===\n")
	fmt.Fprintf(os.Stderr, "Player 1: %s\n", match.Metadata.Player1Name)
	fmt.Fprintf(os.Stderr, "Player 2: %s\n", match.Metadata.Player2Name)
	fmt.Fprintf(os.Stderr, "Event: %s\n", match.Metadata.Event)
	fmt.Fprintf(os.Stderr, "Location: %s\n", match.Metadata.Location)
	fmt.Fprintf(os.Stderr, "Round: %s\n", match.Metadata.Round)
	fmt.Fprintf(os.Stderr, "Date: %s\n", match.Metadata.DateTime)
	fmt.Fprintf(os.Stderr, "Match Length: %d\n", match.Metadata.MatchLength)
	fmt.Fprintf(os.Stderr, "Number of Games: %d\n", len(match.Games))

	totalMoves := 0
	for _, game := range match.Games {
		totalMoves += len(game.Moves)
	}
	fmt.Fprintf(os.Stderr, "Total Moves: %d\n", totalMoves)
}
