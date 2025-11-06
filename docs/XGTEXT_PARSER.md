# XG Text Position Parser

The XG text position parser extracts analysis data from eXtremeGammon text-format position exports.

## What It Parses

The parser focuses on information **not available in the XGID**:

- **Player names** (X and O player names)
- **Move analysis** (equity evaluations, ply depth, move notation)
- **Cube analysis** (winning chances, cubeful/cubeless equities)
- **Metadata** (XG version, Match Equity Table)
- **Action type** (play move vs cube decision)

## What It Skips

Since the XGID contains complete position information, the parser **does not parse**:

- Board diagram (visual representation only)
- Pip count (computable from XGID position)
- Score (in XGID)
- Cube value and owner (in XGID)
- Dice (in XGID)
- Player to move (in XGID)

## Language Support

Supports text positions in multiple languages:
- **English** (EN)
- **French** (FR) 
- **German** (DE)
- **Japanese** (JP)

## Usage

### Command Line

```bash
# Parse text position
./xgtext position_EN.txt

# JSON output
./xgtext -format=json position_FR.txt

# Show only specific analysis
./xgtext -moves=false position_DE.txt  # Hide move analysis
./xgtext -cube=false position_JP.txt   # Hide cube analysis
```

### Web API

```bash
# Start web server
go run ./cmd/web_example

# Upload text position
curl -F "textfile=@position.txt" http://localhost:8080/text
```

### Programmatic Use

```go
package main

import (
    "os"
    "github.com/kevung/xgparser/xgparser"
)

func main() {
    file, _ := os.Open("position.txt")
    defer file.Close()
    
    pos, err := xgparser.ParseXGTextPosition(file)
    if err != nil {
        panic(err)
    }
    
    // Access parsed data
    fmt.Printf("XGID: %s\n", pos.XGID)
    fmt.Printf("Players: %s vs %s\n", pos.Player1Name, pos.Player2Name)
    fmt.Printf("Action: %s\n", pos.ActionType)
    
    // Move analysis
    for _, move := range pos.Analysis {
        fmt.Printf("%d. %s (eq: %.3f)\n", move.Rank, move.Move, move.Equity)
    }
    
    // Cube analysis
    if pos.CubeAnalysis != nil {
        fmt.Printf("Win: %.2f%%\n", pos.CubeAnalysis.PlayerWin)
    }
}
```

## Data Structure

```go
type XGTextPosition struct {
    XGID         string          // Complete XGID
    Player1Name  string          // X player name
    Player2Name  string          // O player name  
    ActionType   string          // "play" or "cube"
    Analysis     []XGMove        // Move evaluations
    CubeAnalysis *XGCubeAnalysis // Cube decision analysis
    Version      string          // XG version
    MET          string          // Match equity table name
}
```

## Integration

The text parser integrates with existing XG parser infrastructure:

- Uses `ParseXGID()` to decode position from XGID string
- Outputs JSON compatible with web/API interfaces
- Works alongside binary `.xg` file parser

## Examples

See `test/2025-11-04/` for sample positions in all supported languages.
