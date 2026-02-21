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
var testPositionEN = "XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10\n" +
"\n" +
"X:Player 1   O:Player 2\n" +
"Score is X:3 O:6 7 pt.(s) match.\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" |    X           X |   | X  O  O  O  O  O | +---+\n" +
" |                  |   | X  O  O  O  O  O | | 2 |\n" +
" |                  |   |    O           O | +---+\n" +
" |                  |   |                O |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |          X       |\n" +
" |                  |   | X  X  X  X     X |\n" +
" |       O          |   | X  X  X  X     X |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Pip count  X: 111  O: 52 X-O: 3-6/7\n" +
"Cube: 2, O own cube\n" +
"X to play 21\n" +
"\n" +
"    1. 4-ply       19/18 14/12                  eq:-0.491\n" +
"      Player:   25.45% (G:0.00% B:0.00%)\n" +
"      Opponent: 74.55% (G:31.09% B:0.09%)\n" +
"\n" +
"    2. 4-ply       19/18 3/1                    eq:-0.556 (-0.065)\n" +
"      Player:   22.19% (G:0.00% B:0.00%)\n" +
"      Opponent: 77.81% (G:35.24% B:0.12%)\n" +
"\n" +
"eXtreme Gammon Version: 2.10, MET: Kazaross XG2\n"

var testPositionFR = "XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10\n" +
"\n" +
"X:Player 1   O:Player 2\n" +
"Le score est X:3 O:6 match en 7 pt(s)\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" |    X           X |   | X  O  O  O  O  O | +---+\n" +
" |                  |   | X  O  O  O  O  O | | 2 |\n" +
" |                  |   |    O           O | +---+\n" +
" |                  |   |                O |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |          X       |\n" +
" |                  |   | X  X  X  X     X |\n" +
" |       O          |   | X  X  X  X     X |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Course  X: 111  O: 52 X-O: 3-6/7\n" +
"Videau: 2, O a le videau\n" +
"X \u00e0 jouer 21\n" +
"\n" +
"    1. 4-plis      19/18 14/12                  \u00e9q:-0.491\n" +
"      Joueur:     25.45% (G:0.00% B:0.00%)\n" +
"      Adversaire: 74.55% (G:31.09% B:0.09%)\n" +
"\n" +
"eXtreme Gammon Version: 2.10, TEM: Kazaross XG2\n"

var testCubePosition = "XGID=---BBaB-BbA-bC-b--BdAca---:0:0:1:00:0:5:0:9:10\n" +
"\n" +
"X:Player 1   O:Player 2\n" +
"Score is X:0 O:5 9 pt.(s) match.\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" | X     O        X |   | O  X  O  O       |\n" +
" | X     O        X |   | O     O          |\n" +
" | X                |   | O     O          |\n" +
" |                  |   | O                |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" | O        O  X    |   | X     X  X       |\n" +
" | O     X  O  X    |   | X  O  X  X       |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Pip count  X: 147  O: 137 X-O: 0-5/9\n" +
"Cube: 1\n" +
"X on roll, cube action\n" +
"\n" +
"Analyzed in XG Roller++\n" +
"Player Winning Chances:   54.40% (G:18.22% B:0.53%)\n" +
"Opponent Winning Chances: 45.60% (G:12.71% B:0.51%)\n" +
"\n" +
"Cubeless Equities: No Double=+0.103, Double=+0.259\n" +
"\n" +
"Cubeful Equities:\n" +
"       No double:     +0.337\n" +
"       Double/Take:   +0.215 (-0.122)\n" +
"\n" +
"eXtreme Gammon Version: 2.10, MET: Kazaross XG2\n"

func TestParseXGTextPosition_English(t *testing.T) {
reader := strings.NewReader(testPositionEN)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse English position: %v", err)
}

if pos.XGID != "XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10" {
t.Errorf("Wrong XGID: %s", pos.XGID)
}
if pos.Player1Name != "Player 1" {
t.Errorf("Wrong player1: %s", pos.Player1Name)
}
if pos.Player2Name != "Player 2" {
t.Errorf("Wrong player2: %s", pos.Player2Name)
}

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
if len(moves) < 2 {
t.Errorf("Expected at least 2 moves in JSON, got %d", len(moves))
}
} else {
t.Error("Moves not in expected format")
}
}

func TestParseXGTextPosition_MovePlayerStats(t *testing.T) {
reader := strings.NewReader(testPositionEN)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if len(pos.Analysis) < 2 {
t.Fatalf("Expected at least 2 moves, got %d", len(pos.Analysis))
}

m1 := pos.Analysis[0]
if m1.PlayerWin != 25.45 {
t.Errorf("Move 1 PlayerWin: got %.2f, want 25.45", m1.PlayerWin)
}
if m1.PlayerG != 0.00 {
t.Errorf("Move 1 PlayerG: got %.2f, want 0.00", m1.PlayerG)
}
if m1.OppWin != 74.55 {
t.Errorf("Move 1 OppWin: got %.2f, want 74.55", m1.OppWin)
}
if m1.OppG != 31.09 {
t.Errorf("Move 1 OppG: got %.2f, want 31.09", m1.OppG)
}

m2 := pos.Analysis[1]
if m2.PlayerWin != 22.19 {
t.Errorf("Move 2 PlayerWin: got %.2f, want 22.19", m2.PlayerWin)
}
if m2.OppWin != 77.81 {
t.Errorf("Move 2 OppWin: got %.2f, want 77.81", m2.OppWin)
}
}

func TestParseXGTextPosition_CubeAnalysisDepth(t *testing.T) {
reader := strings.NewReader(testCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if pos.CubeAnalysis == nil {
t.Fatal("Expected cube analysis")
}
if pos.CubeAnalysis.AnalysisDepth != "XG Roller++" {
t.Errorf("AnalysisDepth: got %q, want %q", pos.CubeAnalysis.AnalysisDepth, "XG Roller++")
}
}

var testFullCubePosition = "XGID=---BBaB-BbA-bC-b--BdAca---:0:0:1:00:0:5:0:9:10\n" +
"\n" +
"X:Player 1   O:Player 2\n" +
"Score is X:0 O:5 9 pt.(s) match.\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" | X     O        X |   | O  X  O  O       |\n" +
" | X     O        X |   | O     O          |\n" +
" | X                |   | O     O          |\n" +
" |                  |   | O                |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" | O        O  X    |   | X     X  X       |\n" +
" | O     X  O  X    |   | X  O  X  X       |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Pip count  X: 147  O: 137 X-O: 0-5/9\n" +
"Cube: 1\n" +
"X on roll, cube action\n" +
"\n" +
"Analyzed in XG Roller++\n" +
"Player Winning Chances:   54.40% (G:18.22% B:0.53%)\n" +
"Opponent Winning Chances: 45.60% (G:12.71% B:0.51%)\n" +
"\n" +
"Cubeless Equities: No Double=+0.103, Double=+0.259\n" +
"\n" +
"Cubeful Equities:\n" +
"       No double:     +0.337\n" +
"       Double/Take:   +0.215 (-0.122)\n" +
"       Double/Pass:   +1.000 (+0.663)\n" +
"\n" +
"Best Cube action: No double\n" +
"Percentage of wrong pass needed to make the double decision right: 12.34%\n" +
"Percentage of wrong take needed to make the double decision right: 56.78%\n" +
"\n" +
"This is a user comment\n" +
"spanning multiple lines.\n" +
"\n" +
"eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2\n"

func TestParseXGTextPosition_CubeErrors(t *testing.T) {
reader := strings.NewReader(testFullCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

ca := pos.CubeAnalysis
if ca == nil {
t.Fatal("Expected cube analysis")
}
if ca.DoubleTakeError != -0.122 {
t.Errorf("DoubleTakeError: got %.3f, want -0.122", ca.DoubleTakeError)
}
if ca.DoubleDropError != 0.663 {
t.Errorf("DoubleDropError: got %.3f, want 0.663", ca.DoubleDropError)
}
if ca.NoDoubleError != 0 {
t.Errorf("NoDoubleError: got %.3f, want 0", ca.NoDoubleError)
}
}

func TestParseXGTextPosition_WrongPercentages(t *testing.T) {
reader := strings.NewReader(testFullCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

ca := pos.CubeAnalysis
if ca.WrongPassPct != 12.34 {
t.Errorf("WrongPassPct: got %.2f, want 12.34", ca.WrongPassPct)
}
if ca.WrongTakePct != 56.78 {
t.Errorf("WrongTakePct: got %.2f, want 56.78", ca.WrongTakePct)
}
}

func TestParseXGTextPosition_CubeRecommendation(t *testing.T) {
reader := strings.NewReader(testFullCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if pos.CubeAnalysis.Recommendation != "No double" {
t.Errorf("Recommendation: got %q, want %q", pos.CubeAnalysis.Recommendation, "No double")
}
}

func TestParseXGTextPosition_Comment(t *testing.T) {
reader := strings.NewReader(testFullCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

expected := "This is a user comment\nspanning multiple lines."
if pos.Comment != expected {
t.Errorf("Comment: got %q, want %q", pos.Comment, expected)
}
}

func TestParseXGTextPosition_NoComment(t *testing.T) {
reader := strings.NewReader(testCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if pos.Comment != "" {
t.Errorf("Expected empty comment, got %q", pos.Comment)
}
}

func TestParseXGTextPosition_Version(t *testing.T) {
reader := strings.NewReader(testFullCubePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if pos.Version != "2.19.211.pre-release" {
t.Errorf("Version: got %q, want %q", pos.Version, "2.19.211.pre-release")
}
if pos.MET != "Kazaross XG2" {
t.Errorf("MET: got %q, want %q", pos.MET, "Kazaross XG2")
}
}

var testPastePosition = "XGID=---CBbB-B---bC-b-bcbbC----:0:0:1:42:0:0:0:13:10\n" +
"\n" +
"X:postmanpat   O:nando\n" +
"Score is X:0 O:0 13 pt.(s) match.\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" | X     O     O  O |   | O  O  X          |\n" +
" | X     O     O  O |   | O  O  X          |\n" +
" | X              O |   |       X          |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |          X       |\n" +
" | O           X    |   | X  O  X  X       |\n" +
" | O           X    |   | X  O  X  X       |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Pip count  X: 147  O: 145 X-O: 0-0/13\n" +
"Cube: 1\n" +
"X to play 42\n" +
"\n" +
"    1. 3-ply       8/6 8/4                      eq:-0.646\n" +
"      Player:   31.61% (G:4.25% B:0.07%)\n" +
"      Opponent: 68.39% (G:9.31% B:0.26%)\n" +
"\n" +
"    2. 3-ply       13/7                         eq:-0.751 (-0.105)\n" +
"      Player:   30.01% (G:4.48% B:0.10%)\n" +
"      Opponent: 69.99% (G:11.82% B:0.36%)\n" +
"\n" +
"    3. 2-ply       13/9 3/1                     eq:-0.852 (-0.206)\n" +
"      Player:   28.38% (G:4.06% B:0.08%)\n" +
"      Opponent: 71.62% (G:12.49% B:0.40%)\n" +
"\n" +
"    4. 2-ply       6/4 6/2                      eq:-0.891 (-0.245)\n" +
"      Player:   27.26% (G:3.43% B:0.06%)\n" +
"      Opponent: 72.74% (G:11.09% B:0.44%)\n" +
"\n" +
"    5. 2-ply       8/2                          eq:-0.919 (-0.273)\n" +
"      Player:   27.15% (G:4.38% B:0.08%)\n" +
"      Opponent: 72.85% (G:13.17% B:0.42%)\n" +
"\n" +
"\n" +
"eXtreme Gammon Version: 2.19.211.pre-release, MET: Kazaross XG2\n"

func TestParseXGTextPosition_PastePosition(t *testing.T) {
reader := strings.NewReader(testPastePosition)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse pasted position: %v", err)
}

if pos.Player1Name != "postmanpat" {
t.Errorf("Player1Name: got %q, want %q", pos.Player1Name, "postmanpat")
}
if pos.Player2Name != "nando" {
t.Errorf("Player2Name: got %q, want %q", pos.Player2Name, "nando")
}
if pos.ActionType != "play" {
t.Errorf("ActionType: got %q, want %q", pos.ActionType, "play")
}

if len(pos.Analysis) != 5 {
t.Fatalf("Expected 5 moves, got %d", len(pos.Analysis))
}

m1 := pos.Analysis[0]
if m1.Rank != 1 {
t.Errorf("Move 1 Rank: got %d, want 1", m1.Rank)
}
if m1.Ply != 3 {
t.Errorf("Move 1 Ply: got %d, want 3", m1.Ply)
}
if m1.Move != "8/6 8/4" {
t.Errorf("Move 1 Move: got %q, want %q", m1.Move, "8/6 8/4")
}
if m1.Equity != -0.646 {
t.Errorf("Move 1 Equity: got %.3f, want -0.646", m1.Equity)
}
if m1.PlayerWin != 31.61 {
t.Errorf("Move 1 PlayerWin: got %.2f, want 31.61", m1.PlayerWin)
}
if m1.PlayerG != 4.25 {
t.Errorf("Move 1 PlayerG: got %.2f, want 4.25", m1.PlayerG)
}
if m1.PlayerB != 0.07 {
t.Errorf("Move 1 PlayerB: got %.2f, want 0.07", m1.PlayerB)
}
if m1.OppWin != 68.39 {
t.Errorf("Move 1 OppWin: got %.2f, want 68.39", m1.OppWin)
}
if m1.OppG != 9.31 {
t.Errorf("Move 1 OppG: got %.2f, want 9.31", m1.OppG)
}
if m1.OppB != 0.26 {
t.Errorf("Move 1 OppB: got %.2f, want 0.26", m1.OppB)
}

m2 := pos.Analysis[1]
if m2.EquityDiff != -0.105 {
t.Errorf("Move 2 EquityDiff: got %.3f, want -0.105", m2.EquityDiff)
}

m5 := pos.Analysis[4]
if m5.Move != "8/2" {
t.Errorf("Move 5 Move: got %q, want %q", m5.Move, "8/2")
}
if m5.PlayerWin != 27.15 {
t.Errorf("Move 5 PlayerWin: got %.2f, want 27.15", m5.PlayerWin)
}

if pos.Version != "2.19.211.pre-release" {
t.Errorf("Version: got %q, want %q", pos.Version, "2.19.211.pre-release")
}
if pos.MET != "Kazaross XG2" {
t.Errorf("MET: got %q, want %q", pos.MET, "Kazaross XG2")
}
}

var testCheckerWithComment = "XGID=-B-CBBB---a---A---ABcbbbd-:1:-1:1:21:3:6:0:7:10\n" +
"\n" +
"X:Player 1   O:Player 2\n" +
"Score is X:3 O:6 7 pt.(s) match.\n" +
" +13-14-15-16-17-18------19-20-21-22-23-24-+\n" +
" |    X           X |   | X  O  O  O  O  O |\n" +
" |                  |   | X  O  O  O  O  O |\n" +
" |                  |   |    O           O |\n" +
" |                  |   |                O |\n" +
" |                  |   |                  |\n" +
" |                  |BAR|                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |                  |\n" +
" |                  |   |          X       |\n" +
" |                  |   | X  X  X  X     X |\n" +
" |       O          |   | X  X  X  X     X |\n" +
" +12-11-10--9--8--7-------6--5--4--3--2--1-+\n" +
"Pip count  X: 111  O: 52 X-O: 3-6/7\n" +
"Cube: 2, O own cube\n" +
"X to play 21\n" +
"\n" +
"    1. 4-ply       19/18 14/12                  eq:-0.491\n" +
"      Player:   25.45% (G:0.00% B:0.00%)\n" +
"      Opponent: 74.55% (G:31.09% B:0.09%)\n" +
"\n" +
"\n" +
"My checker comment here.\n" +
"\n" +
"eXtreme Gammon Version: 2.10, MET: Kazaross XG2\n"

func TestParseXGTextPosition_CheckerComment(t *testing.T) {
reader := strings.NewReader(testCheckerWithComment)
pos, err := ParseXGTextPosition(reader)
if err != nil {
t.Fatalf("Failed to parse: %v", err)
}

if pos.Comment != "My checker comment here." {
t.Errorf("Comment: got %q, want %q", pos.Comment, "My checker comment here.")
}
}
