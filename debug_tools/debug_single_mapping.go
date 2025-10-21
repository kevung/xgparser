package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	fname := "tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt"

	file, _ := os.Open(fname)
	scanner := bufio.NewScanner(file)

	xgidRe := regexp.MustCompile(`^XGID=([^:]+):`)
	var xgidPos string
	var boardLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if match := xgidRe.FindStringSubmatch(line); match != nil {
			xgidPos = match[1]
			fmt.Println("XGID:", xgidPos)
		}
		if len(line) > 10 && (line[0] == ' ' || line[0] == '+' || line[0] == '|') {
			boardLines = append(boardLines, line)
		}
	}
	file.Close()

	fmt.Println("\n=== Board Lines ===")
	for i, line := range boardLines {
		fmt.Printf("%2d: %s\n", i, line)
	}

	// Parse using xgparser
	fmt.Println("\n=== XGParser Decoded Position ===")
	pos := xgparser.XGIDToPosition(xgidPos)
	for i := 1; i <= 24; i++ {
		if pos[i] != 0 {
			fmt.Printf("Point %2d: %+2d\n", i, pos[i])
		}
	}

	// Manual ASCII count
	fmt.Println("\n=== ASCII Board Count (Manual) ===")
	// Top header: +13-14-15-16-17-18------19-20-21-22-23-24-+
	// Bottom header: +12-11-10--9--8--7-------6--5--4--3--2--1-+

	// From the file content shown:
	// Point 13: 4 X
	// Point 17: 1 O
	// Point 18: 2 O
	// Point 20-23: O (varying counts)
	// Point 24: 1 O
	// Point 23: X
	// etc.

	// Let's count manually from board
	counts := make(map[int]int8)

	// Looking at the actual board:
	// Top row (points 13-24):
	//  | X        O  O    |   | O  O  O  O  X  O |
	//  | X                |   | O  O  O  O  X  O |
	//  | X                |   |    O             |
	//  | X                | X |                  |
	//  |                  | X |                  |

	// Point 13: 4X
	counts[13] = 4
	// Point 17: 2O
	counts[17] = -2
	// Point 18: 2O
	counts[18] = -2
	// Point 19: 2O
	counts[19] = -2
	// Point 20: 5O
	counts[20] = -5
	// Point 21: 5O
	counts[21] = -5
	// Point 22: 5O
	counts[22] = -5
	// Point 23: 2X + 1O = mixed!
	// Actually looking more carefully:
	// Point 23: 2X
	counts[23] = 2
	// Point 24: 3O
	counts[24] = -3

	// Bottom row (points 1-12):
	//  |                  |   | X                |
	//  |             X    |   | X     X          |
	//  | O           X    |   | X  O  X          |

	// Point 12: 1O
	counts[12] = -1
	// Point 8: 2X
	counts[8] = 2
	// Point 6: 4X
	counts[6] = 4
	// Point 5: 1O
	counts[5] = -1
	// Point 4: 2X
	counts[4] = 2

	for pt := 1; pt <= 24; pt++ {
		if counts[pt] != 0 {
			fmt.Printf("Point %2d: %+2d\n", pt, counts[pt])
		}
	}

	fmt.Println("\n=== Comparison ===")
	for pt := 1; pt <= 24; pt++ {
		xgid := pos[pt]
		ascii := counts[pt]
		if xgid != ascii {
			fmt.Printf("Point %2d: XGID=%+2d  ASCII=%+2d  DIFF=%+2d\n", pt, xgid, ascii, xgid-ascii)
		}
	}
}
