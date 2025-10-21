//
//   xgid_test.go - Unit tests for XGID parsing
//   Copyright (C) 2025 Kevin Unger
//

package xgparser

import (
	"strings"
	"testing"
)

func TestParseXGID(t *testing.T) {
	tests := []struct {
		name     string
		xgid     string
		wantErr  bool
		expected XGIDComponents
	}{
		{
			name: "full XGID with all components",
			xgid: "----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10",
			expected: XGIDComponents{
				PositionID:   "----BaC-B---aD--aa-bcbbBbB",
				CubeOwner:    0,
				CubeValue:    0,
				PlayerToMove: 1,
				Dice:         "22",
				ScoreX:       2,
				ScoreO:       3,
				CrawfordFlag: 0,
				MatchLength:  13,
				MaxCube:      10,
			},
		},
		{
			name: "XGID with prefix",
			xgid: "XGID=-b----E-C---eE---c-e----B-:0:0:-1:51:0:0:0:13:10",
			expected: XGIDComponents{
				PositionID:   "-b----E-C---eE---c-e----B-",
				CubeOwner:    0,
				CubeValue:    0,
				PlayerToMove: -1,
				Dice:         "51",
				ScoreX:       0,
				ScoreO:       0,
				CrawfordFlag: 0,
				MatchLength:  13,
				MaxCube:      10,
			},
		},
		{
			name: "cube owned by X",
			xgid: "----BBB-B---aC----B-B-gdc-:1:1:1:00:1:4:0:7:10",
			expected: XGIDComponents{
				PositionID:   "----BBB-B---aC----B-B-gdc-",
				CubeOwner:    1,
				CubeValue:    1,
				PlayerToMove: 1,
				Dice:         "00",
				ScoreX:       1,
				ScoreO:       4,
				CrawfordFlag: 0,
				MatchLength:  7,
				MaxCube:      10,
			},
		},
		{
			name:    "invalid XGID - too few parts",
			xgid:    "position:0:0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseXGID(tt.xgid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseXGID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.PositionID != tt.expected.PositionID {
				t.Errorf("PositionID = %v, want %v", got.PositionID, tt.expected.PositionID)
			}
			if got.CubeOwner != tt.expected.CubeOwner {
				t.Errorf("CubeOwner = %v, want %v", got.CubeOwner, tt.expected.CubeOwner)
			}
			if got.CubeValue != tt.expected.CubeValue {
				t.Errorf("CubeValue = %v, want %v", got.CubeValue, tt.expected.CubeValue)
			}
			if got.PlayerToMove != tt.expected.PlayerToMove {
				t.Errorf("PlayerToMove = %v, want %v", got.PlayerToMove, tt.expected.PlayerToMove)
			}
			if got.Dice != tt.expected.Dice {
				t.Errorf("Dice = %v, want %v", got.Dice, tt.expected.Dice)
			}
		})
	}
}

func TestParseXGIDFromReader_English(t *testing.T) {
	input := `XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10

X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 | X        O  O    |   | O  O  O  O  X  O |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2
`

	move, metadata, err := ParseXGIDFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseXGIDFromReader() error = %v", err)
	}

	if metadata.Player1Name != "postmanpat" {
		t.Errorf("Player1Name = %v, want postmanpat", metadata.Player1Name)
	}

	if metadata.Player2Name != "marcow777" {
		t.Errorf("Player2Name = %v, want marcow777", metadata.Player2Name)
	}

	if move.Position.Score[0] != 2 || move.Position.Score[1] != 3 {
		t.Errorf("Score = %v, want [2, 3]", move.Position.Score)
	}

	if metadata.MatchLength != 13 {
		t.Errorf("MatchLength = %v, want 13", metadata.MatchLength)
	}

	if move.Position.Cube != 1 {
		t.Errorf("Cube = %v, want 1", move.Position.Cube)
	}

	if move.ActivePlayer != 1 {
		t.Errorf("ActivePlayer = %v, want 1 (X)", move.ActivePlayer)
	}

	if move.Dice[0] != 2 || move.Dice[1] != 2 {
		t.Errorf("Dice = %v, want [2, 2]", move.Dice)
	}

	if len(move.Analysis) != 1 {
		t.Errorf("Analysis count = %v, want 1", len(move.Analysis))
	}

	if len(move.Analysis) > 0 {
		analysis := move.Analysis[0]
		if analysis.AnalysisDepth != 3 {
			t.Errorf("Analysis depth = %v, want 3", analysis.AnalysisDepth)
		}
		if analysis.Equity != -1.0 {
			t.Errorf("Equity = %v, want -1.0", analysis.Equity)
		}
		if analysis.Player1WinRate < 0.34 || analysis.Player1WinRate > 0.36 {
			t.Errorf("Player1WinRate = %v, want ~0.35", analysis.Player1WinRate)
		}
	}

	if metadata.ProductVersion != "2.19.211.pre-release" {
		t.Errorf("ProductVersion = %v, want 2.19.211.pre-release", metadata.ProductVersion)
	}

	if metadata.MET != "Kazaross XG2" {
		t.Errorf("MET = %v, want Kazaross XG2", metadata.MET)
	}
}

func TestParseXGIDFromReader_French(t *testing.T) {
	input := `XGID=-b----E-C---eE---c-e----B-:0:0:-1:51:0:0:0:13:10

X:marcow777   O:postmanpat
Le score est X:0 O:0 match en 13 pt(s)
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 | X           O    |   | O              X |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Videau: 1
X à jouer 51

    1. Livre¹      24/23 13/8                   éq:+0.007
      Joueur:     49.93% (G:13.21% B:0.51%)
      Adversaire: 50.07% (G:12.68% B:0.48%)

eXtreme Gammon Version: 2.19.211.pre-release, TEM: Kazaross XG2
`

	move, metadata, err := ParseXGIDFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseXGIDFromReader() error = %v", err)
	}

	if metadata.Player1Name != "marcow777" {
		t.Errorf("Player1Name = %v, want marcow777", metadata.Player1Name)
	}

	if move.Position.Score[0] != 0 || move.Position.Score[1] != 0 {
		t.Errorf("Score = %v, want [0, 0]", move.Position.Score)
	}

	if move.Position.Cube != 1 {
		t.Errorf("Cube = %v, want 1", move.Position.Cube)
	}

	if move.Dice[0] != 5 || move.Dice[1] != 1 {
		t.Errorf("Dice = %v, want [5, 1]", move.Dice)
	}

	if len(move.Analysis) != 1 {
		t.Errorf("Analysis count = %v, want 1", len(move.Analysis))
	}

	if len(move.Analysis) > 0 {
		analysis := move.Analysis[0]
		if analysis.AnalysisDepth != 0 {
			t.Errorf("AnalysisDepth = %v, want 0 (book)", analysis.AnalysisDepth)
		}
		if analysis.Equity < 0.006 || analysis.Equity > 0.008 {
			t.Errorf("Equity = %v, want ~0.007", analysis.Equity)
		}
	}
}

func TestParseXGIDFromReader_German(t *testing.T) {
	input := `XGID=-b----E-C---eE---c-e----B-:0:0:-1:51:0:0:0:13:10

X:marcow777   O:postmanpat
Spielstand ist S:0 G:0 13 Punkte(e) Match.
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 | X           O    |   | O              X |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Doppler: 1
X zu spielen 51

    1. Buch¹       24/23 13/8                   eq:+0.007
      Spieler: 49.93% (G:13.21% B:0.51%)
      Gegner:  50.07% (G:12.68% B:0.48%)

    2. 2 Züge      24/23 8/3                    eq:-0.164 (-0.171)
      Spieler: 46.08% (G:12.07% B:0.47%)
      Gegner:  53.92% (G:15.70% B:0.74%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2
`

	move, _, err := ParseXGIDFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseXGIDFromReader() error = %v", err)
	}

	if len(move.Analysis) != 2 {
		t.Errorf("Analysis count = %v, want 2", len(move.Analysis))
	}

	if len(move.Analysis) > 1 {
		// Check first move (book)
		analysis1 := move.Analysis[0]
		if analysis1.AnalysisDepth != 0 {
			t.Errorf("Analysis 1 depth = %v, want 0 (book)", analysis1.AnalysisDepth)
		}

		// Check second move (2-ply in German)
		analysis2 := move.Analysis[1]
		if analysis2.AnalysisDepth != 2 {
			t.Errorf("Analysis 2 depth = %v, want 2", analysis2.AnalysisDepth)
		}
	}
}

func TestXGIDToPosition(t *testing.T) {
	// Test starting position encoding
	// This is a simplified test - actual XGID encoding is complex
	positionID := "-b----E-C---eE---c-e----B-"
	position := XGIDToPosition(positionID)

	// Position should be a valid [26]int8 array
	if len(position) != 26 {
		t.Errorf("Position length = %v, want 26", len(position))
	}

	// Basic sanity check - starting position shouldn't have extreme values
	for i, count := range position {
		if count > 15 || count < -15 {
			t.Errorf("Position[%d] = %v, invalid checker count", i, count)
		}
	}
}

func TestToCheckerMove(t *testing.T) {
	input := `XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10

X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)

eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2
`

	move, _, err := ParseXGIDFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseXGIDFromReader() error = %v", err)
	}

	if move == nil {
		t.Fatal("ParseXGIDFromReader() returned nil CheckerMove")
	}

	if move.ActivePlayer != 1 {
		t.Errorf("ActivePlayer = %v, want 1", move.ActivePlayer)
	}

	if move.Dice[0] != 2 || move.Dice[1] != 2 {
		t.Errorf("Dice = %v, want [2, 2]", move.Dice)
	}

	if move.Position.Cube != 1 {
		t.Errorf("Position.Cube = %v, want 1", move.Position.Cube)
	}

	if move.Position.CubePos != 0 {
		t.Errorf("Position.CubePos = %v, want 0", move.Position.CubePos)
	}

	if move.Position.Score[0] != 2 || move.Position.Score[1] != 3 {
		t.Errorf("Position.Score = %v, want [2, 3]", move.Position.Score)
	}

	if len(move.Analysis) != 1 {
		t.Errorf("Analysis length = %v, want 1", len(move.Analysis))
	}

	if len(move.Analysis) > 0 {
		analysis := move.Analysis[0]
		if analysis.Equity != -1.0 {
			t.Errorf("Analysis.Equity = %v, want -1.0", analysis.Equity)
		}
		if analysis.AnalysisDepth != 3 {
			t.Errorf("Analysis.AnalysisDepth = %v, want 3", analysis.AnalysisDepth)
		}
	}
}
