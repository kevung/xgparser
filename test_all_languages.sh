#!/bin/bash

# Test all XGID language support
echo "Testing XGID Parser Multi-Language Support"
echo "=========================================="
echo ""

FAIL_COUNT=0
PASS_COUNT=0

for lang_dir in tmp/xgid/*/; do
    lang=$(basename "$lang_dir")
    file=$(ls "$lang_dir"*.txt 2>/dev/null | head -1)
    
    if [ -n "$file" ]; then
        echo -n "Testing $lang... "
        
        # Parse the file and check for key indicators of success
        output=$(./xgid_parser "$file" 2>&1)
        
        # Check if parsing succeeded
        if echo "$output" | grep -q "Analysis entries: [1-9]"; then
            echo "✓ PASS"
            PASS_COUNT=$((PASS_COUNT + 1))
        else
            echo "✗ FAIL"
            echo "  File: $file"
            echo "  Output preview:"
            echo "$output" | head -10 | sed 's/^/    /'
            FAIL_COUNT=$((FAIL_COUNT + 1))
        fi
    fi
done

echo ""
echo "=========================================="
echo "Results: $PASS_COUNT passed, $FAIL_COUNT failed"
echo "=========================================="

exit $FAIL_COUNT
