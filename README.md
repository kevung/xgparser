# XGParser - Go Implementation

This is a Go implementation of the Python xgdatatools package for parsing ExtremeGammon (.xg) files.

## Current Status - SIGNIFICANT PROGRESS!

The package successfully implements:

✅ **xgutils.go** - Utility functions for:
- CRC32 calculation on streams
- UTF-16 string conversion
- Delphi datetime conversion
- Delphi shortstring parsing

✅ **xgzarc.go** - ZLib archive handling:
- Archive record parsing
- File record extraction
- Decompression of archive segments
- CRC verification

✅ **xgstruct.go** - Data structures for all XG record types:
- GameDataFormatHdrRecord
- TimeSettingRecord
- EvalLevelRecord
- EngineStructDoubleAction
- EngineStructBestMoveRecord
- HeaderMatchEntry ✅ **FIXED**
- HeaderGameEntry ✅ **FIXED**
- CubeEntry ⚠️ **PARTIALLY FIXED**
- MoveEntry
- FooterGameEntry
- FooterMatchEntry
- GameFileRecord

✅ **xgimport.go** - File import and segment extraction:
- File segment extraction
- Game file validation
- Record parsing

✅ **cmd/xgparser/main.go** - Command-line tool

## Fixed Issues

### HeaderMatchEntry ✅ COMPLETE
**Problem**: Several fields had wrong values due to incorrect struct field types and misalignment
- `Version`: Was 536870912, now correctly reads **32**
- `Invert`: Changed from `int8` to `int32` (Python uses `'l'` format = 4 bytes)
- `Magic`: Now correctly reads **1229737284** (0x494C4D44)
- `CubeLimit`: Now **10** (was 184549375)
- `CommentHeaderMatch`: Now **-1** (was -16777216)
- `Transcribed`: Now **true** (was false)
- `isMoneyMatch`: Now **false** (was true)

**Root Cause**: The Python struct format `'<9x41B41BxllBBBBddlld129Bxxx...'` uses specific padding and field sizes. The Go code was reading `Invert` as `int8` when Python reads it as `'l'` (int32). This created a cascading misalignment for all subsequent fields.

### HeaderGameEntry ✅ COMPLETE
**Problem**: `GameNumber` was 0 instead of 1, `CommentFooterGame` was 16777215 instead of -1

**Solution**: Fixed padding from 4 bytes to 3 bytes. Python format `'<9xxxxllB26bxlBxxxlll'` means:
- `9x` = skip 9 bytes
- `xxx` = skip 3 MORE bytes (each `x` is one byte)
- Total skip = **12 bytes** (not 13!)

**Key Learning**: In Python struct format strings like `'9xxxx'`:
- The number prefix (9) only applies to the immediately following character
- `9x` means "skip 9 bytes"
- `xxx` means "skip 1 + skip 1 + skip 1" = skip 3 bytes
- So `9xxxx` = 9 + 3 = **12 bytes** total

### CubeEntry ⚠️ PARTIALLY FIXED
**Progress**:
- ✅ Fixed initial padding (9 + 3 = 12 bytes, not 13)
- ✅ `ActiveP`: Now **-1** (was -16777217)
- ✅ `Double`: Now **-2** (was -1)
- ⚠️ Still wrong: `AnalyzeC`, `AnalyzeCR`, `CommentCube` fields in second part of struct

**Next Steps**: The second part of CubeEntry parsing after `Doubled` needs padding fixes. Python format is:
```
'<xxxxd3BxxxxxdlllxxxxddllbbxxxxxxddBxxxlBBBxlll'
```
Need to carefully count each padding section.

## Critical Learning: Python struct.unpack Padding Rules

The main challenge in porting from Python to Go is understanding Python's struct format strings:

1. **Number prefixes apply ONLY to the next character**:
   - `'9x'` = skip 9 bytes
   - `'9xxxx'` = skip 9 bytes, then skip 1, then skip 1, then skip 1 = **12 bytes total** (NOT 13!)

2. **Each format character must be counted individually**:
   - `'xxxx'` = 4 separate skip operations = 4 bytes
   - `'4x'` = one skip operation of 4 bytes = 4 bytes (same result, different notation)

3. **Verify with `struct.calcsize()`**: Always test the format string to see actual byte size

4. **Test with synthetic data**: Create test arrays and parse to see which byte positions map to which values

## What Works Now

- File decompression and archive extraction ✅
- Record type identification ✅  
- **HeaderMatchEntry**: All fields parse correctly ✅
  - Version, Magic, GameId, MatchLength, Date ✅
  - Player names, Event, Location (UTF-16 decoded) ✅
  - Boolean fields (Crawford, Jacoby, Transcribed) ✅
  - Monetary fields, ELO ratings ✅
- **HeaderGameEntry**: All fields parse correctly ✅
  - GameNumber, Scores, Position arrays ✅
  - Comment indices ✅
- **CubeEntry**: Basic fields correct, needs completion ⚠️
  - ActiveP, Double, Take, BeaverR ✅
  - Position array ✅
  - Second section needs work ⚠️

## Remaining Work

To make the Go parser output match Python exactly:

1. **Complete CubeEntry parsing**: Fix second struct section padding
2. **Fix MoveEntry parsing**: Apply same padding analysis
3. **Fix FooterGameEntry/FooterMatchEntry**: Apply same techniques
4. **Add comprehensive tests**: Compare Go vs Python output field by field
5. **Create test suite**: Unit tests for each record type

## Debugging Methodology

The successful approach to fix these issues:

1. **Extract binary test data**: Save the raw binary records to files
2. **Parse in Python with struct.unpack**: Get correct values
3. **Examine bytes manually**: Use hex dumps to see exact byte positions
4. **Test format strings**: Use `struct.calcsize()` and synthetic data
5. **Fix padding byte-by-byte**: Match Python format string exactly
6. **Verify**: Compare parsed values against Python output

## Usage

```bash
# Build
go build -o xgparser/xgparser ./cmd/xgparser

# Run
./xgparser/xgparser tmp/test.xg

# Compare with Python
cd tmp/xgdatatools
python3 extractxgdata.py ../test.xg > ../python_output.txt
cd ../..
./xgparser/xgparser tmp/test.xg > go_output.txt
diff python_output.txt go_output.txt
```

## File Structure

```
xgparser/
├── xgutils.go      # Utility functions
├── xgzarc.go       # Archive handling  
├── xgstruct.go     # Data structures ⚠️ IN PROGRESS
└── xgimport.go     # File import

cmd/xgparser/
└── main.go         # CLI tool
```

## Architecture

The package mirrors the Python implementation:

1. **Import** - Reads XG file, extracts Game Data Format header
2. **Archive** - Decompresses embedded ZLib archive containing game data
3. **Segments** - Extracts file segments (temp.xg, temp.xgr, temp.xgc, temp.xgi)
4. **Records** - Parses game file records (Match header, game headers, moves, cube actions, etc.)
5. **Output** - Formats output to match Python pprint style

The main implementation challenge is matching Python's C-style struct packing exactly in Go, which requires careful manual control over byte offsets and understanding Python's struct format notation.
