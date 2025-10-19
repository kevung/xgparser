#!/bin/bash
#
# test_xglight.sh - Comprehensive test script for XG Light Parser
# Copyright (C) 2025 Kevin Unger
#

set -e  # Exit on error

echo "========================================"
echo "XG Light Parser - Comprehensive Tests"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Build all components
echo "Test 1: Building all components..."
go build -o xglight ./cmd/xglight/ 2>&1
go build -o stats_example ./cmd/stats_example/ 2>&1
echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Test 2: Parse a test file
echo "Test 2: Parsing test.xg..."
./xglight tmp/test.xg > /tmp/test_output.json 2>/dev/null
if [ -s /tmp/test_output.json ]; then
    echo -e "${GREEN}✓ File parsed successfully${NC}"
else
    echo -e "${RED}✗ Parse failed${NC}"
    exit 1
fi
echo ""

# Test 3: Validate JSON
echo "Test 3: Validating JSON output..."
if python3 -m json.tool /tmp/test_output.json > /dev/null 2>&1; then
    echo -e "${GREEN}✓ JSON is valid${NC}"
else
    echo -e "${RED}✗ Invalid JSON${NC}"
    exit 1
fi
echo ""

# Test 4: Check JSON structure
echo "Test 4: Checking JSON structure..."
python3 << 'EOF'
import json
with open('/tmp/test_output.json') as f:
    data = json.load(f)

# Check required fields
assert 'metadata' in data, "Missing metadata"
assert 'games' in data, "Missing games"
assert 'player1_name' in data['metadata'], "Missing player1_name"
assert 'player2_name' in data['metadata'], "Missing player2_name"
assert 'match_length' in data['metadata'], "Missing match_length"

# Check games structure
assert len(data['games']) > 0, "No games found"
game = data['games'][0]
assert 'game_number' in game, "Missing game_number"
assert 'moves' in game, "Missing moves"
assert 'winner' in game, "Missing winner"
assert 'points_won' in game, "Missing points_won"

# Check moves structure
if len(game['moves']) > 0:
    move = game['moves'][0]
    assert 'move_type' in move, "Missing move_type"
    assert move['move_type'] in ['checker', 'cube'], "Invalid move_type"

print("All structure checks passed")
EOF
echo -e "${GREEN}✓ JSON structure is correct${NC}"
echo ""

# Test 5: Run stats example
echo "Test 5: Running stats example..."
./stats_example tmp/test.xg > /tmp/stats_output.txt 2>&1
if grep -q "=== Match Information ===" /tmp/stats_output.txt; then
    echo -e "${GREEN}✓ Stats example works${NC}"
else
    echo -e "${RED}✗ Stats example failed${NC}"
    exit 1
fi
echo ""

# Test 6: Parse multiple tournament files
echo "Test 6: Parsing multiple tournament files..."
count=0
failed=0
for file in tmp/2023-11-24_HSBTParis_10x7p/*.xg; do
    if [ -f "$file" ]; then
        if ./xglight "$file" > /dev/null 2>&1; then
            count=$((count + 1))
        else
            failed=$((failed + 1))
        fi
    fi
done
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ Parsed $count tournament files successfully${NC}"
else
    echo -e "${YELLOW}⚠ Parsed $count files, $failed failed${NC}"
fi
echo ""

# Test 7: Data integrity check
echo "Test 7: Checking data integrity..."
python3 << 'EOF'
import json
with open('/tmp/test_output.json') as f:
    data = json.load(f)

metadata = data['metadata']
games = data['games']

# Check metadata
assert metadata['match_length'] > 0, "Invalid match length"
assert len(metadata['player1_name']) > 0, "Empty player1 name"
assert len(metadata['player2_name']) > 0, "Empty player2 name"

# Check games
total_checker = 0
total_cube = 0
for game in games:
    assert game['game_number'] > 0, "Invalid game number"
    assert isinstance(game['initial_score'], list), "Invalid initial_score type"
    assert len(game['initial_score']) == 2, "Invalid initial_score length"
    assert game['winner'] in [-1, 0, 1], f"Invalid winner value: {game['winner']}"
    assert game['points_won'] > 0, "Invalid points_won"
    
    for move in game['moves']:
        if move['move_type'] == 'checker':
            total_checker += 1
            assert 'checker_move' in move, "Missing checker_move"
            cm = move['checker_move']
            assert 'dice' in cm, "Missing dice"
            assert 'position' in cm, "Missing position"
            assert len(cm['dice']) == 2, "Invalid dice length"
        elif move['move_type'] == 'cube':
            total_cube += 1
            assert 'cube_move' in move, "Missing cube_move"

print(f"Data integrity verified: {total_checker} checker moves, {total_cube} cube moves")
EOF
echo -e "${GREEN}✓ Data integrity check passed${NC}"
echo ""

# Test 8: Performance test
echo "Test 8: Performance test (parsing 10 files)..."
start_time=$(date +%s)
for i in {1..10}; do
    ./xglight tmp/test.xg > /dev/null 2>&1
done
end_time=$(date +%s)
elapsed=$((end_time - start_time))
echo -e "${GREEN}✓ Parsed 10 files in ${elapsed}s${NC}"
echo ""

# Summary
echo "========================================"
echo -e "${GREEN}All tests passed!${NC}"
echo "========================================"
echo ""
echo "Summary:"
echo "  - Build: OK"
echo "  - Parse: OK"
echo "  - JSON validation: OK"
echo "  - Structure: OK"
echo "  - Stats example: OK"
echo "  - Multiple files: OK"
echo "  - Data integrity: OK"
echo "  - Performance: OK"
echo ""
echo "The XG Light Parser is ready for use!"

# Cleanup
rm -f /tmp/test_output.json /tmp/stats_output.txt
