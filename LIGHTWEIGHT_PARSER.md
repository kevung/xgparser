# Lightweight XG Parser

A Go library for parsing eXtremeGammon (.xg) match files with a lightweight, database-friendly structure.

## Overview

This parser extracts essential match information from XG files, focusing on data needed for database storage and statistical analysis. It provides a simplified view compared to the full XG file format.

## Features

- **Flexible Input**: Parse from files, HTTP uploads, memory buffers, or any `io.ReadSeeker`
- **JSON Serializable**: All structures have JSON tags for easy export
- **Database Ready**: Designed for SQL storage with clean relational structure
- **Essential Data Only**: Omits rollouts, comments, and thumbnails for simplicity
- **Clean API**: Simple, readable structure names without unnecessary prefixes

## Installation

```bash
go get github.com/kevung/xgparser
```

## Quick Start

### Parse from File

```go
package main

import (
    "fmt"
    "log"
    "github.com/kevung/xgparser/xgparser"
)

func main() {
    // Parse XG file
    match, err := xgparser.ParseXGFromFile("match.xg")
    if err != nil {
        log.Fatal(err)
    }

    // Access match data
    fmt.Printf("%s vs %s\n", 
        match.Metadata.Player1Name,
        match.Metadata.Player2Name)
    fmt.Printf("Event: %s\n", match.Metadata.Event)
    fmt.Printf("Games: %d\n", len(match.Games))
    
    // Export to JSON
    jsonData, _ := match.ToJSON()
    fmt.Println(string(jsonData))
}
```

### Parse from HTTP Upload

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    file, _, _ := r.FormFile("xgfile")
    defer file.Close()
    
    // Read into memory
    data, _ := io.ReadAll(file)
    reader := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))
    
    // Parse
    match, err := xgparser.ParseXGFromReader(reader)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    jsonData, _ := match.ToJSON()
    w.Write(jsonData)
}
```

## API Reference

### Parsing Functions

#### ParseXGFromFile
```go
func ParseXGFromFile(filename string) (*Match, error)
```
Parse an XG file from disk. Recommended for simple file-based usage.

#### ParseXGFromReader
```go
func ParseXGFromReader(r io.ReadSeeker) (*Match, error)
```
Parse from any `io.ReadSeeker` source - HTTP uploads, memory buffers, network streams, etc.

#### ParseXG
```go
func ParseXG(segments []*Segment) (*Match, error)
```
Core parsing function. Use when you need custom segment extraction logic.

### Data Structures

#### Match
```go
type Match struct {
    Metadata MatchMetadata `json:"metadata"`
    Games    []Game        `json:"games"`
}
```
Root structure representing a complete match.

#### MatchMetadata
```go
type MatchMetadata struct {
    Player1Name    string `json:"player1_name"`
    Player2Name    string `json:"player2_name"`
    Location       string `json:"location"`
    Event          string `json:"event"`
    Round          string `json:"round"`
    DateTime       string `json:"date_time"`
    MatchLength    int32  `json:"match_length"`
    EngineVersion  int32  `json:"engine_version"`   // File format version (e.g., 30)
    ProductVersion string `json:"product_version"` // XG product version (e.g., "eXtreme Gammon 2.19.1")
}
```
The `EngineVersion` field indicates the XG file format version (typically 30 for recent versions).
The `ProductVersion` field contains the XG software version string if available in the file.

#### Game
```go
type Game struct {
    GameNumber   int32    `json:"game_number"`
    InitialScore [2]int32 `json:"initial_score"`
    Moves        []Move   `json:"moves"`
    Winner       int32    `json:"winner"`       // -1=player1, 1=player2
    PointsWon    int32    `json:"points_won"`
}
```

#### Move
```go
type Move struct {
    MoveType    string       `json:"move_type"` // "checker" or "cube"
    CheckerMove *CheckerMove `json:"checker_move,omitempty"`
    CubeMove    *CubeMove    `json:"cube_move,omitempty"`
}
```

#### CheckerMove
```go
type CheckerMove struct {
    Position     Position          `json:"position"`
    ActivePlayer int32             `json:"active_player"`
    Dice         [2]int32          `json:"dice"`
    PlayedMove   [8]int32          `json:"played_move"`
    Analysis     []CheckerAnalysis `json:"analysis"`
}
```

#### CheckerAnalysis
```go
type CheckerAnalysis struct {
    Position          Position `json:"position"`
    Move              [8]int8  `json:"move"`
    Player1WinRate    float32  `json:"player1_win_rate"`
    Player1GammonRate float32  `json:"player1_gammon_rate"`
    Player1BgRate     float32  `json:"player1_bg_rate"`
    Player2GammonRate float32  `json:"player2_gammon_rate"`
    Player2BgRate     float32  `json:"player2_bg_rate"`
    Equity            float32  `json:"equity"`
    AnalysisDepth     int16    `json:"analysis_depth"`
}
```

#### CubeMove
```go
type CubeMove struct {
    Position     Position      `json:"position"`
    ActivePlayer int32         `json:"active_player"`
    CubeAction   int32         `json:"cube_action"`
    Analysis     *CubeAnalysis `json:"analysis"`
}
```

#### CubeAnalysis
```go
type CubeAnalysis struct {
    Player1WinRate       float32 `json:"player1_win_rate"`
    Player1GammonRate    float32 `json:"player1_gammon_rate"`
    Player1BgRate        float32 `json:"player1_bg_rate"`
    Player2GammonRate    float32 `json:"player2_gammon_rate"`
    Player2BgRate        float32 `json:"player2_bg_rate"`
    CubelessNoDouble     float32 `json:"cubeless_no_double"`
    CubelessDouble       float32 `json:"cubeless_double"`
    CubefulNoDouble      float32 `json:"cubeful_no_double"`
    CubefulDoubleTake    float32 `json:"cubeful_double_take"`
    CubefulDoublePass    float32 `json:"cubeful_double_pass"`
    WrongPassTakePercent float32 `json:"wrong_pass_take_percent"`
    AnalysisDepth        int32   `json:"analysis_depth"`
}
```

#### Position
```go
type Position struct {
    Checkers [26]int8 `json:"checkers"` // Board position
    Cube     int32    `json:"cube"`     // Cube value
    CubePos  int32    `json:"cube_pos"` // Cube owner
    Score    [2]int32 `json:"score"`    // Match score
}
```

## Position Representation

Board positions use a 26-element array:
- `[0]`: Bar (negative values = player 2)
- `[1-24]`: Points (positive = player 1, negative = player 2)
- `[25]`: Borne off

Example starting position:
```
[0, -2, 0, 0, 0, 0, 5, 0, 3, 0, 0, 0, -5, 5, 0, 0, 0, -3, 0, -5, 0, 0, 0, 0, 2, 0]
```

## Move Representation

Checker moves are 8-element arrays containing from/to pairs:
- `[0,1]`: First checker move
- `[2,3]`: Second checker move  
- `[4,5]`: Third checker move
- `[6,7]`: Fourth checker move

Value `-1` indicates unused slots. Point 0 = bar, point 25 = bearing off.

Example: `[24, 22, 24, 22, -1, -1, -1, -1]` = two checkers from point 24 to 22

## Database Integration

### PostgreSQL Schema Example

```sql
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    player1_name VARCHAR(255),
    player2_name VARCHAR(255),
    location VARCHAR(255),
    event VARCHAR(255),
    round VARCHAR(255),
    date_time TIMESTAMP,
    match_length INTEGER,
    engine_version INTEGER
);

CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    match_id INTEGER REFERENCES matches(id),
    game_number INTEGER,
    initial_score_p1 INTEGER,
    initial_score_p2 INTEGER,
    winner INTEGER,
    points_won INTEGER
);

CREATE TABLE moves (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id),
    move_number INTEGER,
    move_type VARCHAR(10),
    active_player INTEGER,
    position JSONB,
    dice JSONB,
    played_move JSONB
);

CREATE TABLE checker_analysis (
    id SERIAL PRIMARY KEY,
    move_id INTEGER REFERENCES moves(id),
    rank INTEGER,
    move JSONB,
    player1_win_rate REAL,
    equity REAL,
    analysis_depth INTEGER
);
```

## Command-Line Tools

### xglight
Parse XG files to JSON:
```bash
go build -o xglight ./cmd/xglight/
./xglight match.xg > match.json
```

### stats_example
Extract match statistics:
```bash
go build -o stats_example ./cmd/stats_example/
./stats_example match.xg
```

### reader_example
Demonstrate all parsing methods:
```bash
go build -o reader_example ./cmd/reader_example/
./reader_example match.xg
```

### web_example
Web server with file upload:
```bash
go build -o web_example ./cmd/web_example/
./web_example
# Visit http://localhost:8080
```

## Common Usage Patterns

### Calculate Average Equity Loss
```go
totalLoss := float32(0.0)
count := 0
for _, game := range match.Games {
    for _, move := range game.Moves {
        if move.MoveType == "checker" && len(move.CheckerMove.Analysis) > 1 {
            bestEquity := move.CheckerMove.Analysis[0].Equity
            for _, a := range move.CheckerMove.Analysis {
                if a.Equity > bestEquity {
                    bestEquity = a.Equity
                }
            }
            totalLoss += bestEquity - move.CheckerMove.Analysis[0].Equity
            count++
        }
    }
}
avgLoss := totalLoss / float32(count)
```

### Count Move Types
```go
checkerMoves := 0
cubeMoves := 0
for _, game := range match.Games {
    for _, move := range game.Moves {
        if move.MoveType == "checker" {
            checkerMoves++
        } else {
            cubeMoves++
        }
    }
}
```

### Find Match Winner
```go
finalScore := [2]int32{0, 0}
for _, game := range match.Games {
    if game.Winner == -1 {
        finalScore[0] += game.PointsWon
    } else if game.Winner == 1 {
        finalScore[1] += game.PointsWon
    }
}
```

## What's Not Included

To keep the structure lightweight, these are omitted:
- Rollout data
- Comments and annotations
- Thumbnail images
- Detailed time control data
- ELO rating calculations
- Transcription metadata

These can be added in future versions if needed.

## Performance

- Parses typical 7-point matches in milliseconds
- JSON output is ~10-20% the size of full parser
- Suitable for batch processing large match collections
- No memory leaks, efficient allocation

## Testing

Run the test suite:
```bash
./test_xglight.sh
```

All tests validate:
- Build process
- JSON parsing and validation  
- Data structure correctness
- Multiple file parsing
- Data integrity
- Performance benchmarks

## License

This library is licensed under the **GNU Lesser General Public License v2.1 (LGPL-2.1)**, same as the original xgdatatools library.

### Credits

- **Original Python library**: Michael Petch (Copyright © 2013-2014)
  - GitHub: https://github.com/oysteijo/xgdatatools
- **Go port**: Kevin Unger (Copyright © 2025)
- **Lightweight parser**: Kevin Unger (Copyright © 2025)

## Contributing

This is part of the xgparser project. See the main README for contribution guidelines.
