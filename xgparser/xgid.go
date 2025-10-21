//
//   xgid.go - XGID position file parsing module
//   Copyright (C) 2025 Kevin Unger
//
//   This library is free software; you can redistribute it and/or
//   modify it under the terms of the GNU Lesser General Public
//   License as published by the Free Software Foundation; either
//   version 2.1 of the License, or (at your option) any later version.
//
//   This library is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//   Lesser General Public License for more details.
//
//   You should have received a copy of the GNU Lesser General Public
//   License along with this library; if not, write to the Free Software
//   Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301
//   USA
//

package xgparser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// XGIDComponents represents the parsed components of an XGID string
// This is kept for low-level XGID parsing utilities
type XGIDComponents struct {
	PositionID   string // Position part (e.g., "----BaC-B---aD--aa-bcbbBbB")
	CubeOwner    int32  // 0=centered, 1=X owns, -1=O owns
	CubeValue    int32  // Cube value as power of 2 (0=1, 1=2, 2=4, etc.)
	PlayerToMove int32  // 1=X to move, -1=O to move
	Dice         string // Dice as string (e.g., "22", "51", "00" for no dice)
	ScoreX       int32  // X's score
	ScoreO       int32  // O's score
	CrawfordFlag int32  // 0=not crawford, 1=crawford
	MatchLength  int32  // Match length
	MaxCube      int32  // Maximum cube value (usually 10)
}

// ParseXGID parses an XGID string into its components
// Format: XGID=position:cubeOwner:cubeValue:playerToMove:dice:scoreX:scoreO:crawford:matchLength:maxCube
func ParseXGID(xgid string) (*XGIDComponents, error) {
	// Remove "XGID=" prefix if present
	xgid = strings.TrimPrefix(xgid, "XGID=")

	parts := strings.Split(xgid, ":")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid XGID format: expected at least 5 colon-separated parts, got %d", len(parts))
	}

	components := &XGIDComponents{
		PositionID: parts[0],
	}

	// Parse numeric components
	if len(parts) > 1 {
		val, _ := strconv.ParseInt(parts[1], 10, 32)
		components.CubeOwner = int32(val)
	}
	if len(parts) > 2 {
		val, _ := strconv.ParseInt(parts[2], 10, 32)
		components.CubeValue = int32(val)
	}
	if len(parts) > 3 {
		val, _ := strconv.ParseInt(parts[3], 10, 32)
		components.PlayerToMove = int32(val)
	}
	if len(parts) > 4 {
		components.Dice = parts[4]
	}
	if len(parts) > 5 {
		val, _ := strconv.ParseInt(parts[5], 10, 32)
		components.ScoreX = int32(val)
	}
	if len(parts) > 6 {
		val, _ := strconv.ParseInt(parts[6], 10, 32)
		components.ScoreO = int32(val)
	}
	if len(parts) > 7 {
		val, _ := strconv.ParseInt(parts[7], 10, 32)
		components.CrawfordFlag = int32(val)
	}
	if len(parts) > 8 {
		val, _ := strconv.ParseInt(parts[8], 10, 32)
		components.MatchLength = int32(val)
	}
	if len(parts) > 9 {
		val, _ := strconv.ParseInt(parts[9], 10, 32)
		components.MaxCube = int32(val)
	}

	return components, nil
}

// swapMove converts a move from one player's perspective to the other
// This is needed when the XGID position is swapped but the move notation
// is always from the active player's perspective
func swapMove(move [8]int8) [8]int8 {
	var swapped [8]int8
	for i := 0; i < 8; i++ {
		swapped[i] = -1 // Initialize to unused
	}

	for i := 0; i < 8; i += 2 {
		if move[i] == -1 {
			break // End of move
		}

		from := move[i]
		to := move[i+1]

		// Swap point numbers (1 becomes 24, 24 becomes 1, etc.)
		// Special cases: 25 (bar) stays 25, -2 (bear off) stays -2
		if from == 25 {
			swapped[i] = 25 // Bar stays bar
		} else if from >= 1 && from <= 24 {
			swapped[i] = 25 - from
		} else {
			swapped[i] = from
		}

		if to == -2 {
			swapped[i+1] = -2 // Bear off stays bear off
		} else if to == 25 {
			swapped[i+1] = 25 // Bar stays bar
		} else if to >= 1 && to <= 24 {
			swapped[i+1] = 25 - to
		} else {
			swapped[i+1] = to
		}
	}

	return swapped
}

// ApplyMove applies a checker move to a position and returns the resulting position
// move is an array of [from, to, from, to, ...] pairs
// activePlayer indicates whose turn it is (1 for X, -1 for O)
func ApplyMove(pos Position, move [8]int8, activePlayer int32) Position {
	// Make a copy of the position
	result := pos
	checkers := result.Checkers

	// Apply each from/to pair in the move
	for i := 0; i < 8; i += 2 {
		if move[i] == -1 || move[i+1] == -1 {
			break // End of move
		}

		from := int(move[i])
		to := int(move[i+1])

		// Remove checker from source position
		if from >= 0 && from < 26 {
			if activePlayer == 1 {
				// Player X (positive checkers)
				if checkers[from] > 0 {
					checkers[from]--
				}
			} else {
				// Player O (negative checkers)
				if checkers[from] < 0 {
					checkers[from]++
				}
			}
		}

		// Place checker at destination position (unless bearing off)
		if to == -2 {
			// Bearing off - checker removed from board
			continue
		}
		if to >= 0 && to < 26 {
			if activePlayer == 1 {
				// Player X (positive checkers)
				// Check if opponent has a blot at destination
				if checkers[to] == -1 {
					checkers[to] = 0 // Remove opponent's blot
					checkers[0]--    // Place opponent checker on bar (opponent's bar is at index 0)
				}
				checkers[to]++
			} else {
				// Player O (negative checkers)
				// Check if opponent has a blot at destination
				if checkers[to] == 1 {
					checkers[to] = 0 // Remove opponent's blot
					checkers[25]++   // Place opponent checker on bar (opponent's bar is at index 25)
				}
				checkers[to]--
			}
		}
	}

	result.Checkers = checkers
	return result
}

// ParseXGIDFile parses an XGID position file and returns a CheckerMove with metadata
// Returns the unified CheckerMove structure and MatchMetadata
// For cube decisions, use ParseXGIDCubeFile instead
func ParseXGIDFile(filename string) (*CheckerMove, *MatchMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	return ParseXGIDFromReader(file)
}

// DetectXGIDFileType detects whether an XGID file contains checker moves or cube decisions
// Returns "checker" or "cube"
func DetectXGIDFileType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "cube action") {
			return "cube", nil
		}
		if strings.Contains(line, "to play") || strings.Contains(line, "à jouer") ||
			strings.Contains(line, "zu spielen") || strings.Contains(line, "para jugar") ||
			strings.Contains(line, "をプレイ") {
			return "checker", nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "unknown", nil
}

// ParseXGIDFromReader parses an XGID position from an io.Reader
// Returns the unified CheckerMove structure and MatchMetadata
func ParseXGIDFromReader(r io.Reader) (*CheckerMove, *MatchMetadata, error) {
	scanner := bufio.NewScanner(r)

	metadata := &MatchMetadata{}
	move := &CheckerMove{
		Analysis: make([]CheckerAnalysis, 0),
	}

	var boardLines []string
	inBoard := false
	inAnalysis := false
	var xgidString string // Store the full XGID string for reference

	// Regex patterns (language-independent where possible)
	xgidRegex := regexp.MustCompile(`^XGID=([^:]+(?::[^:]+)*)`)
	playersRegex := regexp.MustCompile(`^X:(\S+)\s+O:(\S+)`)

	// Multi-language patterns for score/match
	// English: "Score is X:2 O:3 13 pt.(s) match"
	// French: "Le score est X:0 O:0 match en 13 pt(s)"
	// German: "Spielstand ist S:0 G:0 13 Punkte(e) Match"
	// Spanish: "La puntuación es X:0 O:0 13 pt.(s) partida"
	// Italian: "Il Punteggio è X:0 O:0. Partita ai 13 punto/i"
	// Finnish: "Tulos on X:0 O:0 13 pt. ottelu"
	// Greek: "Το σκορ είναι X:0 O:0 13 pt.(s) παρτίδα"
	// Russian: "Показатель X: 0 O: 0 13 Pt (S) совпадают"
	scoreRegex := regexp.MustCompile(`(?:Score|score|Spielstand|puntuación|Punteggio|Tulos|σκορ|Показатель)[^X]*[XS]:\s*(\d+)\s+[OG]:\s*(\d+)\s+(\d+)\s+`)

	// Multi-language patterns for cube
	// English: "Cube: 2"
	// French: "Videau: 2" or "Cube: 2"
	// German: "Doppler: 2"
	// Spanish: "Cubo: 2"
	// Italian: "Cubo: 2"
	// Finnish: "Kuutio: 2"
	// Greek: "Βίδος: 2"
	// Russian: "Куб: 2"
	// Japanese: "キューブ: 2"
	cubeRegex := regexp.MustCompile(`(?:Cube|Cubo|Videau|Doppler|Dado|Kuutio|Βίδος|Куб|キューブ):\s*(\d+)`)

	// Multi-language patterns for player to move and dice
	// English: "X to play 22"
	// French: "X à jouer 51"
	// German: "X zu spielen 51"
	// Spanish: "X para jugar 51"
	// Italian: "X gioca 51" or "X da giocare 51"
	// Finnish: "X heitti 51"
	// Greek: "X να παίξει51" (note: may have no space before dice!)
	// Russian: "X играть 51"
	// Japanese: "X をプレイ 51"
	toPlayRegex := regexp.MustCompile(`([XO])\s+(?:to play|à jouer|zu spielen|para jugar|da giocare|gioca|heitti|να παίξει|играть|をプレイ)\s*(\d+)`)

	// Analysis line pattern - using simpler positional approach
	// Format: "    1. <depth>      <move notation>              eq:<value>"
	// The rank number starts the line, and equity appears at the end with "eq:" or "éq:" or "экв:"
	// We'll parse this more robustly by looking for these markers rather than specific depth keywords
	analysisLineRegex := regexp.MustCompile(`^\s+(\d+)\.\s+(.+?)\s+(?:eq|éq|экв):([+-]?\d+\.\d+)(?:\s+\(([+-]?\d+\.\d+)\))?`)

	// Ply depth extraction patterns (to extract from the depth field if present)
	plyRegex := regexp.MustCompile(`^(\d+)[-\s](?:ply|plis|Züge|полухода|полухода)`)

	// Win/gammon/bg rate patterns - LANGUAGE INDEPENDENT
	// All statistics lines follow the format: "  <label>: <win%> (G:<gammon%> B:<bg%>)"
	// The label can be in any language, so we match the structure instead of specific words
	// Pattern: starts with whitespace, has a colon, followed by percentage and (G:...% B:...%)
	// First stats line after analysis = Player, second = Opponent
	statsRegex := regexp.MustCompile(`^\s+.+?:\s+(\d+\.\d+)%\s+\(G:\s*(\d+\.\d+)%\s+B:\s*(\d+\.\d+)%\)`)

	// Version and MET pattern
	// "eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2"
	// Spanish: "eXtreme Gammon Versión: ..."
	// Finnish: "eXtreme Gammon Versio: ..."
	versionRegex := regexp.MustCompile(`eXtreme Gammon (?:Version|Versión|Versione|Versio|versio|Έκδοση|Версия|バージョン):\s+([^,]+),\s+(?:MET|TEM):\s+(.+)`)

	// Board boundary markers
	boardTopRegex := regexp.MustCompile(`^\s+\+13-14-15-16-17-18------19-20-21-22-23-24-\+`)
	boardBottomRegex := regexp.MustCompile(`^\s+\+12-11-10--9--8--7-------6--5--4--3--2--1-\+`)

	// Temporary storage for current move being parsed
	var currentAnalysis *CheckerAnalysis
	var xgidComponents XGIDComponents // Store parsed XGID components

	for scanner.Scan() {
		line := scanner.Text()

		// Parse XGID line
		if matches := xgidRegex.FindStringSubmatch(line); matches != nil {
			xgidString = matches[1]
			components, err := ParseXGID(xgidString)
			if err == nil {
				xgidComponents = *components
			}
			continue
		}

		// Parse player names
		if matches := playersRegex.FindStringSubmatch(line); matches != nil {
			metadata.Player1Name = matches[1]
			metadata.Player2Name = matches[2]
			continue
		}

		// Parse score and match length
		if matches := scoreRegex.FindStringSubmatch(line); matches != nil {
			scoreX, _ := strconv.ParseInt(matches[1], 10, 32)
			scoreO, _ := strconv.ParseInt(matches[2], 10, 32)
			matchLength, _ := strconv.ParseInt(matches[3], 10, 32)
			move.Position.Score = [2]int32{int32(scoreX), int32(scoreO)}
			metadata.MatchLength = int32(matchLength)
			continue
		}

		// Parse cube value
		if matches := cubeRegex.FindStringSubmatch(line); matches != nil {
			cube, _ := strconv.ParseInt(matches[1], 10, 32)
			move.Position.Cube = int32(cube)
			continue
		}

		// Parse player to move and dice
		if matches := toPlayRegex.FindStringSubmatch(line); matches != nil {
			if matches[1] == "X" {
				move.ActivePlayer = 1
			} else {
				move.ActivePlayer = -1
			}
			dice := matches[2]
			if len(dice) == 2 {
				d1, _ := strconv.ParseInt(string(dice[0]), 10, 32)
				d2, _ := strconv.ParseInt(string(dice[1]), 10, 32)
				move.Dice = [2]int32{int32(d1), int32(d2)}
			}
			inAnalysis = true
			continue
		}

		// Track board diagram
		if boardTopRegex.MatchString(line) {
			inBoard = true
			boardLines = []string{line}
			continue
		}
		if inBoard {
			boardLines = append(boardLines, line)
			if boardBottomRegex.MatchString(line) {
				inBoard = false
				// Board diagram is optional - we don't need to store it in metadata
			}
			continue
		}

		// Parse analysis lines
		if inAnalysis {
			if matches := analysisLineRegex.FindStringSubmatch(line); matches != nil {
				// Save previous move if exists
				if currentAnalysis != nil {
					move.Analysis = append(move.Analysis, *currentAnalysis)
				}

				// matches[1] = rank number
				// matches[2] = everything between rank and "eq:" (depth + move notation)
				// matches[3] = equity value
				// matches[4] = equity difference (optional)

				fullMiddle := matches[2]
				equity, _ := strconv.ParseFloat(matches[3], 64)

				// Split the middle part into depth field and move notation
				// The depth field is roughly the first 13 characters (columns 7-20)
				// But we need to be flexible. Strategy: split by multiple spaces
				parts := strings.SplitN(fullMiddle, "  ", 2) // Split on 2+ spaces

				var depthField, moveNotation string
				if len(parts) >= 2 {
					depthField = strings.TrimSpace(parts[0])
					moveNotation = strings.TrimSpace(parts[1])
				} else {
					// Fallback: no clear separator, try to extract from position
					if len(fullMiddle) > 13 {
						depthField = strings.TrimSpace(fullMiddle[:13])
						moveNotation = strings.TrimSpace(fullMiddle[13:])
					} else {
						depthField = strings.TrimSpace(fullMiddle)
						moveNotation = ""
					}
				}

				// Extract ply depth from depth field
				ply := 0
				if plyMatches := plyRegex.FindStringSubmatch(depthField); plyMatches != nil {
					ply, _ = strconv.Atoi(plyMatches[1])
				}
				// If no ply found, it's either Book/Livre/XG Roller/etc (depth stays 0)

				// Parse move notation into Move array
				moveArray := ParseMoveNotation(moveNotation)

				// Create CheckerAnalysis structure
				currentAnalysis = &CheckerAnalysis{
					Position:      Position{}, // Will be filled with resulting position
					Move:          moveArray,  // Parsed from notation
					Equity:        float32(equity),
					AnalysisDepth: int16(ply),
				}
				continue
			}

			// Parse statistics lines (Player and Opponent)
			// Both follow the same format, we use order to distinguish:
			// First stats line after analysis = Player, second = Opponent
			if currentAnalysis != nil {
				if matches := statsRegex.FindStringSubmatch(line); matches != nil {
					winRate, _ := strconv.ParseFloat(matches[1], 64)
					gammonRate, _ := strconv.ParseFloat(matches[2], 64)
					bgRate, _ := strconv.ParseFloat(matches[3], 64)

					// If Player1WinRate is still 0, this is the player line
					if currentAnalysis.Player1WinRate == 0 {
						currentAnalysis.Player1WinRate = float32(winRate / 100.0)
						currentAnalysis.Player1GammonRate = float32(gammonRate / 100.0)
						currentAnalysis.Player1BgRate = float32(bgRate / 100.0)
					} else {
						// Otherwise, this is the opponent line
						currentAnalysis.Player2GammonRate = float32(gammonRate / 100.0)
						currentAnalysis.Player2BgRate = float32(bgRate / 100.0)
					}
					continue
				}
			}
		}

		// Parse version and MET
		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			metadata.ProductVersion = matches[1]
			metadata.MET = matches[2]

			// Save last move if exists
			if currentAnalysis != nil {
				move.Analysis = append(move.Analysis, *currentAnalysis)
				currentAnalysis = nil
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// Save any remaining analysis that wasn't saved yet
	if currentAnalysis != nil {
		move.Analysis = append(move.Analysis, *currentAnalysis)
	}

	// Convert XGID position to checker array and populate position
	if xgidComponents.PositionID != "" {
		move.Position.Checkers = XGIDToPosition(xgidComponents.PositionID)

		// Set cube position based on cube owner
		if xgidComponents.CubeOwner == 1 {
			move.Position.CubePos = 1 // X owns
		} else if xgidComponents.CubeOwner == -1 {
			move.Position.CubePos = -1 // O owns
		} else {
			move.Position.CubePos = 0 // centered
		}

		// Calculate actual cube value (2^cubeValue) if not already set
		if move.Position.Cube == 0 && xgidComponents.CubeValue >= 0 {
			cubeValue := int32(1)
			for i := int32(0); i < xgidComponents.CubeValue; i++ {
				cubeValue *= 2
			}
			move.Position.Cube = cubeValue
		}

		// XGID positions are stored from the perspective of playerToMove in the XGID
		// If the text file indicates a different active player, we need to swap
		// Example: XGID says playerToMove=-1 (O's perspective) but text says "X to play"
		// This happens because XG stores positions consistently but can export from either perspective
		//
		// IMPORTANT: Move notation in XGID text files is ALWAYS from the active player's perspective
		// (the player indicated in "X to play" or "O on roll"), NOT the XGID playerToMove perspective.
		// So when we swap the position, we must also swap the move coordinates.
		needsSwap := xgidComponents.PlayerToMove != move.ActivePlayer
		if needsSwap {
			move.Position = swapPosition(move.Position)
		}

		// Compute final positions by applying each move to the initial position
		for i := range move.Analysis {
			moveToApply := move.Analysis[i].Move
			// If we swapped the position, we need to swap the move coordinates too
			// because the move notation is always from the active player's perspective
			if needsSwap {
				moveToApply = swapMove(moveToApply)
				// Also update the stored move to the swapped version
				move.Analysis[i].Move = moveToApply
			}
			move.Analysis[i].Position = ApplyMove(move.Position, moveToApply, move.ActivePlayer)
		}
	}

	return move, metadata, nil
}

// ParseMoveNotation converts human-ireadable move notation to Move array
// Format examples: "Bar/21 16/10", "24/23 13/8", "8/5(2) 6/5(2)", "Bar/23(2) 13/11(2)"
// Returns: [8]int8 array where pairs represent from/to positions
//
//	25 = bar, 1-24 = points, -2 = bear off, -1 = unused
func ParseMoveNotation(notation string) [8]int8 {
	var move [8]int8
	// Initialize all positions to -1 (unused)
	for i := range move {
		move[i] = -1
	}

	if notation == "" {
		return move
	}

	// Split by spaces to get individual moves
	parts := strings.Fields(notation)
	moveIndex := 0

	for _, part := range parts {
		// Handle multiplier notation like "8/5(2)"
		multiplier := 1
		if idx := strings.Index(part, "("); idx != -1 {
			multStr := part[idx+1 : len(part)-1]
			multiplier, _ = strconv.Atoi(multStr)
			part = part[:idx]
		}

		// Parse from/to notation
		if !strings.Contains(part, "/") {
			continue
		}

		fromTo := strings.Split(part, "/")
		if len(fromTo) != 2 {
			continue
		}

		// Parse 'from' position
		var from int8
		if strings.ToLower(fromTo[0]) == "bar" {
			from = 25
		} else {
			val, err := strconv.Atoi(fromTo[0])
			if err != nil {
				continue
			}
			from = int8(val)
		}

		// Parse 'to' position
		var to int8
		if strings.ToLower(fromTo[1]) == "off" {
			to = -2
		} else if strings.ToLower(fromTo[1]) == "bar" {
			to = 25
		} else {
			val, err := strconv.Atoi(fromTo[1])
			if err != nil {
				continue
			}
			to = int8(val)
		}

		// Add the move (with multiplier)
		for i := 0; i < multiplier && moveIndex < 8; i++ {
			move[moveIndex] = from
			move[moveIndex+1] = to
			moveIndex += 2
		}
	}

	return move
}

// XGIDToPosition converts an XGID position string to a checker array
// XGID format uses base-64 encoding: '-' = 0, 'A'=1, 'B'=2, ..., 'Z'=26, 'a'=27, ..., 'o'=40
// Lowercase letters represent checkers for player O (negative in our format)
func XGIDToPosition(positionID string) [26]int8 {
	var position [26]int8

	if len(positionID) != 26 {
		return position // Invalid position
	}

	for i := 0; i < 26; i++ {
		char := positionID[i]
		var count int8

		if char == '-' {
			count = 0
		} else if char >= 'A' && char <= 'Z' {
			// X player checkers (positive)
			count = int8(char - 'A' + 1)
		} else if char >= 'a' && char <= 'o' {
			// O player checkers (negative)
			count = -int8(char - 'a' + 1)
		}

		// XGID positions are in reverse order: index 0 = point 24, index 23 = point 1
		// XGID indices 24-25 are the bar positions
		// Our internal format: index 0 = opponent bar, index 1-24 = points, index 25 = player bar
		//
		// IMPORTANT: XGID always encodes from X's absolute perspective:
		// - XGID index 24 = O's checkers on the bar (lowercase letters)
		// - XGID index 25 = X's checkers on the bar (uppercase letters)
		//
		// Our internal format is from the active player's perspective:
		// - position[25] = active player's checkers on bar
		// - position[0] = opponent's checkers on bar
		//
		// So when parsing an XGID where playerToMove=1 (X):
		// - XGID[25] (X's bar) -> position[25] (player's bar)
		// - XGID[24] (O's bar) -> position[0] (opponent's bar)
		//
		// When playerToMove=-1 (O), this will be swapped later by swapPosition()
		if i < 24 {
			position[24-i] = count // REVERSED: index 0 -> point 24, index 23 -> point 1
		} else if i == 24 {
			position[0] = count // XGID index 24 is the bar for O
		} else if i == 25 {
			position[25] = count // XGID index 25 is the bar for X
		}
	}

	return position
}

// ParseXGIDCubeFile parses an XGID cube decision file and returns a CubeMove with metadata
// Returns the unified CubeMove structure and MatchMetadata
func ParseXGIDCubeFile(filename string) (*CubeMove, *MatchMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	return ParseXGIDCubeFromReader(file)
}

// ParseXGIDCubeFromReader parses an XGID cube decision from an io.Reader
// Returns the unified CubeMove structure and MatchMetadata
func ParseXGIDCubeFromReader(r io.Reader) (*CubeMove, *MatchMetadata, error) {
	scanner := bufio.NewScanner(r)

	metadata := &MatchMetadata{}
	cubeMove := &CubeMove{
		Analysis: &CubeAnalysis{},
	}

	var xgidString string

	// Regex patterns
	xgidRegex := regexp.MustCompile(`^XGID=([^:]+(?::[^:]+)*)`)
	playersRegex := regexp.MustCompile(`^X:(\S+)\s+O:(\S+)`)
	scoreRegex := regexp.MustCompile(`(?:Score|score|Spielstand|puntuación|Punteggio|Tulos|σκορ|Показатель)[^X]*[XS]:\s*(\d+)\s+[OG]:\s*(\d+)\s+(\d+)\s+`)
	cubeRegex := regexp.MustCompile(`(?:Cube|Cubo|Videau|Doppler|Dado|Kuutio|Βίδος|Куб|キューブ):\s*(\d+)`)

	// Cube action line: "X on roll, cube action" / "O on roll, cube action"
	cubeActionRegex := regexp.MustCompile(`([XO])\s+on roll,\s+cube action`)

	// Analysis depth: "Analyzed in 4-ply"
	analyzedRegex := regexp.MustCompile(`Analyzed in (\d+)-ply`)

	// Player/Opponent winning chances: "Player Winning Chances:   61.89% (G:37.15% B:0.42%)"
	winChancesRegex := regexp.MustCompile(`(Player|Opponent|Joueur|Adversaire|Spieler|Gegner|プレーヤー|対戦相手)\s+(?:Winning Chances|chances de gagner):\s+(\d+\.\d+)%\s+\(G:\s*(\d+\.\d+)%\s+B:\s*(\d+\.\d+)%\)`)

	// Cubeless equities: "Cubeless Equities: No Double=+0.513, Double=+1.048"
	cubelessRegex := regexp.MustCompile(`Cubeless.*?No\s+(?:Double|double|Doublet|redouble)=([+-]?\d+\.\d+).*?(?:Double|double|Doublet|redouble)=([+-]?\d+\.\d+)`)

	// Cubeful equities lines: "       No double:     +0.637 (-0.109)"
	cubefulNoDoubleRegex := regexp.MustCompile(`No\s+(?:double|redouble|Doublet):\s+([+-]?\d+\.\d+)`)
	cubefulDoubleTakeRegex := regexp.MustCompile(`(?:Double|Redouble|Doublet)/(?:Take|Prendre|prendre):\s+([+-]?\d+\.\d+)`)
	cubefulDoublePassRegex := regexp.MustCompile(`(?:Double|Redouble|Doublet)/(?:Pass|Passer|passer):\s+([+-]?\d+\.\d+)`)

	// Best cube action: "Best Cube action: Double / Take"
	bestActionRegex := regexp.MustCompile(`Best.*?:\s*(.*?)$`)

	versionRegex := regexp.MustCompile(`eXtreme Gammon (?:Version|Versión|Versione|Versio|versio|Έκδοση|Версия|バージョン):\s+([^,]+),\s+(?:MET|TEM):\s+(.+)`)

	var xgidComponents XGIDComponents
	playerStatsCollected := false

	for scanner.Scan() {
		line := scanner.Text()

		// Parse XGID line
		if matches := xgidRegex.FindStringSubmatch(line); matches != nil {
			xgidString = matches[1]
			components, err := ParseXGID(xgidString)
			if err == nil {
				xgidComponents = *components
			}
			continue
		}

		// Parse player names
		if matches := playersRegex.FindStringSubmatch(line); matches != nil {
			metadata.Player1Name = matches[1]
			metadata.Player2Name = matches[2]
			continue
		}

		// Parse score and match length
		if matches := scoreRegex.FindStringSubmatch(line); matches != nil {
			scoreX, _ := strconv.ParseInt(matches[1], 10, 32)
			scoreO, _ := strconv.ParseInt(matches[2], 10, 32)
			matchLength, _ := strconv.ParseInt(matches[3], 10, 32)
			cubeMove.Position.Score = [2]int32{int32(scoreX), int32(scoreO)}
			metadata.MatchLength = int32(matchLength)
			continue
		}

		// Parse cube value
		if matches := cubeRegex.FindStringSubmatch(line); matches != nil {
			cube, _ := strconv.ParseInt(matches[1], 10, 32)
			cubeMove.Position.Cube = int32(cube)
			continue
		}

		// Parse player on roll and cube action
		if matches := cubeActionRegex.FindStringSubmatch(line); matches != nil {
			if matches[1] == "X" {
				cubeMove.ActivePlayer = 1
			} else {
				cubeMove.ActivePlayer = -1
			}
			continue
		}

		// Parse analysis depth
		if matches := analyzedRegex.FindStringSubmatch(line); matches != nil {
			depth, _ := strconv.ParseInt(matches[1], 10, 32)
			cubeMove.Analysis.AnalysisDepth = int32(depth)
			continue
		}

		// Parse winning chances
		if matches := winChancesRegex.FindStringSubmatch(line); matches != nil {
			winRate, _ := strconv.ParseFloat(matches[2], 64)
			gammonRate, _ := strconv.ParseFloat(matches[3], 64)
			bgRate, _ := strconv.ParseFloat(matches[4], 64)

			// First match is Player, second is Opponent
			if !playerStatsCollected {
				cubeMove.Analysis.Player1WinRate = float32(winRate / 100.0)
				cubeMove.Analysis.Player1GammonRate = float32(gammonRate / 100.0)
				cubeMove.Analysis.Player1BgRate = float32(bgRate / 100.0)
				playerStatsCollected = true
			} else {
				cubeMove.Analysis.Player2GammonRate = float32(gammonRate / 100.0)
				cubeMove.Analysis.Player2BgRate = float32(bgRate / 100.0)
			}
			continue
		}

		// Parse cubeless equities
		if matches := cubelessRegex.FindStringSubmatch(line); matches != nil {
			noDouble, _ := strconv.ParseFloat(matches[1], 64)
			double, _ := strconv.ParseFloat(matches[2], 64)
			cubeMove.Analysis.CubelessNoDouble = float32(noDouble)
			cubeMove.Analysis.CubelessDouble = float32(double)
			continue
		}

		// Parse cubeful equities
		if matches := cubefulNoDoubleRegex.FindStringSubmatch(line); matches != nil {
			eq, _ := strconv.ParseFloat(matches[1], 64)
			cubeMove.Analysis.CubefulNoDouble = float32(eq)
			continue
		}
		if matches := cubefulDoubleTakeRegex.FindStringSubmatch(line); matches != nil {
			eq, _ := strconv.ParseFloat(matches[1], 64)
			cubeMove.Analysis.CubefulDoubleTake = float32(eq)
			continue
		}
		if matches := cubefulDoublePassRegex.FindStringSubmatch(line); matches != nil {
			eq, _ := strconv.ParseFloat(matches[1], 64)
			cubeMove.Analysis.CubefulDoublePass = float32(eq)
			continue
		}

		// Parse best cube action
		if matches := bestActionRegex.FindStringSubmatch(line); matches != nil {
			if strings.Contains(line, "Best") {
				action := strings.ToLower(matches[1])
				if strings.Contains(action, "double") && strings.Contains(action, "take") {
					cubeMove.CubeAction = 2 // Take
				} else if strings.Contains(action, "double") && strings.Contains(action, "pass") {
					cubeMove.CubeAction = 3 // Pass
				} else if strings.Contains(action, "double") || strings.Contains(action, "redouble") {
					cubeMove.CubeAction = 1 // Double
				} else if strings.Contains(action, "take") || strings.Contains(action, "pass") {
					// Response to opponent's double
					if strings.Contains(action, "take") {
						cubeMove.CubeAction = 2
					} else {
						cubeMove.CubeAction = 3
					}
				} else {
					cubeMove.CubeAction = 0 // No double
				}
			}
			continue
		}

		// Parse version and MET
		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			metadata.ProductVersion = matches[1]
			metadata.MET = matches[2]
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// Convert XGID position to checker array and populate position
	if xgidComponents.PositionID != "" {
		cubeMove.Position.Checkers = XGIDToPosition(xgidComponents.PositionID)

		// Set cube position based on cube owner
		if xgidComponents.CubeOwner == 1 {
			cubeMove.Position.CubePos = 1 // X owns
		} else if xgidComponents.CubeOwner == -1 {
			cubeMove.Position.CubePos = -1 // O owns
		} else {
			cubeMove.Position.CubePos = 0 // centered
		}
	}

	// Calculate wrong pass/take percentage if we have the data
	if cubeMove.Analysis.CubefulDoubleTake != 0 && cubeMove.Analysis.CubefulDoublePass != 0 {
		cubeMove.Analysis.WrongPassTakePercent = (cubeMove.Analysis.CubefulDoubleTake - cubeMove.Analysis.CubefulDoublePass) * 100
	}

	return cubeMove, metadata, nil
}
