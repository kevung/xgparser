//
//   xglight.go - XG lightweight parsing module
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
	"bytes"
	"encoding/json"
	"io"
)

// MatchMetadata contains essential match information
type MatchMetadata struct {
	Player1Name    string `json:"player1_name"`
	Player2Name    string `json:"player2_name"`
	Location       string `json:"location"`
	Event          string `json:"event"`
	Round          string `json:"round"`
	DateTime       string `json:"date_time"`
	MatchLength    int32  `json:"match_length"`
	EngineVersion  int32  `json:"engine_version"`  // File format version (e.g., 30)
	ProductVersion string `json:"product_version"` // XG product version (e.g., "eXtreme Gammon 2.19.1")
}

// Position represents a backgammon position
type Position struct {
	Checkers [26]int8 `json:"checkers"` // Position of checkers
	Cube     int32    `json:"cube"`     // Cube value
	CubePos  int32    `json:"cube_pos"` // Cube position (0=center, 1=player1, -1=player2)
	Score    [2]int32 `json:"score"`    // Match score [player1, player2]
}

// CheckerAnalysis contains analysis for a single checker move
// Note: player1/player2 in analysis refer to the player on roll (active_player) and their opponent,
// not the players in player1_name/player2_name metadata
type CheckerAnalysis struct {
	Position          Position `json:"position"`            // Resulting position
	Move              [8]int8  `json:"move"`                // The move itself (25=bar, 1-24=points, -2=bear off, -1=unused)
	Player1WinRate    float32  `json:"player1_win_rate"`    // Win rate for player on roll (1 - eval[2])
	Player1GammonRate float32  `json:"player1_gammon_rate"` // Gammon rate for player on roll (eval[4])
	Player1BgRate     float32  `json:"player1_bg_rate"`     // Backgammon rate for player on roll (eval[5])
	Player2GammonRate float32  `json:"player2_gammon_rate"` // Gammon rate for opponent (eval[1])
	Player2BgRate     float32  `json:"player2_bg_rate"`     // Backgammon rate for opponent (eval[0])
	Equity            float32  `json:"equity"`              // eval[6] - normalized equity
	AnalysisDepth     int16    `json:"analysis_depth"`      // EvalLevel.Level
}

// CubeAnalysis contains analysis for a cube decision
// Note: For cube decisions, eval is always from active player's perspective
// player1 in analysis = player on roll, player2 = opponent (no swap needed)
type CubeAnalysis struct {
	Player1WinRate       float32 `json:"player1_win_rate"`        // Win rate for player on roll - eval[2]
	Player1GammonRate    float32 `json:"player1_gammon_rate"`     // Gammon rate for player on roll - eval[1]
	Player1BgRate        float32 `json:"player1_bg_rate"`         // Backgammon rate for player on roll - eval[0]
	Player2GammonRate    float32 `json:"player2_gammon_rate"`     // Gammon rate for opponent - eval[4]
	Player2BgRate        float32 `json:"player2_bg_rate"`         // Backgammon rate for opponent - eval[5]
	CubelessNoDouble     float32 `json:"cubeless_no_double"`      // eval[6]
	CubelessDouble       float32 `json:"cubeless_double"`         // eval[7] (if available)
	CubefulNoDouble      float32 `json:"cubeful_no_double"`       // equB
	CubefulDoubleTake    float32 `json:"cubeful_double_take"`     // equDouble
	CubefulDoublePass    float32 `json:"cubeful_double_pass"`     // equDrop
	WrongPassTakePercent float32 `json:"wrong_pass_take_percent"` // Calculated metric
	AnalysisDepth        int32   `json:"analysis_depth"`          // Level
}

// CheckerMove represents a checker play decision
type CheckerMove struct {
	Position     Position          `json:"position"`      // Position before the move
	ActivePlayer int32             `json:"active_player"` // Player making the move
	Dice         [2]int32          `json:"dice"`          // Dice rolled
	PlayedMove   [8]int32          `json:"played_move"`   // The move that was played (25=bar, 1-24=points, -2=bear off, -1=unused)
	Analysis     []CheckerAnalysis `json:"analysis"`      // Analysis of possible moves
}

// CubeMove represents a cube decision
type CubeMove struct {
	Position     Position      `json:"position"`      // Position when cube decision was made
	ActivePlayer int32         `json:"active_player"` // Player making the decision
	CubeAction   int32         `json:"cube_action"`   // 0=no double, 1=double, 2=take, 3=pass
	Analysis     *CubeAnalysis `json:"analysis"`      // Analysis of cube decision
}

// Move represents either a checker or cube move
type Move struct {
	MoveType    string       `json:"move_type"` // "checker" or "cube"
	CheckerMove *CheckerMove `json:"checker_move,omitempty"`
	CubeMove    *CubeMove    `json:"cube_move,omitempty"`
}

// Game represents a single game within a match
type Game struct {
	GameNumber   int32    `json:"game_number"`
	InitialScore [2]int32 `json:"initial_score"` // Score at start of game
	Moves        []Move   `json:"moves"`
	Winner       int32    `json:"winner"` // -1=player1, 1=player2, 0=not completed
	PointsWon    int32    `json:"points_won"`
}

// Match represents the complete match structure
type Match struct {
	Metadata MatchMetadata `json:"metadata"`
	Games    []Game        `json:"games"`
}

// ToJSON serializes the Match to JSON
func (m *Match) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// ParseXG parses XG file segments and returns a lightweight match structure
// This function accepts already extracted segments, allowing the caller to
// provide data from memory, network, or any other source.
func ParseXG(segments []*Segment) (*Match, error) {
	var match Match
	var currentGame *Game
	fileVersion := int32(-1)

	// Extract product version from GDF header if present
	for _, segment := range segments {
		if segment.Type == SegmentGDFHdr {
			gdfHeader := &GameDataFormatHdrRecord{}
			reader := bytes.NewReader(segment.Data)
			if err := gdfHeader.FromStream(reader); err == nil {
				match.Metadata.ProductVersion = gdfHeader.GameName
			}
			break
		}
	}

	for _, segment := range segments {
		if segment.Type == SegmentXGGameFile {
			records, err := ParseGameFile(segment.Data, fileVersion)
			if err != nil {
				return nil, err
			}

			for _, rec := range records {
				switch r := rec.(type) {
				case *HeaderMatchEntry:
					fileVersion = r.Version
					// Extract match metadata
					match.Metadata = MatchMetadata{
						Player1Name:   getPreferredString(r.Player1, r.SPlayer1),
						Player2Name:   getPreferredString(r.Player2, r.SPlayer2),
						Location:      getPreferredString(r.Location, r.SLocation),
						Event:         getPreferredString(r.Event, r.SEvent),
						Round:         getPreferredString(r.Round, r.SRound),
						DateTime:      r.Date,
						MatchLength:   r.MatchLength,
						EngineVersion: r.Version,
					}

				case *HeaderGameEntry:
					// Start a new game
					currentGame = &Game{
						GameNumber:   r.GameNumber,
						InitialScore: [2]int32{r.Score1, r.Score2},
						Moves:        make([]Move, 0),
					}

				case *CubeEntry:
					if currentGame != nil {
						// Skip initial position cube entries (Double == -2) which don't represent actual cube decisions
						if r.Double != -2 {
							cubeMove := convertCubeEntry(r)
							currentGame.Moves = append(currentGame.Moves, Move{
								MoveType: "cube",
								CubeMove: cubeMove,
							})
						}
					}

				case *MoveEntry:
					if currentGame != nil {
						checkerMove := convertMoveEntry(r)
						currentGame.Moves = append(currentGame.Moves, Move{
							MoveType:    "checker",
							CheckerMove: checkerMove,
						})
					}

				case *FooterGameEntry:
					if currentGame != nil {
						currentGame.Winner = r.Winner
						currentGame.PointsWon = r.PointsWon
						match.Games = append(match.Games, *currentGame)
						currentGame = nil
					}
				}
			}
		}
	}

	return &match, nil
}

// ParseXGFromFile parses an XG file from disk and returns a lightweight match structure
// This is a convenience wrapper around ParseXG that handles file reading.
func ParseXGFromFile(filename string) (*Match, error) {
	imp := NewImport(filename)
	segments, err := imp.GetFileSegments()
	if err != nil {
		return nil, err
	}
	return ParseXG(segments)
}

// ParseXGFromReader parses an XG file from an io.Reader and returns a lightweight match structure
// This allows parsing XG files from network streams, memory buffers, or any io.Reader source.
func ParseXGFromReader(r io.ReadSeeker) (*Match, error) {
	// Read and extract the Game Data Format Header
	gdfHeader := &GameDataFormatHdrRecord{}
	err := gdfHeader.FromStream(r)
	if err != nil {
		return nil, err
	}

	// Get segments using the same logic as Import.GetFileSegments
	r.Seek(0, io.SeekStart)
	var segments []*Segment

	// Read GDF header
	gdfData := make([]byte, gdfHeader.HeaderSize)
	_, err = io.ReadFull(r, gdfData)
	if err != nil {
		return nil, err
	}

	segments = append(segments, &Segment{
		Type: SegmentGDFHdr,
		Data: gdfData,
	})

	// Extract thumbnail if present
	if gdfHeader.ThumbnailSize > 0 {
		r.Seek(gdfHeader.ThumbnailOffset, io.SeekCurrent)
		imgData := make([]byte, gdfHeader.ThumbnailSize)
		_, err = io.ReadFull(r, imgData)
		if err != nil {
			return nil, err
		}
		segments = append(segments, &Segment{
			Type: SegmentGDFImage,
			Data: imgData,
		})
	}

	// Get archive object
	archiveObj, err := NewZlibArchive(r)
	if err != nil {
		return nil, err
	}

	// Process all files in the archive
	for _, fileRec := range archiveObj.ArcRegistry {
		data, err := archiveObj.GetArchiveFile(&fileRec)
		if err != nil {
			return nil, err
		}

		segmentType := XGFileMap[fileRec.Name]
		segments = append(segments, &Segment{
			Type:     segmentType,
			Data:     data,
			Filename: fileRec.Name,
		})
	}

	return ParseXG(segments)
}

// ParseXGLight is deprecated. Use ParseXGFromFile instead.
func ParseXGLight(filename string) (*Match, error) {
	return ParseXGFromFile(filename)
}

// getPreferredString returns the first non-empty string
func getPreferredString(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}

// swapPositionCheckers flips the board checkers from one player's perspective to the other
// Index 0: opponent's bar, Index 1-24: board points, Index 25: player on roll's bar
func swapPositionCheckers(pos [26]int8) [26]int8 {
	var swapped [26]int8
	// Swap bars: player's bar (index 25) becomes opponent's bar (index 0)
	swapped[0] = -pos[25]
	// Points 1-24 are reversed and negated (point 1 becomes point 24, etc.)
	for i := 1; i <= 24; i++ {
		swapped[i] = -pos[25-i]
	}
	// Opponent's bar (index 0) becomes player's bar (index 25)
	swapped[25] = -pos[0]
	return swapped
}

// swapPosition swaps a complete Position from one player's perspective to the other
// This includes swapping checkers, score, and cube position
func swapPosition(pos Position) Position {
	return Position{
		Checkers: swapPositionCheckers(pos.Checkers),
		Cube:     pos.Cube,
		CubePos:  -pos.CubePos,                         // Swap cube position: 1 becomes -1, -1 becomes 1, 0 stays 0
		Score:    [2]int32{pos.Score[1], pos.Score[0]}, // Swap score array
	}
}

// convertCubeEntry converts a full CubeEntry to CubeMove
func convertCubeEntry(c *CubeEntry) *CubeMove {
	// Build initial position
	position := Position{
		Checkers: c.Position,
		Cube:     c.CubeB,
		CubePos:  0,              // Default, could be extracted from more context
		Score:    [2]int32{0, 0}, // Would need to track from game state
	}

	// Swap position to player on roll's perspective only when active_player == -1
	if c.ActiveP == -1 {
		position = swapPosition(position)
	}

	move := &CubeMove{
		Position:     position,
		ActivePlayer: c.ActiveP,
		CubeAction:   c.Double, // Simplified - may need more logic
	}

	// Add cube analysis if available
	if c.Doubled != nil {
		// For cube decisions, Eval is ALWAYS from the active player's perspective (player on roll)
		// Eval[0] = opponent's backgammon rate
		// Eval[1] = opponent's gammon rate
		// Eval[2] = opponent's win rate
		// Eval[4] = player on roll's gammon rate
		// Eval[5] = player on roll's backgammon rate
		// Eval[6] = cubeless equity
		//
		// player1 in our output = player on roll (active_player)
		// player2 in our output = opponent
		var p1Win, p1Gammon, p1Bg, p2Gammon, p2Bg float32

		// XG's Eval[2] is opponent's win rate, so player on roll's win rate is 1 - Eval[2]
		p1Win = 1.0 - c.Doubled.Eval[2] // Player on roll's win rate
		p1Gammon = c.Doubled.Eval[4]    // Player on roll's gammon rate
		p1Bg = c.Doubled.Eval[5]        // Player on roll's backgammon rate
		p2Gammon = c.Doubled.Eval[1]    // Opponent's gammon rate
		p2Bg = c.Doubled.Eval[0]        // Opponent's backgammon rate

		// Wrong pass/take percentage: the probability threshold where making the wrong
		// cube decision (take vs pass after double) results in the same equity as no double.
		// Formula: p * equity_wrong + (1-p) * equity_right = equity_no_double
		// Solving for p: p = (equity_no_double - equity_right) / (equity_wrong - equity_right)
		// Only compute this when cubeful no double equity is the best (highest) equity
		// Use -1.0 to indicate "not applicable" when no double is not the best equity
		wrongPassTakePercent := float32(-1.0)

		equNoDouble := c.Doubled.EquB
		equDoubleTake := c.Doubled.EquDouble
		equDoublePass := c.Doubled.EquDrop

		// Check if no double is the best equity (highest value)
		if equNoDouble >= equDoubleTake && equNoDouble >= equDoublePass {
			// From the opponent's perspective after being doubled:
			// - Taking gives them -equDoubleTake (negated)
			// - Passing gives them -equDoublePass (negated)
			// The correct decision is the one with higher equity for them (less negative)
			// which means lower equity for us (the doubler)
			var equRight, equWrong float32
			if equDoubleTake < equDoublePass {
				// Correct decision for opponent is take (better for them = worse for us = more negative)
				// Wrong decision is pass (worse for them = better for us = more positive)
				equRight = equDoubleTake
				equWrong = equDoublePass
			} else {
				// Correct decision for opponent is pass
				// Wrong decision is take
				equRight = equDoublePass
				equWrong = equDoubleTake
			}

			// Calculate the threshold probability
			denominator := equWrong - equRight
			if denominator != 0 {
				p := (equNoDouble - equRight) / denominator
				wrongPassTakePercent = p * 100.0 // Convert to percentage
			} else {
				wrongPassTakePercent = 0.0 // Edge case: equities are equal
			}
		}

		// Cubeless double: compute as double the cubeless no-double equity
		// The XG file format doesn't store this separately (despite Python struct suggesting 4 floats)
		// Empirically, doubling the equity gives the correct value (~-0.008 for first cube)
		cubelessDouble := c.Doubled.Eval[6] * 2.0

		analysis := &CubeAnalysis{
			Player1WinRate:       p1Win,
			Player1GammonRate:    p1Gammon,
			Player1BgRate:        p1Bg,
			Player2GammonRate:    p2Gammon,
			Player2BgRate:        p2Bg,
			CubelessNoDouble:     c.Doubled.Eval[6],
			CubelessDouble:       cubelessDouble,
			CubefulNoDouble:      c.Doubled.EquB,
			CubefulDoubleTake:    c.Doubled.EquDouble,
			CubefulDoublePass:    c.Doubled.EquDrop,
			WrongPassTakePercent: wrongPassTakePercent,
			AnalysisDepth:        c.Doubled.Level,
		}
		move.Analysis = analysis
		move.Position.Score = c.Doubled.Score
		move.Position.Cube = c.Doubled.Cube
		move.Position.CubePos = c.Doubled.CubePos
	}

	return move
}

// convertMoveEntry converts a full MoveEntry to CheckerMove
func convertMoveEntry(m *MoveEntry) *CheckerMove {
	// Build initial position
	position := Position{
		Checkers: m.PositionI,
		Cube:     m.CubeA,
		CubePos:  0,              // Default
		Score:    [2]int32{0, 0}, // Would need to track
	}

	// Swap position to player on roll's perspective only when active_player == -1
	if m.ActiveP == -1 {
		position = swapPosition(position)
	}

	// Convert moves from XG internal format to desired output format
	// XG internal: -1=unused, 0-23=points (0-based), 24=bar, -2=bear off
	// We output: -1=unused, 1-24=points (1-based), 25=bar, -2=bear off
	var playedMove [8]int32
	for i := 0; i < 8; i++ {
		if m.Moves[i] == -1 {
			playedMove[i] = -1 // unused
		} else if m.Moves[i] == -2 {
			playedMove[i] = -2 // bear off
		} else if m.Moves[i] == 24 {
			playedMove[i] = 25 // bar (24 -> 25)
		} else {
			playedMove[i] = m.Moves[i] + 1 // points: add 1 to convert from 0-based to 1-based (0->1, 1->2, ..., 23->24)
		}
	}

	move := &CheckerMove{
		Position:     position,
		ActivePlayer: m.ActiveP,
		Dice:         m.Dice,
		PlayedMove:   playedMove,
		Analysis:     make([]CheckerAnalysis, 0),
	}

	// Extract analysis from DataMoves if available
	if m.DataMoves != nil {
		move.Position.Score = m.DataMoves.Score
		move.Position.Cube = m.DataMoves.Cube
		move.Position.CubePos = m.DataMoves.Cubepos

		// Add analysis for each evaluated move
		numMoves := int(m.DataMoves.NMoves)
		if numMoves > 32 {
			numMoves = 32
		}

		for i := 0; i < numMoves; i++ {
			// Convert move from XG internal format to desired output format
			// XG internal: -1=end of move/unused, 0-23=points (0-based), 24=bar, -2=bear off
			// We output: -1=unused, 1-24=points (1-based), 25=bar, -2=bear off
			var moveArray [8]int8
			endOfMove := false
			for j := 0; j < 8; j++ {
				if m.DataMoves.Moves[i][j] == -1 || endOfMove {
					moveArray[j] = -1 // unused (and everything after first -1)
					endOfMove = true
				} else if m.DataMoves.Moves[i][j] == -2 {
					moveArray[j] = -2 // bear off
				} else if m.DataMoves.Moves[i][j] == 24 {
					moveArray[j] = 25 // bar (24 -> 25)
				} else {
					moveArray[j] = m.DataMoves.Moves[i][j] + 1 // points: add 1 to convert from 0-based to 1-based (0->1, 1->2, ..., 23->24)
				}
			}

			analysisPosition := m.DataMoves.PosPlayed[i]

			analysis := CheckerAnalysis{
				Position: Position{
					Checkers: analysisPosition,
					Cube:     m.DataMoves.Cube,
					CubePos:  m.DataMoves.Cubepos,
					Score:    m.DataMoves.Score,
				},
				Move:              moveArray,
				Player1WinRate:    1.0 - m.DataMoves.Eval[i][2], // 1 - opponent win rate
				Player1GammonRate: m.DataMoves.Eval[i][4],
				Player1BgRate:     m.DataMoves.Eval[i][5],
				Player2GammonRate: m.DataMoves.Eval[i][1],
				Player2BgRate:     m.DataMoves.Eval[i][0],
				Equity:            m.DataMoves.Eval[i][6],
				AnalysisDepth:     m.DataMoves.EvalLevel[i].Level,
			}
			move.Analysis = append(move.Analysis, analysis)
		}
	}

	return move
}
