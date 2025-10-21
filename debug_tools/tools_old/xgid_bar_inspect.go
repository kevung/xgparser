package main

import (
	"fmt"
	"github.com/kevung/xgparser/xgparser"
)

func inspect(path string) {
	move, _, err := xgparser.ParseXGIDFile(path)
	if err != nil {
		fmt.Println(path, "error:", err)
		return
	}
	// Extract XGID components by re-parsing first line in file
	// ParseXGIDFile already parsed and applied swap; we need raw XGID field
	// So open file and read XGID string manually
	// But ParseXGIDFile returned move.Position.Checkers etc; instead, print Position array and ActivePlayer
	fmt.Println("File:", path)
	fmt.Println("  ActivePlayer:", move.ActivePlayer)
	fmt.Println("  position[0] (opponent bar):", move.Position.Checkers[0])
	fmt.Println("  position[25] (player bar):", move.Position.Checkers[25])
	fmt.Println()
}

func main() {
	paths := []string{
		"tmp/xgid/en/XGID=----BaC-B---aD--aa-bcbbBbB:0:0:1:22.txt",
		"tmp/xgid/en/XGID=-A--B-DCC----A------bbbcdA:1:1:-1:4.txt",
		"tmp/xgid/en/XGID=----C-E-A----C-AB------cb-:0:0:-1:6.txt",
		"tmp/xgid/en/XGID=--B-BcBBBB--bA-bBa-c-bb---:0:0:1:00.txt",
		"tmp/xgid/en/XGID=--BCEb---BA-----b-bcbBbb--:0:0:-1:0.txt",
		"tmp/xgid/en/XGID=a--aB-BBA--acDa-Ab-db---BA:0:0:1:64.txt",
	}
	for _, p := range paths {
		inspect(p)
	}
}
