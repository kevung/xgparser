package main

import (
	"fmt"
	"os"

	"github.com/unger/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <xgfile>\n", os.Args[0])
		os.Exit(1)
	}

	xgFilename := os.Args[1]
	fmt.Printf("Processing file: %s\n", xgFilename)

	imp := xgparser.NewImport(xgFilename)
	segments, err := imp.GetFileSegments()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing file: %v\n", err)
		os.Exit(1)
	}

	fileVersion := int32(-1)

	for _, segment := range segments {
		if segment.Type == xgparser.SegmentXGGameFile {
			records, err := xgparser.ParseGameFile(segment.Data, fileVersion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing game file: %v\n", err)
				os.Exit(1)
			}

			for _, rec := range records {
				switch r := rec.(type) {
				case *xgparser.HeaderMatchEntry:
					fileVersion = r.Version
					printHeaderMatchEntry(r)
				case *xgparser.HeaderGameEntry:
					printHeaderGameEntry(r)
				case *xgparser.CubeEntry:
					printCubeEntry(r)
				case *xgparser.MoveEntry:
					printMoveEntry(r)
				case *xgparser.FooterGameEntry:
					printFooterGameEntry(r)
				case *xgparser.FooterMatchEntry:
					printFooterMatchEntry(r)
				}
			}
		}
	}
}

func printHeaderMatchEntry(h *xgparser.HeaderMatchEntry) {
	fmt.Printf("{'AddtoProfile1': %v,\n", h.AddtoProfile1)
	fmt.Printf(" 'AddtoProfile2': %v,\n", h.AddtoProfile2)
	fmt.Printf(" 'AutoDouble': %v,\n", h.AutoDouble)
	fmt.Printf(" 'AutoDoubleMax': %d,\n", h.AutoDoubleMax)
	fmt.Printf(" 'Beaver': %v,\n", h.Beaver)
	fmt.Printf(" 'CommentFooterMatch': %d,\n", h.CommentFooterMatch)
	fmt.Printf(" 'CommentHeaderMatch': %d,\n", h.CommentHeaderMatch)
	fmt.Printf(" 'CompLevel1': %d,\n", h.CompLevel1)
	fmt.Printf(" 'CompLevel2': %d,\n", h.CompLevel2)
	fmt.Printf(" 'CountForElo': %v,\n", h.CountForElo)
	fmt.Printf(" 'Counted': %v,\n", h.Counted)
	fmt.Printf(" 'Crawford': %v,\n", h.Crawford)
	fmt.Printf(" 'CubeLimit': %d,\n", h.CubeLimit)
	fmt.Printf(" 'Currency': %d,\n", h.Currency)
	fmt.Printf(" 'Date': '%s',\n", h.Date)
	fmt.Printf(" 'Elo1': %g,\n", h.Elo1)
	fmt.Printf(" 'Elo2': %g,\n", h.Elo2)
	fmt.Printf(" 'Entered': %v,\n", h.Entered)
	fmt.Printf(" 'EntryType': %d,\n", h.EntryType)
	fmt.Printf(" 'Event': b'%s',\n", h.Event)
	fmt.Printf(" 'Exp1': %d,\n", h.Exp1)
	fmt.Printf(" 'Exp2': %d,\n", h.Exp2)
	fmt.Printf(" 'FeeMoney': %g,\n", h.FeeMoney)
	fmt.Printf(" 'GameId': %d,\n", h.GameId)
	fmt.Printf(" 'GameMode': %d,\n", h.GameMode)
	fmt.Printf(" 'Imported': %v,\n", h.Imported)
	fmt.Printf(" 'Invert': %d,\n", h.Invert)
	fmt.Printf(" 'Jacoby': %v,\n", h.Jacoby)
	fmt.Printf(" 'Location': b'%s',\n", h.Location)
	fmt.Printf(" 'LoseMoney': %g,\n", h.LoseMoney)
	fmt.Printf(" 'Magic': %d,\n", h.Magic)
	fmt.Printf(" 'MatchLength': %d,\n", h.MatchLength)
	fmt.Printf(" 'MoneyInitG': %d,\n", h.MoneyInitG)
	fmt.Printf(" 'MoneyInitScore': (%d, %d),\n", h.MoneyInitScore[0], h.MoneyInitScore[1])
	fmt.Printf(" 'Name': '%s',\n", h.Name)
	fmt.Printf(" 'Player1': b'%s',\n", h.Player1)
	fmt.Printf(" 'Player2': b'%s',\n", h.Player2)
	fmt.Printf(" 'Round': b'%s',\n", h.Round)
	fmt.Printf(" 'SEvent': '%s',\n", h.SEvent)
	fmt.Printf(" 'SLocation': '%s',\n", h.SLocation)
	fmt.Printf(" 'SPlayer1': '%s',\n", h.SPlayer1)
	fmt.Printf(" 'SPlayer2': '%s',\n", h.SPlayer2)
	fmt.Printf(" 'SRound': '%s',\n", h.SRound)
	fmt.Printf(" 'SiteId': %d,\n", h.SiteId)
	fmt.Printf(" 'TableStake': %d,\n", h.TableStake)
	if h.TimeSetting != nil {
		fmt.Printf(" 'TimeSetting': {'ClockType': %d, 'Penalty': %d, 'PenaltyMoney': %d, 'PerGame': %v, 'Time1': %d, 'Time2': %d, 'TimeLeft1': %d, 'TimeLeft2': %d},\n",
			h.TimeSetting.ClockType, h.TimeSetting.Penalty, h.TimeSetting.PenaltyMoney, h.TimeSetting.PerGame,
			h.TimeSetting.Time1, h.TimeSetting.Time2, h.TimeSetting.TimeLeft1, h.TimeSetting.TimeLeft2)
	}
	fmt.Printf(" 'TotTimeDelayCube': %d,\n", h.TotTimeDelayCube)
	fmt.Printf(" 'TotTimeDelayCubeDone': %d,\n", h.TotTimeDelayCubeDone)
	fmt.Printf(" 'TotTimeDelayMove': %d,\n", h.TotTimeDelayMove)
	fmt.Printf(" 'TotTimeDelayMoveDone': %d,\n", h.TotTimeDelayMoveDone)
	fmt.Printf(" 'Transcribed': %v,\n", h.Transcribed)
	fmt.Printf(" 'Transcriber': b'%s',\n", h.Transcriber)
	fmt.Printf(" 'UnratedImp': %v,\n", h.UnratedImp)
	fmt.Printf(" 'Variation': %d,\n", h.Variation)
	fmt.Printf(" 'Version': %d,\n", h.Version)
	fmt.Printf(" 'WinMoney': %g,\n", h.WinMoney)
	fmt.Printf(" 'isMoneyMatch': %v}\n", h.IsMoneyMatch)
}

func printHeaderGameEntry(h *xgparser.HeaderGameEntry) {
	fmt.Printf("{'CommentFooterGame': %d,\n", h.CommentFooterGame)
	fmt.Printf(" 'CommentHeaderGame': %d,\n", h.CommentHeaderGame)
	fmt.Printf(" 'CrawfordApply': %v,\n", h.CrawfordApply)
	fmt.Printf(" 'EntryType': %d,\n", h.EntryType)
	fmt.Printf(" 'GameNumber': %d,\n", h.GameNumber)
	fmt.Printf(" 'InProgress': %v,\n", h.InProgress)
	fmt.Printf(" 'Name': '%s',\n", h.Name)
	fmt.Printf(" 'NumberOfAutoDoubles': %d,\n", h.NumberOfAutoDoubles)
	fmt.Printf(" 'PosInit': (")
	for i, v := range h.PosInit {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'Score1': %d,\n", h.Score1)
	fmt.Printf(" 'Score2': %d,\n", h.Score2)
	fmt.Printf(" 'Version': %d}\n", h.Version)
}

func printCubeEntry(c *xgparser.CubeEntry) {
	fmt.Printf("{'ActiveP': %d,\n", c.ActiveP)
	fmt.Printf(" 'AnalyzeC': %d,\n", c.AnalyzeC)
	fmt.Printf(" 'AnalyzeCR': %d,\n", c.AnalyzeCR)
	fmt.Printf(" 'BeaverR': %d,\n", c.BeaverR)
	fmt.Printf(" 'CommentCube': %d,\n", c.CommentCube)
	fmt.Printf(" 'CompChoiceD': %d,\n", c.CompChoiceD)
	fmt.Printf(" 'CubeB': %d,\n", c.CubeB)
	fmt.Printf(" 'DiceRolled': '%s',\n", c.DiceRolled)
	fmt.Printf(" 'Double': %d,\n", c.Double)
	if c.Doubled != nil {
		fmt.Printf(" 'Doubled': {'Crawford': %d,\n", c.Doubled.Crawford)
		fmt.Printf("             'Cube': %d,\n", c.Doubled.Cube)
		fmt.Printf("             'CubePos': %d,\n", c.Doubled.CubePos)
		fmt.Printf("             'DoubleChoice3': %d,\n", c.Doubled.DoubleChoice3)
		fmt.Printf("             'Eval': (")
		for i, v := range c.Doubled.Eval {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%g", v)
		}
		fmt.Printf("),\n")
		fmt.Printf("             'EvalDouble': (")
		for i, v := range c.Doubled.EvalDouble {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%g", v)
		}
		fmt.Printf("),\n")
		fmt.Printf("             'FlagDouble': %d,\n", c.Doubled.FlagDouble)
		fmt.Printf("             'Jacoby': %d,\n", c.Doubled.Jacoby)
		fmt.Printf("             'Level': %d,\n", c.Doubled.Level)
		fmt.Printf("             'LevelRequest': %d,\n", c.Doubled.LevelRequest)
		fmt.Printf("             'Pos': (")
		for i, v := range c.Doubled.Pos {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", v)
		}
		fmt.Printf("),\n")
		fmt.Printf("             'Score': (%d, %d),\n", c.Doubled.Score[0], c.Doubled.Score[1])
		fmt.Printf("             'equB': %g,\n", c.Doubled.EquB)
		fmt.Printf("             'equDouble': %g,\n", c.Doubled.EquDouble)
		fmt.Printf("             'equDrop': %g,\n", c.Doubled.EquDrop)
		fmt.Printf("             'isBeaver': %d,\n", c.Doubled.IsBeaver)
		fmt.Printf("             'met': %d},\n", c.Doubled.Met)
	}
	fmt.Printf(" 'EditedCube': %v,\n", c.EditedCube)
	fmt.Printf(" 'EntryType': %d,\n", c.EntryType)
	fmt.Printf(" 'ErrBeaver': %g,\n", c.ErrBeaver)
	fmt.Printf(" 'ErrCube': %g,\n", c.ErrCube)
	fmt.Printf(" 'ErrRaccoon': %g,\n", c.ErrRaccoon)
	fmt.Printf(" 'ErrTake': %g,\n", c.ErrTake)
	fmt.Printf(" 'ErrTutorCube': %g,\n", c.ErrTutorCube)
	fmt.Printf(" 'ErrTutorTake': %g,\n", c.ErrTutorTake)
	fmt.Printf(" 'FlaggedDouble': %v,\n", c.FlaggedDouble)
	fmt.Printf(" 'Name': '%s',\n", c.Name)
	fmt.Printf(" 'NumberOfAutoDoubleCube': %d,\n", c.NumberOfAutoDoubleCube)
	fmt.Printf(" 'Position': (")
	for i, v := range c.Position {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'RaccoonR': %d,\n", c.RaccoonR)
	fmt.Printf(" 'RolloutIndexD': %d,\n", c.RolloutIndexD)
	fmt.Printf(" 'Take': %d,\n", c.Take)
	fmt.Printf(" 'TimeBot': %d,\n", c.TimeBot)
	fmt.Printf(" 'TimeDelayCube': %v,\n", c.TimeDelayCube)
	fmt.Printf(" 'TimeDelayCubeDone': %v,\n", c.TimeDelayCubeDone)
	fmt.Printf(" 'TimeTop': %d,\n", c.TimeTop)
	fmt.Printf(" 'TutorCube': %d,\n", c.TutorCube)
	fmt.Printf(" 'TutorTake': %d,\n", c.TutorTake)
	fmt.Printf(" 'Version': %d,\n", c.Version)
	fmt.Printf(" 'isValid': %d}\n", c.IsValid)
}

func printMoveEntry(m *xgparser.MoveEntry) {
	fmt.Printf("{'ActiveP': %d,\n", m.ActiveP)
	fmt.Printf(" 'AnalyzeL': %d,\n", m.AnalyzeL)
	fmt.Printf(" 'AnalyzeM': %d,\n", m.AnalyzeM)
	fmt.Printf(" 'CommentMove': %d,\n", m.CommentMove)
	fmt.Printf(" 'CompChoice': %d,\n", m.CompChoice)
	fmt.Printf(" 'CubeA': %d,\n", m.CubeA)

	if m.DataMoves != nil {
		fmt.Printf(" 'DataMoves': {'Choice0': %d,\n", m.DataMoves.Choice0)
		fmt.Printf("               'Choice3': %d,\n", m.DataMoves.Choice3)
		fmt.Printf("               'Crawford': %d,\n", m.DataMoves.Crawford)
		fmt.Printf("               'Cube': %d,\n", m.DataMoves.Cube)
		fmt.Printf("               'CubePos': %d,\n", m.DataMoves.CubePos)
		fmt.Printf("               'Cubepos': %d,\n", m.DataMoves.CubePos)
		fmt.Printf("               'Dice': (%d, %d),\n", m.DataMoves.Dice[0], m.DataMoves.Dice[1])

		// Print Eval
		fmt.Printf("               'Eval': (")
		for i := 0; i < int(m.DataMoves.NMoves); i++ {
			if i > 0 {
				fmt.Printf(",\n                        ")
			} else {
				fmt.Printf("(")
			}
			for j := 0; j < 7; j++ {
				if j > 0 {
					fmt.Printf(",\n                         ")
				}
				fmt.Printf("%g", m.DataMoves.Eval[i][j])
			}
			fmt.Printf(")")
		}
		// Fill remaining with zeros
		for i := int(m.DataMoves.NMoves); i < 32; i++ {
			fmt.Printf(",\n                        (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0)")
		}
		fmt.Printf("),\n")

		// Print EvalLevel
		fmt.Printf("               'EvalLevel': (")
		for i := 0; i < 32; i++ {
			if i > 0 {
				fmt.Printf(",\n                             ")
			}
			fmt.Printf("{'Level': %d, 'isDouble': %v}", m.DataMoves.EvalLevel[i].Level, m.DataMoves.EvalLevel[i].IsDouble)
		}
		fmt.Printf("),\n")

		fmt.Printf("               'Jacoby': %d,\n", m.DataMoves.Jacoby)
		fmt.Printf("               'Level': %d,\n", m.DataMoves.Level)

		// Print Moves
		fmt.Printf("               'Moves': (")
		for i := 0; i < 32; i++ {
			if i > 0 {
				fmt.Printf(",\n                         ")
			}
			fmt.Printf("(")
			for j := 0; j < 8; j++ {
				if j > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d", m.DataMoves.Moves[i][j])
			}
			fmt.Printf(")")
		}
		fmt.Printf("),\n")

		fmt.Printf("               'NMoves': %d,\n", m.DataMoves.NMoves)

		// Print Pos
		fmt.Printf("               'Pos': (")
		for i, v := range m.DataMoves.Pos {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", v)
		}
		fmt.Printf("),\n")

		// Print PosPlayed
		fmt.Printf("               'PosPlayed': (")
		for i := 0; i < 32; i++ {
			if i > 0 {
				fmt.Printf(",\n                             ")
			}
			fmt.Printf("(")
			for j := 0; j < 26; j++ {
				if j > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d", m.DataMoves.PosPlayed[i][j])
			}
			fmt.Printf(")")
		}
		fmt.Printf("),\n")

		fmt.Printf("               'Score': (%d, %d),\n", m.DataMoves.Score[0], m.DataMoves.Score[1])
		fmt.Printf("               'Unused': %d,\n", m.DataMoves.Unused)
		fmt.Printf("               'met': %d},\n", m.DataMoves.Met)
	}

	fmt.Printf(" 'Dice': (%d, %d),\n", m.Dice[0], m.Dice[1])
	fmt.Printf(" 'EditedMove': %v,\n", m.EditedMove)
	fmt.Printf(" 'EntryType': %d,\n", m.EntryType)
	fmt.Printf(" 'ErrLuck': %g,\n", m.ErrLuck)
	fmt.Printf(" 'ErrMove': %g,\n", m.ErrMove)
	fmt.Printf(" 'ErrTutorMove': %g,\n", m.ErrTutorMove)
	fmt.Printf(" 'ErrorM': %g,\n", m.ErrorM)
	fmt.Printf(" 'Flagged': %v,\n", m.Flagged)
	fmt.Printf(" 'InitEq': %g,\n", m.InitEq)
	fmt.Printf(" 'InvalidM': %d,\n", m.InvalidM)
	fmt.Printf(" 'Moves': (")
	for i, v := range m.Moves {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'NMoveEval': %d,\n", m.NMoveEval)
	fmt.Printf(" 'Name:': '%s',\n", m.Name)
	fmt.Printf(" 'NumberOfAutoDoubleMove': %d,\n", m.NumberOfAutoDoubleMove)
	fmt.Printf(" 'Played': %v,\n", m.Played)
	fmt.Printf(" 'PositionEnd': (")
	for i, v := range m.PositionEnd {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'PositionI': (")
	for i, v := range m.PositionI {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'PositionTutor': (")
	for i, v := range m.PositionTutor {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'RolloutIndexM': (")
	for i, v := range m.RolloutIndexM {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'TimeDelayMove': %d,\n", m.TimeDelayMove)
	fmt.Printf(" 'TimeDelayMoveDone': %d,\n", m.TimeDelayMoveDone)
	fmt.Printf(" 'Tutor': %d,\n", m.Tutor)
	fmt.Printf(" 'Version': %d}\n", m.Version)
}

func printFooterGameEntry(f *xgparser.FooterGameEntry) {
	fmt.Printf("{'CrawfordApplyg': %v,\n", f.CrawfordApplyg)
	fmt.Printf(" 'EntryType': %d,\n", f.EntryType)
	fmt.Printf(" 'ErrResign': %g,\n", f.ErrResign)
	fmt.Printf(" 'ErrTakeResign': %g,\n", f.ErrTakeResign)
	fmt.Printf(" 'Eval': (")
	for i, v := range f.Eval {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%g", v)
	}
	fmt.Printf("),\n")
	fmt.Printf(" 'EvalLevel': %d,\n", f.EvalLevel)
	fmt.Printf(" 'Name': '%s',\n", f.Name)
	fmt.Printf(" 'PointsWon': %d,\n", f.PointsWon)
	fmt.Printf(" 'Score1g': %d,\n", f.Score1g)
	fmt.Printf(" 'Score2g': %d,\n", f.Score2g)
	fmt.Printf(" 'Termination': %d,\n", f.Termination)
	fmt.Printf(" 'Version': %d,\n", f.Version)
	fmt.Printf(" 'Winner': %d}\n", f.Winner)
}

func printFooterMatchEntry(f *xgparser.FooterMatchEntry) {
	fmt.Printf("{'Datem': '%s',\n", f.Datem)
	fmt.Printf(" 'Elo1m': %g,\n", f.Elo1m)
	fmt.Printf(" 'Elo2m': %g,\n", f.Elo2m)
	fmt.Printf(" 'EntryType': %d,\n", f.EntryType)
	fmt.Printf(" 'Exp1m': %d,\n", f.Exp1m)
	fmt.Printf(" 'Exp2m': %d,\n", f.Exp2m)
	fmt.Printf(" 'Name': '%s',\n", f.Name)
	fmt.Printf(" 'Score1m': %d,\n", f.Score1m)
	fmt.Printf(" 'Score2m': %d,\n", f.Score2m)
	fmt.Printf(" 'Version': %d,\n", f.Version)
	fmt.Printf(" 'WinnerM': %d}\n", f.WinnerM)
}
