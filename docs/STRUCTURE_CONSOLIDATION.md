# Structure Consolidation Summary

## Overview

The XG parser codebase has been refactored to use unified data structures for both XG binary file parsing and XGID position text file parsing. This consolidation minimizes duplication and provides a consistent interface for database schema design.

## Unified Core Structures

### 1. `MatchMetadata` (in `xglight.go`)
Used by both XG binary and XGID parsers to store match/game metadata.

```go
type MatchMetadata struct {
    Player1Name    string // X player name
    Player2Name    string // O player name  
    Location       string // Tournament/match location
    Event          string // Event name
    Round          string // Round identifier
    DateTime       string // Date and time
    MatchLength    int32  // Match length (points)
    EngineVersion  int32  // File format version (XG binary only)
    ProductVersion string // XG version (e.g., "eXtreme Gammon 2.19.1")
    MET            string // Match equity table (XGID only)
}
```

**Field Usage:**
- `EngineVersion`: Only populated from XG binary files (file format version)
- `MET`: Only populated from XGID text files (match equity table name)
- All other fields: Populated from both formats when available

### 2. `Position` (in `xglight.go`)
Board position state - used identically by both parsers.

```go
type Position struct {
    Checkers [26]int8 // Checker positions
    Cube     int32    // Cube value (1, 2, 4, 8, ...)
    CubePos  int32    // Cube owner (0=center, 1=player1, -1=player2)
    Score    [2]int32 // Match score [player1, player2]
}
```

### 3. `CheckerAnalysis` (in `xglight.go`)
Analysis for a single checker move alternative - used identically by both parsers.

```go
type CheckerAnalysis struct {
    Position          Position // Resulting position after move
    Move              [8]int8  // The move (1-24=points, 25=bar, -2=bear off, -1=unused)
    Player1WinRate    float32  // Win rate for player on roll
    Player1GammonRate float32  // Gammon rate for player on roll
    Player1BgRate     float32  // Backgammon rate for player on roll
    Player2GammonRate float32  // Gammon rate for opponent
    Player2BgRate     float32  // Backgammon rate for opponent
    Equity            float32  // Normalized equity
    AnalysisDepth     int16    // Evaluation level (ply depth, 0=book)
}
```

### 4. `CheckerMove` (in `xglight.go`)
Complete checker move decision - primary return type for both parsers.

```go
type CheckerMove struct {
    Position     Position          // Position before the move
    ActivePlayer int32             // Player making the move (1 or -1)
    Dice         [2]int32          // Dice rolled
    PlayedMove   [8]int32          // The actual move played (XG binary only)
    Analysis     []CheckerAnalysis // Analysis of possible moves
}
```

**Note:** `PlayedMove` is only populated from XG binary files. XGID text files don't specify which move was actually played.

## API Changes

### XGID Parser Functions

**Before:**
```go
func ParseXGIDFile(filename string) (*CheckerMove, *XGIDPositionInfo, error)
func ParseXGIDFromReader(r io.Reader) (*CheckerMove, *XGIDPositionInfo, error)
```

**After:**
```go
func ParseXGIDFile(filename string) (*CheckerMove, *MatchMetadata, error)
func ParseXGIDFromReader(r io.Reader) (*CheckerMove, *MatchMetadata, error)
```

The `XGIDPositionInfo` struct has been removed. All relevant metadata is now in `MatchMetadata`.

### Removed Structures

1. **`XGIDPositionInfo`** - Replaced by `MatchMetadata`
2. **`XGIDMoveNotation`** - Was not used; move notation strings are not stored in the unified structures

### Retained XGID-Specific Structures

**`XGIDComponents`** - Still available for low-level XGID parsing utilities:

```go
type XGIDComponents struct {
    PositionID   string // Position encoding
    CubeOwner    int32  // 0=centered, 1=X, -1=O
    CubeValue    int32  // Cube as power of 2 (0=1, 1=2, 2=4, etc.)
    PlayerToMove int32  // 1=X, -1=O
    Dice         string // Dice string (e.g., "22", "51")
    ScoreX       int32  // X's score
    ScoreO       int32  // O's score
    CrawfordFlag int32  // Crawford status
    MatchLength  int32  // Match length
    MaxCube      int32  // Maximum cube value
}
```

This is only used by the `ParseXGID()` helper function for parsing raw XGID strings.

## Database Schema Implications

The consolidated structures map naturally to database tables:

### `matches` Table
```sql
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    player1_name TEXT,
    player2_name TEXT,
    location TEXT,
    event TEXT,
    round TEXT,
    date_time TIMESTAMP,
    match_length INTEGER,
    engine_version INTEGER,     -- NULL for XGID sources
    product_version TEXT,
    met TEXT,                    -- NULL for XG binary sources
    source_type TEXT             -- 'xg_binary' or 'xgid_text'
);
```

### `positions` Table
```sql
CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    match_id INTEGER REFERENCES matches(id),
    game_number INTEGER,
    move_number INTEGER,
    checkers SMALLINT[26],       -- Array of checker positions
    cube INTEGER,
    cube_pos INTEGER,
    score INTEGER[2]
);
```

### `checker_moves` Table
```sql
CREATE TABLE checker_moves (
    id SERIAL PRIMARY KEY,
    position_id INTEGER REFERENCES positions(id),
    active_player INTEGER,
    dice INTEGER[2],
    played_move INTEGER[8]       -- NULL for XGID sources
);
```

### `move_analysis` Table
```sql
CREATE TABLE move_analysis (
    id SERIAL PRIMARY KEY,
    move_id INTEGER REFERENCES checker_moves(id),
    rank INTEGER,                -- Sequence within alternatives
    move INTEGER[8],
    player1_win_rate REAL,
    player1_gammon_rate REAL,
    player1_bg_rate REAL,
    player2_gammon_rate REAL,
    player2_bg_rate REAL,
    equity REAL,
    analysis_depth INTEGER
);
```

## Migration Guide

### For Code Using XGID Parser

**Before:**
```go
move, info, err := xgparser.ParseXGIDFile("position.txt")
if err != nil {
    // handle error
}
fmt.Printf("Players: %s vs %s\n", info.Player1Name, info.Player2Name)
fmt.Printf("Version: %s\n", info.XGVersion)
fmt.Printf("MET: %s\n", info.MET)
```

**After:**
```go
move, metadata, err := xgparser.ParseXGIDFile("position.txt")
if err != nil {
    // handle error
}
fmt.Printf("Players: %s vs %s\n", metadata.Player1Name, metadata.Player2Name)
fmt.Printf("Version: %s\n", metadata.ProductVersion)
fmt.Printf("MET: %s\n", metadata.MET)
```

## Benefits

1. **Unified Interface**: Both XG binary and XGID parsers return the same structure types
2. **Simpler Code**: Less duplication, easier to maintain
3. **Database-Friendly**: Structures map directly to table schemas
4. **Extensible**: Easy to add new parsers (e.g., GNU Backgammon format) using the same structures
5. **Type Safety**: Single source of truth for data structures

## Testing

All existing tests pass with no regressions:
- ✅ XGID parsing (English, French, German, etc.)
- ✅ XG binary file parsing
- ✅ Command-line tools updated and working

## Files Modified

1. `xgparser/xglight.go` - Added `MET` field to `MatchMetadata`
2. `xgparser/xgid.go` - Removed `XGIDPositionInfo`, updated function signatures
3. `xgparser/xgid_test.go` - Updated tests to use `MatchMetadata`
4. `cmd/xgid_parser/main.go` - Updated to use new API

## Backward Compatibility

This is a **breaking change** for code using the XGID parser. The function signatures have changed from returning `*XGIDPositionInfo` to `*MatchMetadata`. However, the migration is straightforward (see Migration Guide above).

XG binary file parsing is **fully backward compatible** - no changes to those APIs.
