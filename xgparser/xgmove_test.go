package xgparser

import (
	"math"
	"testing"
)

func TestDecodeXGMove(t *testing.T) {
	u := int32(-1)
	tests := []struct {
		name string
		raw  [8]int32
		want [8]int8
	}{
		{"ordinary move, junk after terminator", [8]int32{7, 2, 5, 2, u, 1, 2, 1}, [8]int8{8, 3, 6, 3, -1, -1, -1, -1}},
		{"enter from the bar", [8]int32{24, 21, 24, 19, u, 6, 6, 0}, [8]int8{25, 22, 25, 20, -1, -1, -1, -1}},
		{"played record: off is -1", [8]int32{4, u, 4, 2, u, u, u, u}, [8]int8{5, -2, 5, 3, -1, -1, -1, -1}},
		{"played record: four bear-offs", [8]int32{5, u, 5, u, 3, u, 3, u}, [8]int8{6, -2, 6, -2, 4, -2, 4, -2}},
		{"candidate: off is from-die", [8]int32{3, -1, 3, -2, u, 3, 5, 3}, [8]int8{4, -2, 4, -2, -1, -1, -1, -1}},
		{"candidate: off then more sub-moves", [8]int32{4, 3, 3, 2, 0, -1, 0, -1}, [8]int8{5, 4, 4, 3, 1, -2, 1, -2}},
		{"candidate: doubles off with terminator", [8]int32{1, -3, 0, -4, 0, -4, u, 9}, [8]int8{2, -2, 1, -2, 1, -2, -1, -1}},
		{"no move", [8]int32{u, u, u, u, u, u, u, u}, [8]int8{-1, -1, -1, -1, -1, -1, -1, -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeXGMove(tt.raw); got != tt.want {
				t.Errorf("decodeXGMove(%v) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

// replayForMover applies a decoded move to a board seen from the player on
// roll (positive checkers) and returns that player's checkers only.
func replayForMover(board [26]int8, move [8]int8) [26]int8 {
	for j := 0; j < 8; j += 2 {
		from, to := move[j], move[j+1]
		if from == -1 {
			break
		}
		if from < 1 || from > 25 || to == 0 || to < -2 || to > 24 || to == -1 {
			board[0] = 127 // not a legal decoded move: matches no position
			return board
		}
		board[from]--
		if to == -2 {
			continue
		}
		if board[to] < 0 {
			board[to] = 0
		}
		board[to]++
	}
	return moverCheckers(board)
}

func moverCheckers(board [26]int8) [26]int8 {
	for i := range board {
		if board[i] < 0 {
			board[i] = 0
		}
	}
	return board
}

func isPlaceholder(move [8]int32) bool {
	return move[0] == move[1]
}

// playedAsDecoded reads PlayedMove, whose bear-offs keep their historical
// encoding (a "to" of 0 or less), in the candidates' format.
func playedAsDecoded(played [8]int32) [8]int8 {
	out := [8]int8{-1, -1, -1, -1, -1, -1, -1, -1}
	for j := 0; j < 8; j += 2 {
		if played[j] == -1 {
			break
		}
		out[j] = int8(played[j])
		out[j+1] = int8(played[j+1])
		if played[j+1] <= 0 {
			out[j+1] = -2
		}
	}
	return out
}

// Every decoded candidate must lead to the position XG recorded for it, and
// every played move must be one of the candidates.
func TestCandidateMovesReplayToRecordedPositions(t *testing.T) {
	for _, file := range []string{"testdata/bearoff_gammonnet_7p.xg", "testdata/test.xg", "testdata/match_with_comment.xg"} {
		t.Run(file, func(t *testing.T) {
			m, err := ParseXGFromFile(file)
			if err != nil {
				t.Fatal(err)
			}
			bearOffCandidates := 0
			for _, g := range m.Games {
				for k, mv := range g.Moves {
					c := mv.CheckerMove
					if c == nil || len(c.Analysis) == 0 {
						continue
					}
					playedFound := false
					playedEnd := replayForMover(c.Position.Checkers, playedAsDecoded(c.PlayedMove))
					for i, a := range c.Analysis {
						if a.Move == [8]int8{1, 1, 1, 1, 1, 1, 1, 1} {
							continue // XG's placeholder for "no legal move"
						}
						for j := 1; j < 8; j += 2 {
							if a.Move[j] == -2 {
								bearOffCandidates++
								break
							}
						}
						want := moverCheckers(a.Position.Checkers)
						if got := replayForMover(c.Position.Checkers, a.Move); got != want {
							t.Errorf("game %d move %d candidate %d %v: replay does not reach XG's position", g.GameNumber, k, i, a.Move)
						}
						if want == playedEnd {
							playedFound = true
						}
					}
					if !isPlaceholder(c.PlayedMove) && !playedFound {
						t.Errorf("game %d move %d: played %v is not among the candidates", g.GameNumber, k, c.PlayedMove)
					}
				}
			}
			if bearOffCandidates == 0 {
				t.Fatal("fixture has no bear-off candidate")
			}
		})
	}
}

// gammonNet's bear-offs in games 3 and 4 of the fixture: the played move must
// decode in full both as played and as a candidate, so its error is XG's.
func TestBearOffPlayedMoveMatchesCandidate(t *testing.T) {
	m, err := ParseXGFromFile("testdata/bearoff_gammonnet_7p.xg")
	if err != nil {
		t.Fatal(err)
	}
	// played is PlayedMove as stored (historical encoding), want the
	// candidate it must match.
	cases := []struct {
		game   int32
		dice   [2]int32
		played [8]int32
		want   [8]int8
		err    float64
	}{
		{3, [2]int32{3, 2}, [8]int32{3, -1, 2, -1, -1, -1, -1, -1}, [8]int8{3, -2, 2, -2, -1, -1, -1, -1}, 0.0013}, // 3/off 2/off
		{4, [2]int32{1, 1}, [8]int32{5, 4, 4, 3, 1, -1, 1, -1}, [8]int8{5, 4, 4, 3, 1, -2, 1, -2}, 0.0205},         // 5/3 1/off(2)
	}
	for _, tc := range cases {
		found := false
		for _, g := range m.Games {
			if g.GameNumber != tc.game {
				continue
			}
			for _, mv := range g.Moves {
				c := mv.CheckerMove
				if c == nil || c.Dice != tc.dice || c.PlayedMove != tc.played {
					continue
				}
				found = true
				want := tc.want
				matched := false
				for _, a := range c.Analysis {
					if a.Move != want {
						continue
					}
					matched = true
					if got := float64(c.Analysis[0].Equity - a.Equity); math.Abs(got-tc.err) > 0.00005 {
						t.Errorf("game %d %v: error %.4f, want %.4f", tc.game, tc.want, got, tc.err)
					}
				}
				if !matched {
					t.Errorf("game %d: played %v not among candidates", tc.game, tc.want)
				}
			}
		}
		if !found {
			t.Errorf("game %d: played move %v not found", tc.game, tc.played)
		}
	}
}
