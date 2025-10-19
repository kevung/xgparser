//package examples

//   stats_example.go - Example of extracting statistics from parsed XG files
//   Copyright (C) 2025 Kevin Unger
//
//   This example demonstrates how to use the xglight parser to extract
//   useful statistics from XG match files.
//

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <xgfile>\n", os.Args[0])
		os.Exit(1)
	}

	xgFilename := os.Args[1]

	// Parse the file
	match, err := xgparser.ParseXGLight(xgFilename)
	if err != nil {
		log.Fatalf("Error parsing file: %v\n", err)
	}

	// Display match information
	fmt.Printf("=== Match Information ===\n")
	fmt.Printf("Players: %s vs %s\n", match.Metadata.Player1Name, match.Metadata.Player2Name)
	fmt.Printf("Event: %s\n", match.Metadata.Event)
	fmt.Printf("Location: %s\n", match.Metadata.Location)
	fmt.Printf("Date: %s\n", match.Metadata.DateTime)
	fmt.Printf("Match Length: %d points\n\n", match.Metadata.MatchLength)

	// Calculate statistics
	totalGames := len(match.Games)
	totalMoves := 0
	totalCheckerMoves := 0
	totalCubeMoves := 0
	player1Wins := 0
	player2Wins := 0

	for _, game := range match.Games {
		totalMoves += len(game.Moves)

		// Count move types
		for _, move := range game.Moves {
			if move.MoveType == "checker" {
				totalCheckerMoves++
			} else if move.MoveType == "cube" {
				totalCubeMoves++
			}
		}

		// Count wins (Winner: -1 = player1, 1 = player2)
		if game.Winner == -1 {
			player1Wins++
		} else if game.Winner == 1 {
			player2Wins++
		}
	}

	fmt.Printf("=== Statistics ===\n")
	fmt.Printf("Total Games: %d\n", totalGames)
	fmt.Printf("Total Moves: %d\n", totalMoves)
	fmt.Printf("Checker Moves: %d\n", totalCheckerMoves)
	fmt.Printf("Cube Decisions: %d\n", totalCubeMoves)
	fmt.Printf("Average Moves per Game: %.1f\n\n", float64(totalMoves)/float64(totalGames))

	fmt.Printf("=== Game Results ===\n")
	fmt.Printf("%s: %d wins\n", match.Metadata.Player1Name, player1Wins)
	fmt.Printf("%s: %d wins\n\n", match.Metadata.Player2Name, player2Wins)

	// Analyze move quality (for games with analysis)
	fmt.Printf("=== Move Quality Analysis ===\n")
	analyzedMoves := 0
	totalEquityLoss := float32(0.0)

	for _, game := range match.Games {
		for _, move := range game.Moves {
			if move.MoveType == "checker" && move.CheckerMove != nil {
				analysis := move.CheckerMove.Analysis
				if len(analysis) > 1 {
					// Compare played move (assumed to be first) with best move
					playedEquity := analysis[0].Equity
					bestEquity := playedEquity

					// Find best equity
					for _, a := range analysis {
						if a.Equity > bestEquity {
							bestEquity = a.Equity
						}
					}

					equityLoss := bestEquity - playedEquity
					if equityLoss > 0 {
						totalEquityLoss += equityLoss
					}
					analyzedMoves++
				}
			}
		}
	}

	if analyzedMoves > 0 {
		fmt.Printf("Analyzed Checker Moves: %d\n", analyzedMoves)
		fmt.Printf("Average Equity Loss: %.4f\n", totalEquityLoss/float32(analyzedMoves))
	} else {
		fmt.Printf("No analyzed moves found in this match.\n")
	}

	// Game-by-game summary
	fmt.Printf("\n=== Game-by-Game Summary ===\n")
	for _, game := range match.Games {
		winner := "Unknown"
		if game.Winner == -1 {
			winner = match.Metadata.Player1Name
		} else if game.Winner == 1 {
			winner = match.Metadata.Player2Name
		}
		fmt.Printf("Game %d: Score %d-%d, %d moves, Winner: %s (%d points)\n",
			game.GameNumber,
			game.InitialScore[0],
			game.InitialScore[1],
			len(game.Moves),
			winner,
			game.PointsWon)
	}
}
