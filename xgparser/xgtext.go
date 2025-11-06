//
//   xgtext.go - XG text position parser module
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
//   This module parses XG text positions exported from eXtreme Gammon
//   in multiple languages (English, French, German, Japanese)
//

package xgparser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// XGTextPosition represents a parsed text position from XG
// Note: Position info (checkers, cube, dice, score) is in XGID
// This structure captures the additional analysis data not in XGID
type XGTextPosition struct {
	XGID         string          // Complete position ID
	Player1Name  string          // X player name
	Player2Name  string          // O player name
	ActionType   string          // "play", "cube", "redouble"
	Analysis     []XGMove        // Move analysis (for play actions)
	CubeAnalysis *XGCubeAnalysis // Cube analysis (for cube actions)
	Version      string          // XG version
	MET          string          // Match equity table
}

// XGMove represents a move analysis
type XGMove struct {
	Rank       int
	Ply        int
	Move       string
	Equity     float64
	EquityDiff float64
	PlayerWin  float64
	PlayerG    float64
	PlayerB    float64
	OppWin     float64
	OppG       float64
	OppB       float64
}

// XGCubeAnalysis represents cube action analysis
type XGCubeAnalysis struct {
	PlayerWin        float64
	PlayerG          float64
	PlayerB          float64
	OppWin           float64
	OppG             float64
	OppB             float64
	CubelessNoDouble float64
	CubelessDouble   float64
	NoDouble         float64
	DoubleTake       float64
	DoubleDrop       float64
	NoRedouble       float64
	RedoubleTake     float64
	RedoubleDrop     float64
	Recommendation   string
	AnalyzerType     string // "XG Roller++" etc
}

// ParseXGTextPosition parses an XG text position from a reader
func ParseXGTextPosition(r io.Reader) (*XGTextPosition, error) {
	scanner := bufio.NewScanner(r)
	pos := &XGTextPosition{
		Analysis: make([]XGMove, 0),
	}

	inBoard := false
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Parse XGID (first line)
		if strings.HasPrefix(line, "XGID=") {
			pos.XGID = strings.TrimSpace(line)
			continue
		}

		// Parse player names
		if strings.Contains(line, "X:") && strings.Contains(line, "O:") &&
			!strings.Contains(line, "Score") && !strings.Contains(line, "score") &&
			!strings.Contains(line, "Punktzahl") && !strings.Contains(line, "Pip") &&
			!strings.Contains(line, "Course") {
			parts := strings.Split(line, "O:")
			if len(parts) == 2 {
				pos.Player1Name = strings.TrimSpace(strings.TrimPrefix(parts[0], "X:"))
				pos.Player2Name = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Skip score, pip count, cube lines (all info is in XGID)
		if strings.Contains(line, "Score is") || strings.Contains(line, "score est") ||
			strings.Contains(line, "Punktzahl ist") ||
			strings.Contains(line, "Pip count") || strings.Contains(line, "Course") ||
			strings.Contains(line, "Cube:") || strings.Contains(line, "Videau:") ||
			strings.Contains(line, "Dopplerwürfel:") || strings.Contains(line, "キューブ:") {
			continue
		}

		// Skip board display lines (position is in XGID)
		if strings.Contains(line, "+13-14-15-16-17-18") {
			inBoard = true
			continue
		}
		if inBoard {
			if strings.Contains(line, "+12-11-10--9--8--7") {
				inBoard = false
			}
			continue
		}

		// Detect action type from "to play" or "on roll" lines
		if parsePlayerToMove(line, pos) {
			continue
		}

		// Parse move analysis
		if move, ok := parseMoveAnalysis(line); ok {
			pos.Analysis = append(pos.Analysis, move)
			continue
		}

		// Parse cube analysis
		if parseCubeAnalysis(line, pos, scanner) {
			continue
		}

		// Parse version info
		if strings.Contains(line, "eXtreme Gammon Version") {
			parseVersionInfo(line, pos)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate required fields
	if pos.XGID == "" {
		return nil, fmt.Errorf("missing XGID in text position")
	}

	return pos, nil
}

// parsePlayerToMove parses which player is to move and action type
func parsePlayerToMove(line string, pos *XGTextPosition) bool {
	// English: "X to play 21" or "X on roll, cube action"
	// French: "X à jouer 21"
	// German: "X zum spielen 21"

	movePatterns := []string{
		`([XO])\s+to play\s+(\d+)`,
		`([XO])\s+à jouer\s+(\d+)`,
		`([XO])\s+zum spielen\s+(\d+)`,
		`([XO])\s+on roll`,
	}

	for _, pattern := range movePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); matches != nil {
			if len(matches) > 2 && matches[2] != "" {
				pos.ActionType = "play"
			} else {
				pos.ActionType = "cube"
			}
			return true
		}
	}
	return false
}

// parseMoveAnalysis parses a move analysis line
func parseMoveAnalysis(line string) (XGMove, bool) {
	// "    1. 4-ply       19/18 14/12                  eq:-0.491"
	// "      Player:   25.45% (G:0.00% B:0.00%)"
	// "      Opponent: 74.55% (G:31.09% B:0.09%)"

	// French: "    1. 4-plis      19/18 14/12                  éq:-0.491"
	// "      Joueur:     25.45% (G:0.00% B:0.00%)"
	// "      Adversaire: 74.55% (G:31.09% B:0.09%)"

	// German: "    1. 4-ply       19/18 14/12                  eq:-0.491"
	// "      Spieler: 25.45% (G:0.00% B:0.00%)"
	// "      Gegner:  74.55% (G:31.09% B:0.09%)"

	// Japanese: "      プレーヤー: 25.45% (G:0.00% B:0.00%)"
	// "      対戦相手:  74.55% (G:31.09% B:0.09%)"

	movePattern := `^\s+(\d+)\.\s+(\d+)-pl(?:y|is)\s+(.+?)\s+(?:eq|éq):([+-]?\d+\.\d+)(?:\s+\(([+-]?\d+\.\d+)\))?`
	re := regexp.MustCompile(movePattern)

	if matches := re.FindStringSubmatch(line); matches != nil {
		move := XGMove{}
		move.Rank, _ = strconv.Atoi(matches[1])
		move.Ply, _ = strconv.Atoi(matches[2])
		move.Move = strings.TrimSpace(matches[3])
		move.Equity, _ = strconv.ParseFloat(matches[4], 64)
		if len(matches) > 5 && matches[5] != "" {
			move.EquityDiff, _ = strconv.ParseFloat(matches[5], 64)
		}
		return move, true
	}

	return XGMove{}, false
}

// parseCubeAnalysis parses cube action analysis
func parseCubeAnalysis(line string, pos *XGTextPosition, scanner *bufio.Scanner) bool {
	// English: "Analyzed in XG Roller++"
	// French: "Analysé avec XG Roller++"
	// German: "Analysiert in XG Roller++"
	// Japanese: "XG Roller++で分析済み"
	if !strings.Contains(line, "Analyzed in") &&
		!strings.Contains(line, "Analysé avec") &&
		!strings.Contains(line, "Analysiert in") &&
		!strings.Contains(line, "で分析済み") {
		return false
	}

	if pos.CubeAnalysis == nil {
		pos.CubeAnalysis = &XGCubeAnalysis{}
	}

	// Extract analyzer type
	if strings.Contains(line, "XG Roller++") {
		pos.CubeAnalysis.AnalyzerType = "XG Roller++"
	}

	// Continue parsing following lines for cube analysis
	for scanner.Scan() {
		nextLine := scanner.Text()

		// Player winning chances
		// English: "Player Winning Chances:"
		// French: "Chance de gain du joueur:"
		// German: "Spieler Gewinnchancen:"
		if strings.Contains(nextLine, "Player Winning") ||
			strings.Contains(nextLine, "gain du joueur") ||
			strings.Contains(nextLine, "Spieler Gewinnchancen") ||
			strings.Contains(nextLine, "プレーヤー") {
			parseWinningChances(nextLine, pos.CubeAnalysis, true)
		}

		// Opponent winning chances
		// English: "Opponent Winning Chances:"
		// French: "Chance de gain de l'adversaire:"
		// German: "Gewinnchancen des Gegners:"
		if strings.Contains(nextLine, "Opponent Winning") ||
			strings.Contains(nextLine, "gain de l'adversaire") ||
			strings.Contains(nextLine, "Gewinnchancen des Gegners") ||
			strings.Contains(nextLine, "対戦相手") {
			parseWinningChances(nextLine, pos.CubeAnalysis, false)
		}

		// Cubeless equities
		if strings.Contains(nextLine, "Cubeless Equities") ||
			strings.Contains(nextLine, "Equités sans videau") ||
			strings.Contains(nextLine, "Equities ohne Dopplerwürfel") {
			parseCubelessEquities(nextLine, pos.CubeAnalysis)
		}

		// Cubeful equities
		// Note: We use same fields for both double and redouble scenarios
		if strings.Contains(nextLine, "No double") || strings.Contains(nextLine, "No redouble") ||
			strings.Contains(nextLine, "Pas de double:") || strings.Contains(nextLine, "Pas de redouble:") ||
			strings.Contains(nextLine, "Nicht Doppeln:") || strings.Contains(nextLine, "Nicht Redoppeln:") ||
			strings.Contains(nextLine, "ノーダブル:") || strings.Contains(nextLine, "ノーリダブル:") {
			parseCubefulEquity(nextLine, pos.CubeAnalysis, "no_double")
		}
		if strings.Contains(nextLine, "Double/Take") || strings.Contains(nextLine, "Redouble/Take") ||
			strings.Contains(nextLine, "Double/Prend") || strings.Contains(nextLine, "Redouble/Prend") ||
			strings.Contains(nextLine, "Doppeln/Annehmen") || strings.Contains(nextLine, "Redoppeln/Annehmen") ||
			strings.Contains(nextLine, "ダブル/テイク") || strings.Contains(nextLine, "リダブル/テイク") {
			parseCubefulEquity(nextLine, pos.CubeAnalysis, "double_take")
		}
		if strings.Contains(nextLine, "Double/Pass") || strings.Contains(nextLine, "Redouble/Pass") ||
			strings.Contains(nextLine, "Double/Passe") || strings.Contains(nextLine, "Redouble/Passe") ||
			strings.Contains(nextLine, "Doppeln/Passe") || strings.Contains(nextLine, "Redoppeln/Ablehnen") ||
			strings.Contains(nextLine, "ダブル/パス") || strings.Contains(nextLine, "リダブル/パス") {
			parseCubefulEquity(nextLine, pos.CubeAnalysis, "double_drop")
		}

		// Parse cube recommendation
		// English: "Best Cube action: Double / Take"
		// French: "Meilleur action du videau: Double / Prend"
		// German: "Beste Dopplerwürfel Aktion Doppeln / Ablehnen"
		// Japanese: "ベストキューブアクション：ツーグッド / パス"
		if strings.Contains(nextLine, "Best Cube action:") ||
			strings.Contains(nextLine, "Meilleur action du videau:") ||
			strings.Contains(nextLine, "Beste Dopplerwürfel") ||
			strings.Contains(nextLine, "ベストキューブアクション") {
			parseCubeRecommendation(nextLine, pos.CubeAnalysis)
		}

		// Stop at version line (end of cube analysis section)
		if strings.HasPrefix(nextLine, "eXtreme Gammon") {
			parseVersionInfo(nextLine, pos)
			break
		}

		// Skip empty lines but continue parsing
		if strings.TrimSpace(nextLine) == "" {
			continue
		}
	}

	return true
}

// parseWinningChances parses winning chances line
func parseWinningChances(line string, cube *XGCubeAnalysis, isPlayer bool) {
	// "Player Winning Chances:   54.40% (G:18.22% B:0.53%)"
	re := regexp.MustCompile(`(\d+\.\d+)%\s+\(G:(\d+\.\d+)%\s+B:(\d+\.\d+)%\)`)
	if matches := re.FindStringSubmatch(line); matches != nil {
		win, _ := strconv.ParseFloat(matches[1], 64)
		g, _ := strconv.ParseFloat(matches[2], 64)
		b, _ := strconv.ParseFloat(matches[3], 64)

		if isPlayer {
			cube.PlayerWin = win
			cube.PlayerG = g
			cube.PlayerB = b
		} else {
			cube.OppWin = win
			cube.OppG = g
			cube.OppB = b
		}
	}
}

// parseCubelessEquities parses cubeless equity line
func parseCubelessEquities(line string, cube *XGCubeAnalysis) {
	// English: "Cubeless Equities: No Double=+0.103, Double=+0.259"
	// French: "Equités sans videau: Pas de double=+0.103, Double=+0.259"
	// German: "Equities ohne Dopplerwürfel: Nicht Doppeln=+0.103, Doppeln=+0.259"

	patterns := []string{
		`No Double=([+-]?\d+\.\d+),\s*Double=([+-]?\d+\.\d+)`,
		`Pas de double=([+-]?\d+\.\d+),\s*Double=([+-]?\d+\.\d+)`,
		`Nicht Doppeln=([+-]?\d+\.\d+),\s*Doppeln=([+-]?\d+\.\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); matches != nil {
			cube.CubelessNoDouble, _ = strconv.ParseFloat(matches[1], 64)
			cube.CubelessDouble, _ = strconv.ParseFloat(matches[2], 64)
			return
		}
	}
}

// parseCubefulEquity parses cubeful equity lines
func parseCubefulEquity(line string, cube *XGCubeAnalysis, equityType string) {
	// "       No double:     +0.337"
	// "       Double/Take:   +0.215 (-0.122)"
	re := regexp.MustCompile(`([+-]?\d+\.\d+)(?:\s+\(([+-]?\d+\.\d+)\))?`)
	if matches := re.FindStringSubmatch(line); matches != nil {
		equity, _ := strconv.ParseFloat(matches[1], 64)

		switch equityType {
		case "no_double":
			cube.NoDouble = equity
		case "double_take":
			cube.DoubleTake = equity
		case "double_drop":
			cube.DoubleDrop = equity
		}
	}
}

// parseCubeRecommendation parses the cube recommendation line
func parseCubeRecommendation(line string, cube *XGCubeAnalysis) {
	// English: "Best Cube action: Double / Take"
	// French: "Meilleur action du videau: Double / Prend"
	// German: "Beste Dopplerwürfel Aktion Doppeln / Annehmen" (no colon!)
	// Japanese: "ベストキューブアクション：ツーグッド / パス" (Japanese colon "：")

	// Try to extract after regular colon first
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		cube.Recommendation = strings.TrimSpace(parts[1])
		return
	}

	// Try Japanese colon "："
	parts = strings.SplitN(line, "：", 2)
	if len(parts) == 2 {
		cube.Recommendation = strings.TrimSpace(parts[1])
		return
	}

	// For German, extract after "Aktion"
	if strings.Contains(line, "Beste Dopplerwürfel Aktion") {
		rec := strings.TrimPrefix(line, "Beste Dopplerwürfel Aktion")
		cube.Recommendation = strings.TrimSpace(rec)
	}
}

// parseVersionInfo parses eXtreme Gammon version information
func parseVersionInfo(line string, pos *XGTextPosition) {
	// "eXtreme Gammon Version: 2.10, MET: Kazaross XG2"
	// "eXtreme Gammon Version: 2.10, TEM: Kazaross XG2" (French)
	parts := strings.Split(line, ",")
	if len(parts) >= 1 {
		versionPart := strings.TrimSpace(parts[0])
		versionPart = strings.TrimPrefix(versionPart, "eXtreme Gammon Version:")
		pos.Version = strings.TrimSpace(versionPart)
	}
	if len(parts) >= 2 {
		metPart := strings.TrimSpace(parts[1])
		metPart = strings.TrimPrefix(metPart, "MET:")
		metPart = strings.TrimPrefix(metPart, "TEM:")
		pos.MET = strings.TrimSpace(metPart)
	}
}

// ToJSON converts the text position to a JSON-friendly map
func (p *XGTextPosition) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"xgid":        p.XGID,
		"player1":     p.Player1Name,
		"player2":     p.Player2Name,
		"action_type": p.ActionType,
		"version":     p.Version,
		"met":         p.MET,
	}

	if len(p.Analysis) > 0 {
		moves := make([]map[string]interface{}, 0, len(p.Analysis))
		for _, move := range p.Analysis {
			moves = append(moves, map[string]interface{}{
				"rank":        move.Rank,
				"ply":         move.Ply,
				"move":        move.Move,
				"equity":      move.Equity,
				"equity_diff": move.EquityDiff,
			})
		}
		result["moves"] = moves
	}

	if p.CubeAnalysis != nil {
		result["cube_analysis"] = map[string]interface{}{
			"player_win":          p.CubeAnalysis.PlayerWin,
			"player_gammon":       p.CubeAnalysis.PlayerG,
			"player_backgammon":   p.CubeAnalysis.PlayerB,
			"opponent_win":        p.CubeAnalysis.OppWin,
			"opponent_gammon":     p.CubeAnalysis.OppG,
			"opponent_backgammon": p.CubeAnalysis.OppB,
			"cubeless_no_double":  p.CubeAnalysis.CubelessNoDouble,
			"cubeless_double":     p.CubeAnalysis.CubelessDouble,
			"no_double":           p.CubeAnalysis.NoDouble,
			"double_take":         p.CubeAnalysis.DoubleTake,
			"double_drop":         p.CubeAnalysis.DoubleDrop,
			"recommendation":      p.CubeAnalysis.Recommendation,
			"analyzer":            p.CubeAnalysis.AnalyzerType,
		}
	}

	return result
}
