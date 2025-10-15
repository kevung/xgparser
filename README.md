# XGParser - Go Implementation

This is a Go implementation of the Python xgdatatools package for parsing ExtremeGammon (.xg) files.

## Current Status

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
- HeaderMatchEntry
- HeaderGameEntry
- CubeEntry
- MoveEntry
- FooterGameEntry
- FooterMatchEntry
- GameFileRecord

✅ **xgimport.go** - File import and segment extraction:
- File segment extraction
- Game file validation
- Record parsing

✅ **cmd/xgparser/main.go** - Command-line tool

## Known Issues

The binary parsing has endianness and struct packing issues that need to be resolved:

1. **Struct Packing**: Python's `struct.unpack` with format strings like `'<9x41B41BxllBBBB...'` precisely defines padding bytes. Go's `binary.Read` doesn't handle the same complex packing, leading to field misalignment.

2. **Version Field**: Reading as 536870912 instead of 32 due to byte order in the struct definition.

3. **UTF-16 Strings**: Some Unicode strings are not being decoded correctly due to alignment issues.

## What Works

- File decompression and archive extraction ✅
- Basic record type identification ✅  
- Many numeric fields parse correctly (MatchLength, GameId, positions, moves) ✅
- The overall structure and flow is correct ✅

## Next Steps to Fix

To make the Go parser output match Python exactly:

1. Rewrite struct parsing to use manual byte reading with explicit offsets instead of `binary.Read`
2. Create a custom binary reader that handles the exact Python struct format strings
3. Test each record type individually to ensure correct field alignment
4. Add comprehensive unit tests comparing Go vs Python output

## Usage

```bash
# Build
go build -o xgparser/xgparser ./cmd/xgparser

# Run
./xgparser/xgparser tmp/test.xg
```

## File Structure

```
xgparser/
├── xgutils.go      # Utility functions
├── xgzarc.go       # Archive handling
├── xgstruct.go     # Data structures
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

The main challenge is matching Python's C-style struct packing exactly in Go, which requires more manual control over byte offsets and alignment.
