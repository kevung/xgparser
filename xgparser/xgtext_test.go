//
//   xgtext_test.go - Tests for XG text position parser
//   Copyright (C) 2025 Kevin Unger
//

package xgparser

import (
	"strings"
	"testing"
)

// Test data from actual XG text export
const testPositionEN = `XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10

X:Player 1   O:Player 2
Score is X:3 O:6 7 pt.(s) match.
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 |    X           X |   | X  O  O  O  O  O | +---+
 |                  |   | X  O  O  O  O  O | | 2 |
 |                  |   |    O           O | +---+
 |                  |   |                O |
 |                  |   |                  |
 |                  |BAR|                  |
 |                  |   |                  |
 |                  |   |                  |
 |                  |   |          X       |
 |                  |   | X  X  X  X     X |
 |       O          |   | X  X  X  X     X |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Pip count  X: 111  O: 52 X-O: 3-6/7
Cube: 2, O own cube
X to play 21

    1. 4-ply       19/18 14/12                  eq:-0.491
      Player:   25.45% (G:0.00% B:0.00%)
      Opponent: 74.55% (G:31.09% B:0.09%)

    2. 4-ply       19/18 3/1                    eq:-0.556 (-0.065)
      Player:   22.19% (G:0.00% B:0.00%)
      Opponent: 77.81% (G:35.24% B:0.12%)

eXtreme Gammon Version: 2.10, MET: Kazaross XG2
`

const testPositionFR = `XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10

X:Player 1   O:Player 2
Le score est X:3 O:6 match en 7 pt(s)
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 |    X           X |   | X  O  O  O  O  O | +---+
 |                  |   | X  O  O  O  O  O | | 2 |
 |                  |   |    O           O | +---+
 |                  |   |                O |
 |                  |   |                  |
 |                  |BAR|                  |
 |                  |   |                  |
 |                  |   |                  |
 |                  |   |          X       |
 |                  |   | X  X  X  X     X |
 |       O          |   | X  X  X  X     X |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Course  X: 111  O: 52 X-O: 3-6/7
Videau: 2, O a le videau
X à jouer 21

    1. 4-plis      19/18 14/12                  éq:-0.491
      Joueur:     25.45% (G:0.00% B:0.00%)
      Adversaire: 74.55% (G:31.09% B:0.09%)

eXtreme Gammon Version: 2.10, TEM: Kazaross XG2
`

const testCubePosition = `XGID=---BBaB-BbA-bC-b--BdAca---:0:0:1:00:0:5:0:9:10

X:Player 1   O:Player 2
Score is X:0 O:5 9 pt.(s) match.
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 | X     O        X |   | O  X  O  O       |
 | X     O        X |   | O     O          |
 | X                |   | O     O          |
 |                  |   | O                |
 |                  |   |                  |
 |                  |BAR|                  |
 |                  |   |                  |
 |                  |   |                  |
 |                  |   |                  |
 | O        O  X    |   | X     X  X       |
 | O     X  O  X    |   | X  O  X  X       |
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Pip count  X: 147  O: 137 X-O: 0-5/9
Cube: 1
X on roll, cube action

Analyzed in XG Roller++
Player Winning Chances:   54.40% (G:18.22% B:0.53%)
Opponent Winning Chances: 45.60% (G:12.71% B:0.51%)

Cubeless Equities: No Double=+0.103, Double=+0.259

Cubeful Equities:
       No double:     +0.337
       Double/Take:   +0.215 (-0.122)

eXtreme Gammon Version: 2.10, MET: Kazaross XG2
`

func TestParseXGTextPosition_English(t *testing.T) {
	reader := strings.NewReader(testPositionEN)
	pos, err := ParseXGTextPosition(reader)
	if err != nil {
		t.Fatalf("Failed to parse English position: %v", err)
	}

	// Verify basic fields
	if pos.XGID != "XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10" {
		t.Errorf("Wrong XGID: %s", pos.XGID)
	}
	if pos.Player1Name != "Player 1" {
		t.Errorf("Wrong player1: %s", pos.Player1Name)
	}
	if pos.Player2Name != "Player 2" {
		t.Errorf("Wrong player2: %s", pos.Player2Name)
	}

	// Parse XGID for position details
	xgidComp, err := ParseXGID(pos.XGID)
	if err != nil {
		t.Fatalf("Failed to parse XGID: %v", err)
	}
	if xgidComp.ScoreX != 3 || xgidComp.ScoreO != 6 || xgidComp.MatchLength != 7 {
		t.Errorf("Wrong score from XGID: X:%d O:%d Match:%d", xgidComp.ScoreX, xgidComp.ScoreO, xgidComp.MatchLength)
	}
	if xgidComp.Dice != "21" {
		t.Errorf("Wrong dice from XGID: %s", xgidComp.Dice)
	}
	if pos.ActionType != "play" {
		t.Errorf("Wrong action type: %s", pos.ActionType)
	}

	// Verify moves (test data only has 2 moves in const, but real file has 5)
	if len(pos.Analysis) < 2 {
		t.Errorf("Expected at least 2 moves, got %d", len(pos.Analysis))
	}
	if len(pos.Analysis) > 0 {
		move1 := pos.Analysis[0]
		if move1.Rank != 1 || move1.Ply != 4 {
			t.Errorf("Wrong move 1: rank=%d ply=%d", move1.Rank, move1.Ply)
		}
		if move1.Equity != -0.491 {
			t.Errorf("Wrong equity: %.3f", move1.Equity)
		}
	}

	// Verify version
	if pos.Version != "2.10" || pos.MET != "Kazaross XG2" {
		t.Errorf("Wrong version: %s MET: %s", pos.Version, pos.MET)
	}
}

func TestParseXGTextPosition_French(t *testing.T) {
	reader := strings.NewReader(testPositionFR)
	pos, err := ParseXGTextPosition(reader)
	if err != nil {
		t.Fatalf("Failed to parse French position: %v", err)
	}

	// Verify it parsed the same position as English
	xgidComp, err := ParseXGID(pos.XGID)
	if err != nil {
		t.Fatalf("Failed to parse XGID: %v", err)
	}
	if xgidComp.ScoreX != 3 || xgidComp.ScoreO != 6 || xgidComp.MatchLength != 7 {
		t.Errorf("Wrong score from French: X:%d O:%d Match:%d", xgidComp.ScoreX, xgidComp.ScoreO, xgidComp.MatchLength)
	}
	if xgidComp.Dice != "21" {
		t.Errorf("Wrong dice: %s", xgidComp.Dice)
	}
	if pos.ActionType != "play" {
		t.Errorf("Expected play action, got: %s", pos.ActionType)
	}

	// Verify French move analysis was parsed
	if len(pos.Analysis) < 1 {
		t.Errorf("Expected at least 1 move, got %d", len(pos.Analysis))
	}
}

func TestParseXGTextPosition_CubeAction(t *testing.T) {
	reader := strings.NewReader(testCubePosition)
	pos, err := ParseXGTextPosition(reader)
	if err != nil {
		t.Fatalf("Failed to parse cube position: %v", err)
	}

	if pos.ActionType != "cube" {
		t.Errorf("Expected cube action, got: %s", pos.ActionType)
	}

	if pos.CubeAnalysis == nil {
		t.Fatal("Expected cube analysis")
	}

	ca := pos.CubeAnalysis
	if ca.PlayerWin != 54.40 {
		t.Errorf("Wrong player win: %.2f", ca.PlayerWin)
	}
	if ca.OppWin != 45.60 {
		t.Errorf("Wrong opp win: %.2f", ca.OppWin)
	}
	if ca.CubelessNoDouble != 0.103 {
		t.Errorf("Wrong cubeless no double: %.3f", ca.CubelessNoDouble)
	}
	if ca.NoDouble != 0.337 {
		t.Errorf("Wrong no double: %.3f", ca.NoDouble)
	}
}

func TestToJSON(t *testing.T) {
	reader := strings.NewReader(testPositionEN)
	pos, err := ParseXGTextPosition(reader)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	jsonMap := pos.ToJSON()
	if jsonMap["xgid"] != pos.XGID {
		t.Errorf("JSON XGID mismatch")
	}
	if jsonMap["player1"] != pos.Player1Name {
		t.Errorf("JSON player1 mismatch")
	}
	if moves, ok := jsonMap["moves"].([]map[string]interface{}); ok {
		// Test const only has 2 moves, real file has 5
		if len(moves) < 2 {
			t.Errorf("Expected at least 2 moves in JSON, got %d", len(moves))
		}
	} else {
		t.Error("Moves not in expected format")
	}
}
