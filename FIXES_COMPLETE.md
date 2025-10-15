# Fixes Completed - CubeEntry

## Summary
Successfully fixed CubeEntry parsing by correcting the EngineStructDoubleAction struct field types.

## The Problem
CubeEntry fields were showing incorrect values:
- `AnalyzeC`, `AnalyzeCR`, `CommentCube`: 65535 or 0 instead of -1
- `CompChoiceD`: -65536 instead of 0
- `DiceRolled`: Empty string instead of dice values
- `ErrCube`, `ErrBeaver`: Garbage floats instead of -1000.0

## The Root Cause
In EngineStructDoubleAction, the `Crawford` field was defined as `int32` when it should be `int16`.

### Python Format Analysis
The Python struct format `'<26bxxl2llllhhhh7ffffhh7f'` (132 bytes) breaks down as:
```
26b    = 26 int8s  (Pos array)
xx     = 2 bytes padding
l      = 1 int32   (Level)
2l     = 2 int32s  (Score)
lll    = 3 int32s  (Cube, CubePos, Jacoby)  
hhhh   = 4 int16s  (Crawford, met, FlagDouble, isBeaver)
7f     = 7 float32s (Eval)
fff    = 3 float32s (equB, equDouble, equDrop)
hh     = 2 int16s  (LevelRequest, DoubleChoice3)
7f     = 7 float32s (EvalDouble)
```

The critical insight: `l2llll` is NOT 7 int32s! It parses as:
- `l` (1 int32) + `2l` (2 int32s) + `lll` (3 int32s) = **6 int32s total**

## The Fix
Changed in `/home/unger/src/xgparser/xgparser/xgstruct.go`:
```go
// Before:
Crawford      int32

// After:
Crawford      int16
```

This corrected the byte alignment, allowing all subsequent fields in CubeEntry to read correctly.

## Verification
Go output now matches Python output for CubeEntry:
```
✅ ActiveP: -1
✅ Double: -2  
✅ AnalyzeC: -1
✅ AnalyzeCR: -1
✅ CommentCube: -1
✅ CompChoiceD: 0
✅ DiceRolled: '43'
✅ ErrCube: -1000
✅ ErrBeaver: -1000
```

## Next Steps
MoveEntry shows similar issues with wrong values for:
- AnalyzeL (shows 262143 instead of 3)
- AnalyzeM (shows -1 instead of 3)  
- CompChoice (shows -1449787392 instead of 1)

Likely needs similar struct field type corrections.
