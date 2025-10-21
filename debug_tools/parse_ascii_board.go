package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	fname := "tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt"

	file, _ := os.Open(fname)
	scanner := bufio.NewScanner(file)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	// Find board section
	var topHeader, bottomHeader string
	var boardLines []string

	for i, line := range lines {
		if strings.HasPrefix(line, " +13-14-15") {
			topHeader = line
			// Get board lines until BAR line
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], "BAR") {
					break
				}
				if strings.HasPrefix(lines[j], " |") {
					boardLines = append(boardLines, lines[j])
				}
			}
		}
		if strings.HasPrefix(line, " +12-11-10") {
			bottomHeader = line
			// Get board lines from line before this
			for j := i - 1; j >= 0; j-- {
				if strings.Contains(lines[j], "BAR") {
					break
				}
				if strings.HasPrefix(lines[j], " |") {
					boardLines = append([]string{lines[j]}, boardLines...)
				}
			}
			break
		}
	}

	fmt.Println("Top header:", topHeader)
	fmt.Println("Bottom header:", bottomHeader)
	fmt.Println("\nBoard lines:")
	for i, line := range boardLines {
		fmt.Printf("%2d: %s\n", i, line)
	}

	// Parse point numbers from headers
	numRe := regexp.MustCompile(`\d+`)

	topPoints := make(map[int]int) // point number -> column
	for _, match := range numRe.FindAllStringIndex(topHeader, -1) {
		num, _ := strconv.Atoi(topHeader[match[0]:match[1]])
		// Find the column where checkers would be for this point
		// The number itself starts at match[0], checkers are typically aligned below
		topPoints[num] = match[0] + 1 // Offset slightly
	}

	bottomPoints := make(map[int]int)
	for _, match := range numRe.FindAllStringIndex(bottomHeader, -1) {
		num, _ := strconv.Atoi(bottomHeader[match[0]:match[1]])
		bottomPoints[num] = match[0] + 1
	}

	fmt.Println("\nTop point columns:")
	for pt := 13; pt <= 24; pt++ {
		if col, ok := topPoints[pt]; ok {
			fmt.Printf("Point %2d at column %d\n", pt, col)
		}
	}

	fmt.Println("\nBottom point columns:")
	for pt := 1; pt <= 12; pt++ {
		if col, ok := bottomPoints[pt]; ok {
			fmt.Printf("Point %2d at column %d\n", pt, col)
		}
	}

	// Count checkers
	counts := make(map[int]int)

	// Top half (first half of board lines)
	midpoint := len(boardLines) / 2
	topLines := boardLines[:midpoint]
	bottomLines := boardLines[midpoint:]

	fmt.Println("\nTop half lines (points 13-24):")
	for i, line := range topLines {
		fmt.Printf("%2d: %s\n", i, line)
		for pt, col := range topPoints {
			if col < len(line) {
				if line[col] == 'X' {
					counts[pt]++
				} else if line[col] == 'O' {
					counts[pt]--
				}
			}
		}
	}

	fmt.Println("\nBottom half lines (points 1-12):")
	for i, line := range bottomLines {
		fmt.Printf("%2d: %s\n", i, line)
		for pt, col := range bottomPoints {
			if col < len(line) {
				if line[col] == 'X' {
					counts[pt]++
				} else if line[col] == 'O' {
					counts[pt]--
				}
			}
		}
	}

	fmt.Println("\nChecker counts from ASCII:")
	for pt := 1; pt <= 24; pt++ {
		if counts[pt] != 0 {
			fmt.Printf("Point %2d: %+3d\n", pt, counts[pt])
		}
	}
}
