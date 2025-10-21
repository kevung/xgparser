# XGID Position File Parser

This module provides functionality to parse Xtreme Gammon XGID position files in multiple languages (English, French, German, Spanish, Italian, Japanese, Russian, Greek, Finnish, etc.).

## Features

- **Language-independent parsing**: Automatically handles position files in any language
- **Complete position analysis**: Extracts XGID, player names, scores, cube information, and move analysis
- **Integration with xglight structures**: Converts parsed positions to `CheckerMove` structures
- **Multi-language support**: Works with English, French, German, and other language variants

## File Format

XGID position files contain:
1. XGID string (position encoding)
2. Player information
3. Match/game context (score, cube, match length)
4. Board diagram (ASCII art)
5. Move analysis with equity and statistics

Example file (English):
```
XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10

X:postmanpat   O:marcow777
Score is X:2 O:3 13 pt.(s) match.
 +13-14-15-16-17-18------19-20-21-22-23-24-+
 | X        O  O    |   | O  O  O  O  X  O |
 ...
 +12-11-10--9--8--7-------6--5--4--3--2--1-+
Cube: 1
X to play 22

    1. 3-ply       Bar/23(2) 13/11(2)           eq:-1.000
      Player:   35.03% (G:4.51% B:0.12%)
      Opponent: 64.97% (G:39.70% B:2.25%)
...
```

## Usage

### Basic Parsing

```go
import "github.com/kevung/xgparser/xgparser"

// Parse a single XGID file
pos, err := xgparser.ParseXGIDFile("path/to/xgid/file.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("XGID: %s\n", pos.XGID)
fmt.Printf("Players: %s vs %s\n", pos.Player1Name, pos.Player2Name)
fmt.Printf("Analysis: %d moves\n", len(pos.Analysis))
```

### Parse from io.Reader

```go
file, _ := os.Open("position.txt")
defer file.Close()

pos, err := xgparser.ParseXGIDFromReader(file)
```

### Convert to CheckerMove Structure

```go
pos, _ := xgparser.ParseXGIDFile("position.txt")

// Convert to xglight CheckerMove structure
checkerMove := pos.ToCheckerMove()

fmt.Printf("Active Player: %d\n", checkerMove.ActivePlayer)
fmt.Printf("Dice: [%d, %d]\n", checkerMove.Dice[0], checkerMove.Dice[1])
fmt.Printf("Analysis entries: %d\n", len(checkerMove.Analysis))
```

### Parse XGID String Components

```go
// Parse just the XGID string
components, err := xgparser.ParseXGID("----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10")

fmt.Printf("Position ID: %s\n", components.PositionID)
fmt.Printf("Cube Owner: %d\n", components.CubeOwner)
fmt.Printf("Score: %d-%d\n", components.ScoreX, components.ScoreO)
```

### Convert XGID Position to Checker Array

```go
// Convert XGID position encoding to internal checker array format
position := xgparser.XGIDToPosition("----BaC-B---aD--aa-bcbbBbB")
// position is [26]int8 with checkers for each point
```

## Data Structures

### XGIDPosition

Main structure containing parsed position data:

```go
type XGIDPosition struct {
    XGID            string            // Full XGID string
    XGIDComponents  XGIDComponents    // Parsed XGID components
    Player1Name     string            // X player name
    Player2Name     string            // O player name
    Score           [2]int32          // [X, O] scores
    MatchLength     int32             // Match length
    Cube            int32             // Cube value
    PlayerToMove    int32             // 1=X, -1=O
    Dice            [2]int32          // Dice rolled
    Analysis        []XGIDMoveAnalysis // Move analysis
    XGVersion       string            // XG version
    MET             string            // Match equity table
    BoardDiagram    string            // ASCII board
}
```

### XGIDMoveAnalysis

Analysis for individual moves:

```go
type XGIDMoveAnalysis struct {
    Rank              int     // Move ranking (1=best)
    Ply               int     // Analysis depth
    IsBook            bool    // Book move flag
    Move              string  // Move notation
    Equity            float64 // Equity
    EquityDiff        float64 // Difference from best
    PlayerWinRate     float64 // Win rate (0.0-1.0)
    PlayerGammonRate  float64 // Gammon rate
    PlayerBgRate      float64 // Backgammon rate
    OpponentGammonRate float64
    OpponentBgRate    float64
}
```

## Command-Line Tool

A sample program `xgid_parser` demonstrates the parsing functionality:

```bash
# Build the tool
go build -o xgid_parser cmd/xgid_parser/main.go

# Parse a single file
./xgid_parser tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt

# Parse all files in a directory
./xgid_parser tmp/xgid/en

# Output JSON
JSON_OUTPUT=1 ./xgid_parser position.txt
```

## Supported Languages

The parser automatically handles these language variants:

- **English**: "Score is", "Cube:", "to play", "Player:", "Opponent:", "ply"
- **French**: "Le score est", "Videau:", "à jouer", "Joueur:", "Adversaire:", "Livre", "plis"
- **German**: "Spielstand ist", "Doppler:", "zu spielen", "Spieler:", "Gegner:", "Buch", "Züge"
- **Spanish, Italian, Japanese, Russian, Greek, Finnish**: Similar patterns

The parser uses regex patterns that match multiple language variants, making it robust across different XG localizations.

## XGID Format Reference

XGID string format: `XGID=position:cubeOwner:cubeValue:playerToMove:dice:scoreX:scoreO:crawford:matchLength:maxCube`

- **position**: 26-character encoding (points 24-1, X bar, O bar)
  - `-` = empty
  - `A-Z` = 1-26 checkers for X (positive)
  - `a-o` = 1-15 checkers for O (negative)
- **cubeOwner**: 0=centered, 1=X owns, -1=O owns
- **cubeValue**: Power of 2 (0=1, 1=2, 2=4, 3=8, etc.)
- **playerToMove**: 1=X, -1=O
- **dice**: Two digits (e.g., "22", "51", "00"=no dice)
- **scoreX, scoreO**: Current score
- **crawford**: 0=no, 1=yes
- **matchLength**: Points in match (0=money game)
- **maxCube**: Maximum cube value (usually 10)

## Integration with XGLight

The `ToCheckerMove()` method converts an `XGIDPosition` to the standard `CheckerMove` structure used in the xglight module, allowing seamless integration with existing XG file parsing and analysis code.

```go
pos, _ := xgparser.ParseXGIDFile("position.txt")
move := pos.ToCheckerMove()

// Now use with existing xglight functions
match := &xgparser.Match{
    Metadata: xgparser.MatchMetadata{...},
    Games: []xgparser.Game{{
        Moves: []xgparser.Move{{
            MoveType: "checker",
            CheckerMove: move,
        }},
    }},
}
```

## License

LGPL 2.1 - See LICENSE file for details.
