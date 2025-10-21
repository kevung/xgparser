package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Find all XGID files in tmp/xgid/en/
	files, err := filepath.Glob("tmp/xgid/en/XGID=*.txt")
	if err != nil {
		fmt.Println("Error finding files:", err)
		return
	}

	xgidRe := regexp.MustCompile(`^XGID=([^:]+):`)
	totalFiles := 0
	matchedFiles := 0
	mismatchFiles := 0

	fmt.Println("=== Verifying XGID Mapping (i -> point i+1) ===")
	fmt.Println()

	for _, fpath := range files {
		totalFiles++
		file, err := os.Open(fpath)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", fpath, err)
			continue
		}

		scanner := bufio.NewScanner(file)
		var xgidPos string
		var boardLines []string

		// Extract XGID and ASCII board
		for scanner.Scan() {
			line := scanner.Text()
			if match := xgidRe.FindStringSubmatch(line); match != nil {
				xgidPos = match[1]
			}
			// Board lines typically start with space
			if len(line) > 0 && (line[0] == ' ' || line[0] == '+') {
				boardLines = append(boardLines, line)
			}
		}
		file.Close()

		if xgidPos == "" || len(boardLines) < 3 {
			fmt.Printf("SKIP %s (no XGID or board)\n", filepath.Base(fpath))
			continue
		}

		// Decode XGID with i+1 mapping
		decoded := decodeXGID(xgidPos)

		// Parse ASCII board
		asciiCounts := parseASCIIBoard(boardLines)

		// Compare
		matches := 0
		mismatches := 0
		var mismatchDetails []string

		for point := 1; point <= 24; point++ {
			xCount := decoded[point]
			aCount := asciiCounts[point]

			if xCount == aCount {
				matches++
			} else {
				mismatches++
				mismatchDetails = append(mismatchDetails,
					fmt.Sprintf("  Point %2d: XGID=%+2d ASCII=%+2d", point, xCount, aCount))
			}
		}

		if mismatches == 0 {
			matchedFiles++
			fmt.Printf("✓ %s (all 24 points match)\n", filepath.Base(fpath))
		} else {
			mismatchFiles++
			fmt.Printf("✗ %s (matches=%d mismatches=%d)\n", filepath.Base(fpath), matches, mismatches)
			if mismatches <= 5 {
				for _, detail := range mismatchDetails {
					fmt.Println(detail)
				}
			}
		}
	}

	fmt.Println()
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("Total files: %d\n", totalFiles)
	fmt.Printf("Perfect matches: %d (%.1f%%)\n", matchedFiles, 100.0*float64(matchedFiles)/float64(totalFiles))
	fmt.Printf("Mismatches: %d (%.1f%%)\n", mismatchFiles, 100.0*float64(mismatchFiles)/float64(totalFiles))
}

// decodeXGID decodes XGID position string using i+1 mapping
// Returns array where index 0=unused, 1-24=points, 25=unused (bars not checked here)
func decodeXGID(posStr string) [26]int8 {
	var result [26]int8

	if len(posStr) < 26 {
		return result
	}

	for i := 0; i < 24; i++ { // Only decode points 0-23
		c := posStr[i]
		var count int8

		if c == '-' {
			count = 0
		} else if c >= 'A' && c <= 'Z' {
			count = int8(c - 'A' + 1)
		} else if c >= 'a' && c <= 'o' {
			count = -int8(c - 'a' + 1)
		}

		// Map: XGID index i -> internal point i+1
		result[i+1] = count
	}

	return result
}

// parseASCIIBoard extracts checker counts from ASCII board diagram
// Returns array where index 0=unused, 1-24=points
func parseASCIIBoard(lines []string) [26]int8 {
	var counts [26]int8

	// Find the header line with point numbers
	// Format: +13-14-15-16-17-18------19-20-21-22-23-24-+
	var headerLine string
	var boardLines []string

	for _, line := range lines {
		if strings.Contains(line, "13-14-15") {
			headerLine = line
		} else if strings.Contains(line, "12-11-10") {
			// Process top half before moving to bottom
			if headerLine != "" {
				parseTopHalf(headerLine, boardLines, counts[:])
			}
			headerLine = line
			boardLines = nil
		} else if len(line) > 10 && strings.Contains(line, "|") && (strings.Contains(line, "X") || strings.Contains(line, "O")) {
			boardLines = append(boardLines, line)
		}
	}

	// Process bottom half
	if headerLine != "" && strings.Contains(headerLine, "12-11-10") {
		parseBottomHalf(headerLine, boardLines, counts[:])
	}

	return counts
}

func parseTopHalf(header string, lines []string, counts []int8) {
	// Header format: +13-14-15-16-17-18------19-20-21-22-23-24-+
	// Extract column positions for each point number
	pointCols := extractPointColumns(header)

	for _, line := range lines {
		for point, col := range pointCols {
			if point < 13 || point > 24 {
				continue
			}
			if col < len(line) {
				if line[col] == 'X' {
					counts[point]++
				} else if line[col] == 'O' {
					counts[point]--
				}
			}
		}
	}
}

func parseBottomHalf(header string, lines []string, counts []int8) {
	// Header format: +12-11-10--9--8--7-------6--5--4--3--2--1-+
	// Extract column positions for each point number
	pointCols := extractPointColumns(header)

	for _, line := range lines {
		for point, col := range pointCols {
			if point < 1 || point > 12 {
				continue
			}
			if col < len(line) {
				if line[col] == 'X' {
					counts[point]++
				} else if line[col] == 'O' {
					counts[point]--
				}
			}
		}
	}
}

func extractPointColumns(header string) map[int]int {
	// Parse header line like: +13-14-15-16-17-18------19-20-21-22-23-24-+
	// Returns map of point number -> column position
	cols := make(map[int]int)

	numRe := regexp.MustCompile(`\d+`)
	matches := numRe.FindAllStringIndex(header, -1)

	for _, match := range matches {
		pointNum := 0
		fmt.Sscanf(header[match[0]:match[1]], "%d", &pointNum)
		if pointNum >= 1 && pointNum <= 24 {
			// Use the first character position of the number as the column
			cols[pointNum] = match[0]
		}
	}

	return cols
}
