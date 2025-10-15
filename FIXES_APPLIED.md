# Fixes Applied to XGParser Go Implementation

## Date: October 16, 2025

## Summary

Successfully debugged and fixed critical struct parsing issues in the Go implementation of the XGParser. The main problem was incorrect padding byte calculations when translating Python's `struct.unpack` format strings to Go's `binary.Read` calls.

## Key Insight: Python struct Format Padding Rules

The critical discovery was understanding how Python's struct format strings work:

### Rule: Number Prefixes Apply ONLY to the Next Character

In a format string like `'<9xxxxllB'`:
- `9x` means "skip 9 bytes" (the '9' applies only to the first 'x')
- `xxx` means "skip 1 + skip 1 + skip 1" = skip 3 bytes
- `ll` means "read two int32 values" (8 bytes)
- `B` means "read one unsigned byte" (1 byte)

Total: Skip 12 bytes (not 13!), then read data.

### Example

`'<9xxxxllllll26bxx'` breaks down as:
- `<` = little-endian
- `9x` = skip 9 bytes
- `xxx` = skip 3 bytes (each 'x' is separate)
- **Total skip = 12 bytes**
- `llllll` = read 6 × int32 (24 bytes)
- `26b` = read 26 × signed byte (26 bytes)
- `xx` = skip 2 bytes

## Fixes Applied

### 1. HeaderMatchEntry ✅ COMPLETE

**File**: `xgparser/xgstruct.go`

**Problem**: Multiple fields had incorrect values due to struct misalignment

**Changes**:
```go
// BEFORE
Invert int8  // Wrong! Python uses 'l' (int32)

// AFTER  
Invert int32  // Correct!
```

**Results**:
| Field | Before (Wrong) | After (Correct) | Python Value |
|-------|---------------|-----------------|--------------|
| Version | 536870912 | 32 | 32 |
| Magic | 1140850688 | 1229737284 | 1229737284 |
| Invert | varies | 1 | 1 |
| CubeLimit | 184549375 | 10 | 10 |
| CommentHeaderMatch | -16777216 | -1 | -1 |
| Transcribed | false | true | True |
| isMoneyMatch | true | false | False |

### 2. HeaderGameEntry ✅ COMPLETE

**File**: `xgparser/xgstruct.go`

**Problem**: Wrong padding caused GameNumber and CommentFooterGame to be incorrect

**Python Format**: `'<9xxxxllB26bxlBxxxlll'`
- Skip: 9 + 3 = **12 bytes** (not 13!)

**Changes**:
```go
// BEFORE
var padding1 [4]byte  // Wrong! Should be 3

// AFTER
var padding1 [3]byte  // Correct! 9x + xxx = 12 bytes total
```

**Results**:
| Field | Before | After | Python |
|-------|--------|-------|--------|
| GameNumber | 0 | 1 | 1 |
| CommentFooterGame | 16777215 | -1 | -1 |

### 3. CubeEntry ⚠️ PARTIALLY COMPLETE

**File**: `xgparser/xgstruct.go`

**Problem**: Same padding issue in first section

**Python Format**: `'<9xxxxllllll26bxx'`
- Skip: 9 + 3 = **12 bytes**

**Changes**:
```go
// BEFORE
var padding1 [4]byte  // Wrong!

// AFTER
var padding1 [3]byte  // Correct!
```

**Results**:
| Field | Before | After | Python |
|-------|--------|-------|--------|
| ActiveP | -16777217 | -1 | -1 |
| Double | -1 | -2 | -2 |

**Remaining Work**: The second part of CubeEntry (after Doubled struct) still needs padding fixes for AnalyzeC, AnalyzeCR, CommentCube fields.

## Testing Methodology

### Step 1: Extract Binary Data
```python
# Save actual game file bytes for analysis
with open('test_gamefile.bin', 'wb') as f:
    f.write(gamefile_data)
```

### Step 2: Test Python Format Strings
```python
import struct

fmt = '<9xxxxllB26bxlBxxxlll'
size = struct.calcsize(fmt)  # Get total bytes

# Create synthetic test data
test = bytearray(size)
for i in range(size):
    test[i] = i & 0xFF

# Parse and find where first value comes from
unpacked = struct.unpack(fmt, bytes(test))
# Trace back byte positions to find skip count
```

### Step 3: Fix Go Padding
```go
// Match the exact byte skip from Python
var skip [9]byte
var padding [3]byte  // Not [4]!
```

### Step 4: Verify
```bash
# Compare outputs
python3 extractxgdata.py test.xg > python.txt
./xgparser test.xg > go.txt
diff python.txt go.txt
```

## Debugging Commands Used

```bash
# Examine specific byte ranges
python3 << EOF
import struct
data = open('test_gamefile.bin', 'rb').read()
# Show hex bytes
print(' '.join(f'{b:02x}' for b in data[0:30]))
# Parse specific section
unpacked = struct.unpack('<9xxxxllB26bxlBxxxlll', data[0:68])
EOF

# Test format string sizes
python3 -c "import struct; print(struct.calcsize('<9xxxxllB'))"

# Compare field values
grep "Version" python_output.txt
grep "Version" go_output.txt
```

## Next Steps

1. **Complete CubeEntry**: Fix second section parsing
   - Python format: `'<xxxxd3BxxxxxdlllxxxxddllbbxxxxxxddBxxxlBBBxlll'`
   - Need to count each padding section carefully

2. **Fix MoveEntry**: Apply same methodology
   - Python format: `'<9x26b26bxxxl8l2lldl'` then `'<Bxxxddlxxxxd32llll26bbxdBxxxl'`

3. **Fix FooterGameEntry**: 
   - Python format: `'<9xxxxllBxxxlllxxxxdd7dl'`

4. **Create Unit Tests**: Test each record type independently

5. **Full Integration Test**: Parse entire file and verify all records match

## Lessons Learned

1. **Never assume padding**: Always verify with `struct.calcsize()`

2. **Python format notation is tricky**: `'9xxxx'` ≠ `'13x'`
   - `'9xxxx'` = 12 bytes (9 + 1 + 1 + 1)
   - `'12x'` = 12 bytes (more explicit!)

3. **Test with synthetic data**: Create test arrays to trace byte positions

4. **Fix one struct at a time**: Verify each before moving to the next

5. **Binary analysis is essential**: Hex dumps reveal the truth

## Files Modified

- `/home/unger/src/xgparser/xgparser/xgstruct.go`
  - Changed `Invert` type from `int8` to `int32`
  - Fixed padding in `HeaderGameEntry.FromStream` ([4] → [3])
  - Fixed padding in `CubeEntry.FromStream` ([4] → [3])

- `/home/unger/src/xgparser/README.md`
  - Complete rewrite with progress status
  - Added debugging methodology section
  - Documented padding rules

## Contact

For questions about these fixes or to continue the work, all test files and binary samples are in:
- `/home/unger/src/xgparser/tmp/test_gamefile.bin`
- `/home/unger/src/xgparser/tmp/log` (Python reference output)
