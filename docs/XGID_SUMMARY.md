# XGID Position Parser - Summary

## Overview

I've implemented a comprehensive XGID position file parser for your xgparser project. This parser can read Xtreme Gammon position files in multiple languages (English, French, German, Spanish, Italian, etc.) and convert them into your existing xglight data structures.

## What Was Created

### 1. Core Parser Module (`xgparser/xgid.go`)

**Key Functions:**
- `ParseXGIDFile(filename)` - Parse a single XGID file
- `ParseXGIDFromReader(io.Reader)` - Parse from any reader (files, network, memory)
- `ParseXGID(xgid)` - Parse XGID string components
- `XGIDToPosition(positionID)` - Convert XGID position encoding to checker array
- `(*XGIDPosition).ToCheckerMove()` - Convert to xglight CheckerMove structure

**Data Structures:**
- `XGIDPosition` - Complete parsed position with analysis
- `XGIDComponents` - Parsed XGID string components
- `XGIDMoveAnalysis` - Individual move analysis data

### 2. Command-Line Tools

**xgid_parser** (`cmd/xgid_parser/main.go`)
- Parse and display individual XGID files
- Parse entire directories of XGID files
- Show analysis with equity, win rates, gammon rates
- Convert to CheckerMove structure for integration

**batch_xgid** (`cmd/batch_xgid/main.go`)
- Batch process multiple XGID files
- Output results as JSON
- Useful for data extraction and analysis

### 3. Documentation

- `XGID_PARSER.md` - Complete documentation with examples
- Unit tests in `xgparser/xgid_test.go` - All tests passing

## Language Support

The parser automatically handles these languages:
- **English**: "Score is", "Cube:", "to play", "Player:", "Opponent:", "ply"
- **French**: "Le score est", "Videau:", "à jouer", "Joueur:", "Adversaire:", "Livre", "plis"
- **German**: "Spielstand ist", "Doppler:", "zu spielen", "Spieler:", "Gegner:", "Buch", "Züge"
- **Others**: Spanish, Italian, Japanese, Russian, Greek, Finnish

## Usage Examples

### Basic File Parsing

```go
import "github.com/kevung/xgparser/xgparser"

pos, err := xgparser.ParseXGIDFile("tmp/xgid/en/position.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("XGID: %s\n", pos.XGID)
fmt.Printf("Players: %s vs %s\n", pos.Player1Name, pos.Player2Name)
fmt.Printf("Score: %d-%d (Match to %d)\n", 
    pos.Score[0], pos.Score[1], pos.MatchLength)

// Access analysis
for _, move := range pos.Analysis {
    fmt.Printf("%d. %s  eq: %.3f\n", 
        move.Rank, move.Move, move.Equity)
}
```

### Convert to xglight Structure

```go
pos, _ := xgparser.ParseXGIDFile("position.txt")
checkerMove := pos.ToCheckerMove()

// Now you can use checkerMove with your existing xglight code
fmt.Printf("Active Player: %d\n", checkerMove.ActivePlayer)
fmt.Printf("Dice: [%d, %d]\n", checkerMove.Dice[0], checkerMove.Dice[1])
fmt.Printf("Analysis entries: %d\n", len(checkerMove.Analysis))

// Access evaluation data
if len(checkerMove.Analysis) > 0 {
    analysis := checkerMove.Analysis[0]
    fmt.Printf("Win rate: %.2f%%\n", analysis.Player1WinRate * 100)
    fmt.Printf("Equity: %.3f\n", analysis.Equity)
}
```

### Command-Line Usage

```bash
# Build the tools
go build -o xgid_parser cmd/xgid_parser/main.go
go build -o batch_xgid cmd/batch_xgid/main.go

# Parse a single file (English)
./xgid_parser tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt

# Parse a single file (French)
./xgid_parser tmp/xgid/fr/XGID=-b----E-C---eE---c-e----B-:0:0:-1:5.txt

# Parse all files in a directory
./xgid_parser tmp/xgid/en

# Batch process and output JSON
./batch_xgid tmp/xgid > positions.json
```

## Test Results

All tests pass successfully:
```
=== RUN   TestParseXGID
=== RUN   TestParseXGIDFromReader_English
=== RUN   TestParseXGIDFromReader_French
=== RUN   TestParseXGIDFromReader_German
=== RUN   TestXGIDToPosition
=== RUN   TestToCheckerMove
PASS
ok      github.com/kevung/xgparser/xgparser     0.003s
```

## Example Output

```
=== Parsing: tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt ===
XGID: ----BaC-B---aD--aa-bcbbBbB:0:0:1:22:2:3:0:13:10
Players: postmanpat (X) vs marcow777 (O)
Score: 2-3 (Match to 13)
Cube: 1
To play: X rolls 22
XG Version: 2.19.211.pre-release, MET: Kazaross XG2

Analysis (5 moves):
  1. [3-ply] Bar/23(2) 13/11(2)  eq: -1.000
     Player: 35.03% (G: 4.51%, BG: 0.12%)
     Opponent: 64.97% (G: 39.70%, BG: 2.25%)
  2. [3-ply] Bar/23(2) 8/6(2)  eq: -1.135 (-0.135)
     Player: 24.87% (G: 1.95%, BG: 0.05%)
     Opponent: 75.13% (G: 45.87%, BG: 2.23%)
  ...

=== Converted to CheckerMove Structure ===
Active Player: 1
Dice: [2, 2]
Position.Cube: 1
Position.CubePos: 0
Position.Score: [2, 3]
Analysis entries: 5
```

## Integration with Existing Code

The `ToCheckerMove()` method seamlessly converts XGID positions into your existing `CheckerMove` structure, so you can:

1. Parse XGID position files
2. Convert to CheckerMove
3. Use with existing xglight analysis code
4. Add to Match/Game structures
5. Output as JSON

This allows you to integrate position data from various sources (books, forums, training materials) that use the XGID format.

## Files Created

- `/home/unger/src/xgparser/xgparser/xgid.go` - Core parser
- `/home/unger/src/xgparser/xgparser/xgid_test.go` - Unit tests
- `/home/unger/src/xgparser/cmd/xgid_parser/main.go` - CLI tool
- `/home/unger/src/xgparser/cmd/batch_xgid/main.go` - Batch processor
- `/home/unger/src/xgparser/XGID_PARSER.md` - Documentation
- `/home/unger/src/xgparser/xgid_parser` - Built executable
- `/home/unger/src/xgparser/batch_xgid` - Built executable

## Next Steps

You can now:
1. Parse XGID files from any language
2. Extract position and analysis data
3. Convert to your xglight structures
4. Integrate with your existing XG file parsing
5. Build tools for position analysis and comparison
