package xgparser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// GameDataFormatHdrRecord represents the game data format header
type GameDataFormatHdrRecord struct {
	MagicNumber     [4]byte
	HeaderVersion   int32
	HeaderSize      int32
	ThumbnailOffset int64
	ThumbnailSize   uint32
	GameGUID        [2]byte // Simplified - full GUID parsing omitted for brevity
	Reserved        [14]byte
	GameName        string
	SaveName        string
	LevelName       string
	Comments        string
}

// FromStream reads GameDataFormatHdrRecord from stream
func (g *GameDataFormatHdrRecord) FromStream(r io.Reader) error {
	// Read first part of header
	var hdr struct {
		Magic           [4]byte
		HeaderVersion   int32
		HeaderSize      int32
		ThumbnailOffset int64
		ThumbnailSize   uint32
		GameGUID1       uint32
		GameGUID2       uint16
		GameGUID3       uint16
		GameGUID4       byte
		GameGUID5       byte
		GameGUID6       [6]byte
	}

	err := binary.Read(r, binary.LittleEndian, &hdr)
	if err != nil {
		return err
	}

	// Reverse magic number bytes
	g.MagicNumber[0] = hdr.Magic[3]
	g.MagicNumber[1] = hdr.Magic[2]
	g.MagicNumber[2] = hdr.Magic[1]
	g.MagicNumber[3] = hdr.Magic[0]

	if string(g.MagicNumber[:]) != "HMGR" || hdr.HeaderVersion != 1 {
		return fmt.Errorf("invalid magic number or version")
	}

	g.HeaderVersion = hdr.HeaderVersion
	g.HeaderSize = hdr.HeaderSize
	g.ThumbnailOffset = hdr.ThumbnailOffset
	g.ThumbnailSize = hdr.ThumbnailSize

	// Read UTF16 strings
	gameName, err := ReadUTF16Array(r, 1024)
	if err != nil {
		return err
	}
	g.GameName = UTF16IntArrayToString(gameName)

	saveName, err := ReadUTF16Array(r, 1024)
	if err != nil {
		return err
	}
	g.SaveName = UTF16IntArrayToString(saveName)

	levelName, err := ReadUTF16Array(r, 1024)
	if err != nil {
		return err
	}
	g.LevelName = UTF16IntArrayToString(levelName)

	comments, err := ReadUTF16Array(r, 1024)
	if err != nil {
		return err
	}
	g.Comments = UTF16IntArrayToString(comments)

	return nil
}

// TimeSettingRecord represents time settings
type TimeSettingRecord struct {
	ClockType    int32
	PerGame      bool
	Time1        int32
	Time2        int32
	Penalty      int32
	TimeLeft1    int32
	TimeLeft2    int32
	PenaltyMoney int32
}

// FromStream reads TimeSettingRecord from stream
func (t *TimeSettingRecord) FromStream(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &t.ClockType)
	if err != nil {
		return err
	}

	var perGame byte
	var padding [3]byte
	binary.Read(r, binary.LittleEndian, &perGame)
	binary.Read(r, binary.LittleEndian, &padding)
	t.PerGame = perGame != 0

	binary.Read(r, binary.LittleEndian, &t.Time1)
	binary.Read(r, binary.LittleEndian, &t.Time2)
	binary.Read(r, binary.LittleEndian, &t.Penalty)
	binary.Read(r, binary.LittleEndian, &t.TimeLeft1)
	binary.Read(r, binary.LittleEndian, &t.TimeLeft2)
	binary.Read(r, binary.LittleEndian, &t.PenaltyMoney)

	return nil
}

// EvalLevelRecord represents evaluation level
type EvalLevelRecord struct {
	Level    int16
	IsDouble bool
}

// FromStream reads EvalLevelRecord from stream
func (e *EvalLevelRecord) FromStream(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &e.Level)
	if err != nil {
		return err
	}

	var isDouble byte
	var padding byte
	binary.Read(r, binary.LittleEndian, &isDouble)
	binary.Read(r, binary.LittleEndian, &padding)
	e.IsDouble = isDouble != 0

	return nil
}

// EngineStructDoubleAction represents double action analysis
type EngineStructDoubleAction struct {
	Pos           [26]int8
	Level         int32
	Score         [2]int32
	Cube          int32
	CubePos       int32
	Jacoby        int32
	Crawford      int16 // Changed from int32 to int16
	Met           int16
	FlagDouble    int16
	IsBeaver      int16
	Eval          [7]float32
	EquB          float32
	EquDouble     float32
	EquDrop       float32
	LevelRequest  int16
	DoubleChoice3 int16
	EvalDouble    [7]float32
}

// FromStream reads EngineStructDoubleAction from stream
func (e *EngineStructDoubleAction) FromStream(r io.Reader) error {
	// Read position
	err := binary.Read(r, binary.LittleEndian, &e.Pos)
	if err != nil {
		return err
	}

	// Skip 2 bytes padding
	var padding [2]byte
	binary.Read(r, binary.LittleEndian, &padding)

	binary.Read(r, binary.LittleEndian, &e.Level)
	binary.Read(r, binary.LittleEndian, &e.Score)
	binary.Read(r, binary.LittleEndian, &e.Cube)
	binary.Read(r, binary.LittleEndian, &e.CubePos)
	binary.Read(r, binary.LittleEndian, &e.Jacoby)
	binary.Read(r, binary.LittleEndian, &e.Crawford)
	binary.Read(r, binary.LittleEndian, &e.Met)
	binary.Read(r, binary.LittleEndian, &e.FlagDouble)
	binary.Read(r, binary.LittleEndian, &e.IsBeaver)
	binary.Read(r, binary.LittleEndian, &e.Eval)
	binary.Read(r, binary.LittleEndian, &e.EquB)
	binary.Read(r, binary.LittleEndian, &e.EquDouble)
	binary.Read(r, binary.LittleEndian, &e.EquDrop)
	binary.Read(r, binary.LittleEndian, &e.LevelRequest)
	binary.Read(r, binary.LittleEndian, &e.DoubleChoice3)
	binary.Read(r, binary.LittleEndian, &e.EvalDouble)

	return nil
}

// EngineStructBestMoveRecord represents best move analysis
type EngineStructBestMoveRecord struct {
	Pos       [26]int8
	Dice      [2]int32 // Fixed: was [2]int8, Python uses 2l (2 int32s)
	Level     int32
	Score     [2]int32
	Cube      int32
	CubePos   int32
	Crawford  int32
	Jacoby    int32
	NMoves    int32
	PosPlayed [32][26]int8
	Moves     [32][8]int8
	EvalLevel [32]EvalLevelRecord
	Eval      [32][7]float32
	Unused    int8
	Met       int8
	Choice0   int8
	Choice3   int8
}

// FromStream reads EngineStructBestMoveRecord from stream
func (e *EngineStructBestMoveRecord) FromStream(r io.Reader) error {
	// Read header
	err := binary.Read(r, binary.LittleEndian, &e.Pos)
	if err != nil {
		return err
	}

	var padding [2]byte
	binary.Read(r, binary.LittleEndian, &padding)

	binary.Read(r, binary.LittleEndian, &e.Dice)
	binary.Read(r, binary.LittleEndian, &e.Level)
	binary.Read(r, binary.LittleEndian, &e.Score)
	binary.Read(r, binary.LittleEndian, &e.Cube)
	binary.Read(r, binary.LittleEndian, &e.CubePos)
	binary.Read(r, binary.LittleEndian, &e.Crawford)
	binary.Read(r, binary.LittleEndian, &e.Jacoby)
	binary.Read(r, binary.LittleEndian, &e.NMoves)

	// Read positions played
	for i := 0; i < 32; i++ {
		binary.Read(r, binary.LittleEndian, &e.PosPlayed[i])
	}

	// Read moves
	for i := 0; i < 32; i++ {
		binary.Read(r, binary.LittleEndian, &e.Moves[i])
	}

	// Read eval levels
	for i := 0; i < 32; i++ {
		e.EvalLevel[i].FromStream(r)
	}

	// Read evaluations
	for i := 0; i < 32; i++ {
		binary.Read(r, binary.LittleEndian, &e.Eval[i])
	}

	binary.Read(r, binary.LittleEndian, &e.Unused)
	binary.Read(r, binary.LittleEndian, &e.Met)
	binary.Read(r, binary.LittleEndian, &e.Choice0)
	binary.Read(r, binary.LittleEndian, &e.Choice3)

	return nil
}

// HeaderMatchEntry represents match information
type HeaderMatchEntry struct {
	Name                 string
	EntryType            int
	Version              int32
	SPlayer1             string
	SPlayer2             string
	MatchLength          int32
	Variation            int32
	Crawford             bool
	Jacoby               bool
	Beaver               bool
	AutoDouble           bool
	Elo1                 float64
	Elo2                 float64
	Exp1                 int32
	Exp2                 int32
	Date                 string
	SEvent               string
	GameId               int32
	CompLevel1           int32
	CompLevel2           int32
	CountForElo          bool
	AddtoProfile1        bool
	AddtoProfile2        bool
	SLocation            string
	GameMode             int32
	Imported             bool
	SRound               string
	Invert               int32
	Magic                uint32
	MoneyInitG           int32
	MoneyInitScore       [2]int32
	Entered              bool
	Counted              bool
	UnratedImp           bool
	CommentHeaderMatch   int32
	CommentFooterMatch   int32
	IsMoneyMatch         bool
	WinMoney             float32
	LoseMoney            float32
	Currency             int32
	FeeMoney             float32
	TableStake           int32
	SiteId               int32
	CubeLimit            int32
	AutoDoubleMax        int32
	Transcribed          bool
	Event                string
	Player1              string
	Player2              string
	Location             string
	Round                string
	TimeSetting          *TimeSettingRecord
	TotTimeDelayMove     int32
	TotTimeDelayCube     int32
	TotTimeDelayMoveDone int32
	TotTimeDelayCubeDone int32
	Transcriber          string
}

// FromStream reads HeaderMatchEntry from stream
func (h *HeaderMatchEntry) FromStream(r io.Reader, version int32) error {
	h.Name = "MatchInfo"
	h.EntryType = 0 // ENTRYTYPE_HEADERMATCH

	// Skip 9 bytes
	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	// Read player names (shortstrings)
	var player1Bytes [41]byte
	var player2Bytes [41]byte
	binary.Read(r, binary.LittleEndian, &player1Bytes)
	binary.Read(r, binary.LittleEndian, &player2Bytes)
	h.SPlayer1 = DelphiShortStrToStr(player1Bytes[:])
	h.SPlayer2 = DelphiShortStrToStr(player2Bytes[:])

	var padding1 byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &h.MatchLength)
	binary.Read(r, binary.LittleEndian, &h.Variation)

	var crawford, jacoby, beaver, autoDouble byte
	binary.Read(r, binary.LittleEndian, &crawford)
	binary.Read(r, binary.LittleEndian, &jacoby)
	binary.Read(r, binary.LittleEndian, &beaver)
	binary.Read(r, binary.LittleEndian, &autoDouble)
	h.Crawford = crawford != 0
	h.Jacoby = jacoby != 0
	h.Beaver = beaver != 0
	h.AutoDouble = autoDouble != 0

	binary.Read(r, binary.LittleEndian, &h.Elo1)
	binary.Read(r, binary.LittleEndian, &h.Elo2)
	binary.Read(r, binary.LittleEndian, &h.Exp1)
	binary.Read(r, binary.LittleEndian, &h.Exp2)

	var dateTime float64
	binary.Read(r, binary.LittleEndian, &dateTime)
	h.Date = DelphiDateTimeConv(dateTime).Format("2006-01-02 15:04:05")

	// Read event (shortstring)
	var eventBytes [129]byte
	binary.Read(r, binary.LittleEndian, &eventBytes)
	h.SEvent = DelphiShortStrToStr(eventBytes[:])

	var padding3 [3]byte
	binary.Read(r, binary.LittleEndian, &padding3)

	binary.Read(r, binary.LittleEndian, &h.GameId)
	binary.Read(r, binary.LittleEndian, &h.CompLevel1)
	binary.Read(r, binary.LittleEndian, &h.CompLevel2)

	var countForElo, addtoProfile1, addtoProfile2 byte
	binary.Read(r, binary.LittleEndian, &countForElo)
	binary.Read(r, binary.LittleEndian, &addtoProfile1)
	binary.Read(r, binary.LittleEndian, &addtoProfile2)
	h.CountForElo = countForElo != 0
	h.AddtoProfile1 = addtoProfile1 != 0
	h.AddtoProfile2 = addtoProfile2 != 0

	// Read location (shortstring)
	var locationBytes [129]byte
	binary.Read(r, binary.LittleEndian, &locationBytes)
	h.SLocation = DelphiShortStrToStr(locationBytes[:])

	binary.Read(r, binary.LittleEndian, &h.GameMode)

	var imported byte
	binary.Read(r, binary.LittleEndian, &imported)
	h.Imported = imported != 0

	// Read round (shortstring)
	var roundBytes [129]byte
	binary.Read(r, binary.LittleEndian, &roundBytes)
	h.SRound = DelphiShortStrToStr(roundBytes[:])

	var padding2 [2]byte
	binary.Read(r, binary.LittleEndian, &padding2)

	binary.Read(r, binary.LittleEndian, &h.Invert)
	binary.Read(r, binary.LittleEndian, &h.Version)
	binary.Read(r, binary.LittleEndian, &h.Magic)
	binary.Read(r, binary.LittleEndian, &h.MoneyInitG)
	binary.Read(r, binary.LittleEndian, &h.MoneyInitScore)

	var entered, counted, unratedImp byte
	binary.Read(r, binary.LittleEndian, &entered)
	binary.Read(r, binary.LittleEndian, &counted)
	binary.Read(r, binary.LittleEndian, &unratedImp)
	h.Entered = entered != 0
	h.Counted = counted != 0
	h.UnratedImp = unratedImp != 0

	var padding4 byte
	binary.Read(r, binary.LittleEndian, &padding4)

	binary.Read(r, binary.LittleEndian, &h.CommentHeaderMatch)
	binary.Read(r, binary.LittleEndian, &h.CommentFooterMatch)

	var isMoneyMatch byte
	binary.Read(r, binary.LittleEndian, &isMoneyMatch)
	h.IsMoneyMatch = isMoneyMatch != 0

	var padding5 [3]byte
	binary.Read(r, binary.LittleEndian, &padding5)

	binary.Read(r, binary.LittleEndian, &h.WinMoney)
	binary.Read(r, binary.LittleEndian, &h.LoseMoney)
	binary.Read(r, binary.LittleEndian, &h.Currency)
	binary.Read(r, binary.LittleEndian, &h.FeeMoney)
	binary.Read(r, binary.LittleEndian, &h.TableStake)
	binary.Read(r, binary.LittleEndian, &h.SiteId)

	// Version-specific fields
	if h.Version >= 8 {
		binary.Read(r, binary.LittleEndian, &h.CubeLimit)
		binary.Read(r, binary.LittleEndian, &h.AutoDoubleMax)
	}

	if h.Version >= 24 {
		var transcribed byte
		binary.Read(r, binary.LittleEndian, &transcribed)
		h.Transcribed = transcribed != 0

		var padding6 byte
		binary.Read(r, binary.LittleEndian, &padding6)

		event, _ := ReadUTF16Array(r, 129)
		h.Event = UTF16IntArrayToString(event)

		player1, _ := ReadUTF16Array(r, 129)
		h.Player1 = UTF16IntArrayToString(player1)

		player2, _ := ReadUTF16Array(r, 129)
		h.Player2 = UTF16IntArrayToString(player2)

		location, _ := ReadUTF16Array(r, 129)
		h.Location = UTF16IntArrayToString(location)

		round, _ := ReadUTF16Array(r, 129)
		h.Round = UTF16IntArrayToString(round)
	}

	if h.Version >= 25 {
		h.TimeSetting = &TimeSettingRecord{}
		h.TimeSetting.FromStream(r)
	}

	if h.Version >= 26 {
		binary.Read(r, binary.LittleEndian, &h.TotTimeDelayMove)
		binary.Read(r, binary.LittleEndian, &h.TotTimeDelayCube)
		binary.Read(r, binary.LittleEndian, &h.TotTimeDelayMoveDone)
		binary.Read(r, binary.LittleEndian, &h.TotTimeDelayCubeDone)
	}

	if h.Version >= 30 {
		transcriber, _ := ReadUTF16Array(r, 129)
		h.Transcriber = UTF16IntArrayToString(transcriber)
	}

	return nil
}

// HeaderGameEntry represents game header
type HeaderGameEntry struct {
	Name                string
	EntryType           int
	Version             int32
	Score1              int32
	Score2              int32
	CrawfordApply       bool
	PosInit             [26]int8
	GameNumber          int32
	InProgress          bool
	CommentHeaderGame   int32
	CommentFooterGame   int32
	NumberOfAutoDoubles int32
}

// FromStream reads HeaderGameEntry from stream
func (h *HeaderGameEntry) FromStream(r io.Reader, version int32) error {
	h.Name = "GameHeader"
	h.EntryType = 1 // ENTRYTYPE_HEADERGAME
	h.Version = version

	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	// Python format: '<9xxxxllB26bxlBxxxlll'
	// After 9 bytes, we have 4x (3 bytes padding here), then ll (Score1, Score2)
	var padding1 [3]byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &h.Score1)
	binary.Read(r, binary.LittleEndian, &h.Score2)

	var crawfordApply byte
	binary.Read(r, binary.LittleEndian, &crawfordApply)
	h.CrawfordApply = crawfordApply != 0

	binary.Read(r, binary.LittleEndian, &h.PosInit)

	var padding2 byte
	binary.Read(r, binary.LittleEndian, &padding2)

	binary.Read(r, binary.LittleEndian, &h.GameNumber)

	var inProgress byte
	binary.Read(r, binary.LittleEndian, &inProgress)
	h.InProgress = inProgress != 0

	var padding3 [3]byte
	binary.Read(r, binary.LittleEndian, &padding3)

	binary.Read(r, binary.LittleEndian, &h.CommentHeaderGame)
	binary.Read(r, binary.LittleEndian, &h.CommentFooterGame)

	if version >= 26 {
		binary.Read(r, binary.LittleEndian, &h.NumberOfAutoDoubles)
	}

	return nil
}

// CubeEntry represents cube action
type CubeEntry struct {
	Name                   string
	EntryType              int
	Version                int32
	ActiveP                int32
	Double                 int32
	Take                   int32
	BeaverR                int32
	RaccoonR               int32
	CubeB                  int32
	Position               [26]int8
	Doubled                *EngineStructDoubleAction
	ErrCube                float64
	DiceRolled             string
	ErrTake                float64
	RolloutIndexD          int32
	CompChoiceD            int32
	AnalyzeC               int32
	ErrBeaver              float64
	ErrRaccoon             float64
	AnalyzeCR              int32
	IsValid                int32
	TutorCube              int8
	TutorTake              int8
	ErrTutorCube           float64
	ErrTutorTake           float64
	FlaggedDouble          bool
	CommentCube            int32
	EditedCube             bool
	TimeDelayCube          bool
	TimeDelayCubeDone      bool
	NumberOfAutoDoubleCube int32
	TimeBot                int32
	TimeTop                int32
}

// FromStream reads CubeEntry from stream
func (c *CubeEntry) FromStream(r io.Reader, version int32) error {
	c.Name = "Cube"
	c.EntryType = 2 // ENTRYTYPE_CUBE
	c.Version = version

	// Python format: '<9xxxxllllll26bxx'
	// 9x + xxx = 9 + 3 = 12 bytes skip (NOT 13!)
	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	var padding1 [3]byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &c.ActiveP)
	binary.Read(r, binary.LittleEndian, &c.Double)
	binary.Read(r, binary.LittleEndian, &c.Take)
	binary.Read(r, binary.LittleEndian, &c.BeaverR)
	binary.Read(r, binary.LittleEndian, &c.RaccoonR)
	binary.Read(r, binary.LittleEndian, &c.CubeB)
	binary.Read(r, binary.LittleEndian, &c.Position)

	var padding2 [2]byte
	binary.Read(r, binary.LittleEndian, &padding2)

	c.Doubled = &EngineStructDoubleAction{}
	c.Doubled.FromStream(r)

	// Python format: '<xxxxd3BxxxxxdlllxxxxddllbbxxxxxxddBxxxlBBBxlll'
	// xxxx = 4 bytes padding
	var padding3 [4]byte
	binary.Read(r, binary.LittleEndian, &padding3)

	// d = 8 bytes (double) → ErrCube
	binary.Read(r, binary.LittleEndian, &c.ErrCube)

	// 3B = 3 bytes → DiceRolled
	var diceBytes [3]uint8
	binary.Read(r, binary.LittleEndian, &diceBytes)
	c.DiceRolled = DelphiShortStrToStr(diceBytes[:])

	// xxxxx = 5 bytes padding
	var padding4 [5]byte
	binary.Read(r, binary.LittleEndian, &padding4)

	// d = 8 bytes → ErrTake
	binary.Read(r, binary.LittleEndian, &c.ErrTake)

	// lll = 12 bytes (3 int32s) → RolloutIndexD, CompChoiceD, AnalyzeC
	binary.Read(r, binary.LittleEndian, &c.RolloutIndexD)
	binary.Read(r, binary.LittleEndian, &c.CompChoiceD)
	binary.Read(r, binary.LittleEndian, &c.AnalyzeC)

	// xxxx = 4 bytes padding
	var padding5 [4]byte
	binary.Read(r, binary.LittleEndian, &padding5)

	// dd = 16 bytes → ErrBeaver, ErrRaccoon
	binary.Read(r, binary.LittleEndian, &c.ErrBeaver)
	binary.Read(r, binary.LittleEndian, &c.ErrRaccoon)

	// ll = 8 bytes → AnalyzeCR, IsValid
	binary.Read(r, binary.LittleEndian, &c.AnalyzeCR)
	binary.Read(r, binary.LittleEndian, &c.IsValid)

	// bb = 2 bytes → TutorCube, TutorTake
	binary.Read(r, binary.LittleEndian, &c.TutorCube)
	binary.Read(r, binary.LittleEndian, &c.TutorTake)

	// xxxxxx = 6 bytes padding
	var padding6 [6]byte
	binary.Read(r, binary.LittleEndian, &padding6)

	// dd = 16 bytes → ErrTutorCube, ErrTutorTake
	binary.Read(r, binary.LittleEndian, &c.ErrTutorCube)
	binary.Read(r, binary.LittleEndian, &c.ErrTutorTake)

	// B = 1 byte → FlaggedDouble
	var flaggedDouble uint8
	binary.Read(r, binary.LittleEndian, &flaggedDouble)
	c.FlaggedDouble = flaggedDouble != 0

	// xxx = 3 bytes padding
	var padding7 [3]byte
	binary.Read(r, binary.LittleEndian, &padding7)

	// l = 4 bytes → CommentCube
	binary.Read(r, binary.LittleEndian, &c.CommentCube)

	if version >= 24 {
		var editedCube byte
		binary.Read(r, binary.LittleEndian, &editedCube)
		c.EditedCube = editedCube != 0
	}

	if version >= 26 {
		var timeDelayCube, timeDelayCubeDone byte
		binary.Read(r, binary.LittleEndian, &timeDelayCube)
		binary.Read(r, binary.LittleEndian, &timeDelayCubeDone)
		c.TimeDelayCube = timeDelayCube != 0
		c.TimeDelayCubeDone = timeDelayCubeDone != 0
	}

	if version >= 27 {
		var padding9 byte
		binary.Read(r, binary.LittleEndian, &padding9)
		binary.Read(r, binary.LittleEndian, &c.NumberOfAutoDoubleCube)
	}

	if version >= 28 {
		binary.Read(r, binary.LittleEndian, &c.TimeBot)
		binary.Read(r, binary.LittleEndian, &c.TimeTop)
	}

	return nil
}

// MoveEntry represents a move
type MoveEntry struct {
	Name                   string
	EntryType              int
	Version                int32
	PositionI              [26]int8
	PositionEnd            [26]int8
	ActiveP                int32
	Moves                  [8]int32
	Dice                   [2]int32
	CubeA                  int32
	ErrorM                 float64
	NMoveEval              int32
	DataMoves              *EngineStructBestMoveRecord
	Played                 bool
	ErrMove                float64
	ErrLuck                float64
	CompChoice             int32
	InitEq                 float64
	RolloutIndexM          [32]int32
	AnalyzeM               int32
	AnalyzeL               int32
	InvalidM               int32
	PositionTutor          [26]int8
	Tutor                  int8
	ErrTutorMove           float64
	Flagged                bool
	CommentMove            int32
	EditedMove             bool
	TimeDelayMove          uint32
	TimeDelayMoveDone      uint32
	NumberOfAutoDoubleMove int32
}

// FromStream reads MoveEntry from stream
func (m *MoveEntry) FromStream(r io.Reader, version int32) error {
	m.Name = "Move"
	m.EntryType = 3 // ENTRYTYPE_MOVE
	m.Version = version

	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	binary.Read(r, binary.LittleEndian, &m.PositionI)
	binary.Read(r, binary.LittleEndian, &m.PositionEnd)

	var padding1 [3]byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &m.ActiveP)
	binary.Read(r, binary.LittleEndian, &m.Moves)
	binary.Read(r, binary.LittleEndian, &m.Dice)
	binary.Read(r, binary.LittleEndian, &m.CubeA)
	binary.Read(r, binary.LittleEndian, &m.ErrorM)
	binary.Read(r, binary.LittleEndian, &m.NMoveEval)

	m.DataMoves = &EngineStructBestMoveRecord{}
	m.DataMoves.FromStream(r)

	var played byte
	binary.Read(r, binary.LittleEndian, &played)
	m.Played = played != 0

	var padding2 [3]byte
	binary.Read(r, binary.LittleEndian, &padding2)

	binary.Read(r, binary.LittleEndian, &m.ErrMove)
	binary.Read(r, binary.LittleEndian, &m.ErrLuck)
	binary.Read(r, binary.LittleEndian, &m.CompChoice)

	var padding3 [4]byte
	binary.Read(r, binary.LittleEndian, &padding3)

	binary.Read(r, binary.LittleEndian, &m.InitEq)
	binary.Read(r, binary.LittleEndian, &m.RolloutIndexM)
	binary.Read(r, binary.LittleEndian, &m.AnalyzeM)
	binary.Read(r, binary.LittleEndian, &m.AnalyzeL)
	binary.Read(r, binary.LittleEndian, &m.InvalidM)
	binary.Read(r, binary.LittleEndian, &m.PositionTutor)
	binary.Read(r, binary.LittleEndian, &m.Tutor)

	var padding4 byte
	binary.Read(r, binary.LittleEndian, &padding4)

	binary.Read(r, binary.LittleEndian, &m.ErrTutorMove)

	var flagged byte
	binary.Read(r, binary.LittleEndian, &flagged)
	m.Flagged = flagged != 0

	var padding5 [3]byte
	binary.Read(r, binary.LittleEndian, &padding5)

	binary.Read(r, binary.LittleEndian, &m.CommentMove)

	if version >= 24 {
		var editedMove byte
		binary.Read(r, binary.LittleEndian, &editedMove)
		m.EditedMove = editedMove != 0
	}

	if version >= 26 {
		var padding6 [3]byte
		binary.Read(r, binary.LittleEndian, &padding6)
		binary.Read(r, binary.LittleEndian, &m.TimeDelayMove)
		binary.Read(r, binary.LittleEndian, &m.TimeDelayMoveDone)
	}

	if version >= 27 {
		binary.Read(r, binary.LittleEndian, &m.NumberOfAutoDoubleMove)
	}

	return nil
}

// FooterGameEntry represents game footer
type FooterGameEntry struct {
	Name           string
	EntryType      int
	Version        int32
	Score1g        int32
	Score2g        int32
	CrawfordApplyg bool
	Winner         int32
	PointsWon      int32
	Termination    int32
	ErrResign      float64
	ErrTakeResign  float64
	Eval           [7]float64
	EvalLevel      int32
}

// FromStream reads FooterGameEntry from stream
func (f *FooterGameEntry) FromStream(r io.Reader, version int32) error {
	f.Name = "GameFooter"
	f.EntryType = 4 // ENTRYTYPE_FOOTERGAME
	f.Version = version

	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	var padding1 [4]byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &f.Score1g)
	binary.Read(r, binary.LittleEndian, &f.Score2g)

	var crawfordApplyg byte
	binary.Read(r, binary.LittleEndian, &crawfordApplyg)
	f.CrawfordApplyg = crawfordApplyg != 0

	var padding2 [3]byte
	binary.Read(r, binary.LittleEndian, &padding2)

	binary.Read(r, binary.LittleEndian, &f.Winner)
	binary.Read(r, binary.LittleEndian, &f.PointsWon)
	binary.Read(r, binary.LittleEndian, &f.Termination)

	var padding3 [4]byte
	binary.Read(r, binary.LittleEndian, &padding3)

	binary.Read(r, binary.LittleEndian, &f.ErrResign)
	binary.Read(r, binary.LittleEndian, &f.ErrTakeResign)
	binary.Read(r, binary.LittleEndian, &f.Eval)
	binary.Read(r, binary.LittleEndian, &f.EvalLevel)

	return nil
}

// FooterMatchEntry represents match footer
type FooterMatchEntry struct {
	Name      string
	EntryType int
	Version   int32
	Score1m   int32
	Score2m   int32
	WinnerM   int32
	Elo1m     float64
	Elo2m     float64
	Exp1m     int32
	Exp2m     int32
	Datem     string
}

// FromStream reads FooterMatchEntry from stream
func (f *FooterMatchEntry) FromStream(r io.Reader, version int32) error {
	f.Name = "MatchFooter"
	f.EntryType = 5 // ENTRYTYPE_FOOTERMATCH
	f.Version = version

	var skip [9]byte
	binary.Read(r, binary.LittleEndian, &skip)

	var padding1 [4]byte
	binary.Read(r, binary.LittleEndian, &padding1)

	binary.Read(r, binary.LittleEndian, &f.Score1m)
	binary.Read(r, binary.LittleEndian, &f.Score2m)
	binary.Read(r, binary.LittleEndian, &f.WinnerM)
	binary.Read(r, binary.LittleEndian, &f.Elo1m)
	binary.Read(r, binary.LittleEndian, &f.Elo2m)
	binary.Read(r, binary.LittleEndian, &f.Exp1m)
	binary.Read(r, binary.LittleEndian, &f.Exp2m)

	var dateTime float64
	binary.Read(r, binary.LittleEndian, &dateTime)
	f.Datem = DelphiDateTimeConv(dateTime).Format("2006-01-02 15:04:05")

	return nil
}

// GameFileRecord represents a record in the game file
type GameFileRecord struct {
	EntryType int
	Version   int32
	Record    interface{}
}

// FromStream reads GameFileRecord from stream
func (g *GameFileRecord) FromStream(r io.Reader, version int32) error {
	startPos, _ := r.(*bytes.Reader).Seek(0, io.SeekCurrent)

	var header struct {
		Skip      [8]byte
		EntryType byte
	}

	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	g.EntryType = int(header.EntryType)
	g.Version = version

	// Seek back to start
	r.(*bytes.Reader).Seek(startPos, io.SeekStart)

	// Read appropriate record type
	switch g.EntryType {
	case 0: // ENTRYTYPE_HEADERMATCH
		rec := &HeaderMatchEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	case 1: // ENTRYTYPE_HEADERGAME
		rec := &HeaderGameEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	case 2: // ENTRYTYPE_CUBE
		rec := &CubeEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	case 3: // ENTRYTYPE_MOVE
		rec := &MoveEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	case 4: // ENTRYTYPE_FOOTERGAME
		rec := &FooterGameEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	case 5: // ENTRYTYPE_FOOTERMATCH
		rec := &FooterMatchEntry{}
		err = rec.FromStream(r, version)
		g.Record = rec
	default:
		// Unimplemented entry type - skip it
		g.Record = nil
	}

	if err != nil {
		return err
	}

	// Each record is 2560 bytes, advance to next
	realRecSize, _ := r.(*bytes.Reader).Seek(0, io.SeekCurrent)
	realRecSize -= startPos
	r.(*bytes.Reader).Seek(2560-realRecSize, io.SeekCurrent)

	return nil
}
