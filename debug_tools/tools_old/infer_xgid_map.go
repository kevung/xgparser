package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	files := []string{
		"tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt",
		"tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt",
		"tmp/xgid/en/XGID=----C-E-A----C-AB------cb-:0:0:-1:6.txt",
		"tmp/xgid/en/XGID=--B-BcBBBB--bA-bBa-c-bb---:0:0:1:00.txt",
		"tmp/xgid/en/XGID=--BCEb---BA-----b-bcbBbb--:0:0:-1:0.txt",
		"tmp/xgid/en/XGID=a--aB-BBA--acDa-Ab-db---BA:0:0:1:64.txt",
	}

	xgidRe := regexp.MustCompile(`^XGID=([^:]+)`) 

	for _, f := range files {
		fmt.Println("File:", f)
		file, err := os.Open(f)
		if err != nil {
			fmt.Println("  error opening:", err)
			continue
		}
		s := bufio.NewScanner(file)
		var posstr string
		var boardLines []string
		inBoard := false
		for s.Scan() {
			line := s.Text()
			if pos := xgidRe.FindStringSubmatch(line); pos != nil {
				posstr = pos[1]
			}
			if len(line) > 0 && line[0] == ' ' {
				// board line guess
				boardLines = append(boardLines, line)
				inBoard = true
			} else {
				if inBoard {break}
			}
		}
		file.Close()
		if posstr == "" {
			fmt.Println("  no XGID found")
			continue
		}
		fmt.Println("  XGID string:", posstr)
		for i := 0; i < len(posstr) && i < 26; i++ {
			c := posstr[i]
			if c == '-' { continue }
			var who string
			if c >= 'A' && c <= 'Z' { who = "X" } else { who = "O" }
			count := 0
			if c >= 'A' && c <= 'Z' { count = int(c - 'A' + 1) }
			if c >= 'a' && c <= 'o' { count = int(c - 'a' + 1) }
			idx := i
			point := 24 - i // current assumption
			fmt.Printf("    idx %2d char '%c' -> %s %d ; assumed point %2d\n", idx, c, who, count, point)
		}
		fmt.Println()
	}
}

