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
	"encoding/json"
	"io"
)

// MatchMetadata contains essential match information
type MatchMetadata struct {
	Player1Name string `json:"player1_name"`
	Player2Name string `json:"player2_name"`
	Location    string `json:"location"`
	Event       string `json:"event"`
	Round       string `json:"round"`
	DateTime    string `json:"date_time"`
	MatchLength int32  `json:"match_length"`
}

// Position represents a backgammon position
type Position struct {
	Checkers [26]int8 `json:"checkers"` // Position of checkers
	Cube     int32    `json:"cube"`     // Cube value
	CubePos  int32    `json:"cube_pos"` // Cube position (0=center, 1=player1, -1=player2)
	Score    [2]int32 `json:"score"`    // Match score [player1, player2]
}

// CheckerAnalysis contains analysis for a single checker move
type CheckerAnalysis struct {
	Position          Position `json:"position"`            // Resulting position
	Move              [8]int8  `json:"move"`                // The move itself
	Player1WinRate    float32  `json:"player1_win_rate"`    // eval[0]
	Player1GammonRate float32  `json:"player1_gammon_rate"` // eval[1]
	Player1BgRate     float32  `json:"player1_bg_rate"`     // eval[2]
	Player2GammonRate float32  `json:"player2_gammon_rate"` // eval[3]
	Player2BgRate     float32  `json:"player2_bg_rate"`     // eval[4]
	Equity            float32  `json:"equity"`              // eval[5] - normalized equity
	AnalysisDepth     int16    `json:"analysis_depth"`      // EvalLevel.Level
}

// CubeAnalysis contains analysis for a cube decision
type CubeAnalysis struct {
	Player1WinRate       float32 `json:"player1_win_rate"`        // eval[0]
	Player1GammonRate    float32 `json:"player1_gammon_rate"`     // eval[1]
	Player1BgRate        float32 `json:"player1_bg_rate"`         // eval[2]
	Player2GammonRate    float32 `json:"player2_gammon_rate"`     // eval[3]
	Player2BgRate        float32 `json:"player2_bg_rate"`         // eval[4]
	CubelessNoDouble     float32 `json:"cubeless_no_double"`      // eval[5]
	CubelessDouble       float32 `json:"cubeless_double"`         // eval[6]
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
	PlayedMove   [8]int32          `json:"played_move"`   // The move that was played
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
						Player1Name: getPreferredString(r.Player1, r.SPlayer1),
						Player2Name: getPreferredString(r.Player2, r.SPlayer2),
						Location:    getPreferredString(r.Location, r.SLocation),
						Event:       getPreferredString(r.Event, r.SEvent),
						Round:       getPreferredString(r.Round, r.SRound),
						DateTime:    r.Date,
						MatchLength: r.MatchLength,
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
						cubeMove := convertCubeEntry(r)
						currentGame.Moves = append(currentGame.Moves, Move{
							MoveType: "cube",
							CubeMove: cubeMove,
						})
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

// convertCubeEntry converts a full CubeEntry to CubeMove
func convertCubeEntry(c *CubeEntry) *CubeMove {
	move := &CubeMove{
		Position: Position{
			Checkers: c.Position,
			Cube:     c.CubeB,
			CubePos:  0,              // Default, could be extracted from more context
			Score:    [2]int32{0, 0}, // Would need to track from game state
		},
		ActivePlayer: c.ActiveP,
		CubeAction:   c.Double, // Simplified - may need more logic
	}

	// Add cube analysis if available
	if c.Doubled != nil {
		analysis := &CubeAnalysis{
			Player1WinRate:       c.Doubled.Eval[0],
			Player1GammonRate:    c.Doubled.Eval[1],
			Player1BgRate:        c.Doubled.Eval[2],
			Player2GammonRate:    c.Doubled.Eval[3],
			Player2BgRate:        c.Doubled.Eval[4],
			CubelessNoDouble:     c.Doubled.Eval[5],
			CubelessDouble:       c.Doubled.Eval[6],
			CubefulNoDouble:      c.Doubled.EquB,
			CubefulDoubleTake:    c.Doubled.EquDouble,
			CubefulDoublePass:    c.Doubled.EquDrop,
			WrongPassTakePercent: 0, // Could be calculated from equity differences
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
	move := &CheckerMove{
		Position: Position{
			Checkers: m.PositionI,
			Cube:     m.CubeA,
			CubePos:  0,              // Default
			Score:    [2]int32{0, 0}, // Would need to track
		},
		ActivePlayer: m.ActiveP,
		Dice:         m.Dice,
		PlayedMove:   m.Moves,
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
			// Convert move from [8]int8 to [8]int8
			var moveArray [8]int8
			for j := 0; j < 8; j++ {
				moveArray[j] = m.DataMoves.Moves[i][j]
			}

			analysis := CheckerAnalysis{
				Position: Position{
					Checkers: m.DataMoves.PosPlayed[i],
					Cube:     m.DataMoves.Cube,
					CubePos:  m.DataMoves.Cubepos,
					Score:    m.DataMoves.Score,
				},
				Move:              moveArray,
				Player1WinRate:    m.DataMoves.Eval[i][0],
				Player1GammonRate: m.DataMoves.Eval[i][1],
				Player1BgRate:     m.DataMoves.Eval[i][2],
				Player2GammonRate: m.DataMoves.Eval[i][3],
				Player2BgRate:     m.DataMoves.Eval[i][4],
				Equity:            m.DataMoves.Eval[i][5],
				AnalysisDepth:     m.DataMoves.EvalLevel[i].Level,
			}
			move.Analysis = append(move.Analysis, analysis)
		}
	}

	return move
}
