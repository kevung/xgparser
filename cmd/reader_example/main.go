//package readerexample

//   reader_example.go - Example of parsing XG files from various sources
//   Copyright (C) 2025 Kevin Unger
//
//   This example demonstrates the flexible parsing API that accepts
//   files from memory, network, or any io.Reader source.
//

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <xgfile>\n", os.Args[0])
		os.Exit(1)
	}

	xgFilename := os.Args[1]

	// Example 1: Parse from file (convenience wrapper)
	fmt.Println("=== Method 1: ParseXGFromFile ===")
	match1, err := xgparser.ParseXGFromFile(xgFilename)
	if err != nil {
		log.Fatalf("Error parsing from file: %v\n", err)
	}
	fmt.Printf("Parsed: %s vs %s\n", match1.Metadata.Player1Name, match1.Metadata.Player2Name)
	fmt.Printf("Games: %d\n\n", len(match1.Games))

	// Example 2: Parse from io.Reader (flexible for network, memory, etc.)
	fmt.Println("=== Method 2: ParseXGFromReader ===")
	file, err := os.Open(xgFilename)
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
	}
	defer file.Close()

	match2, err := xgparser.ParseXGFromReader(file)
	if err != nil {
		log.Fatalf("Error parsing from reader: %v\n", err)
	}
	fmt.Printf("Parsed: %s vs %s\n", match2.Metadata.Player1Name, match2.Metadata.Player2Name)
	fmt.Printf("Games: %d\n\n", len(match2.Games))

	// Example 3: Parse from memory (simulating network download)
	fmt.Println("=== Method 3: ParseXGFromReader (memory buffer) ===")
	// Read entire file into memory
	fileData, err := os.ReadFile(xgFilename)
	if err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	// Create a reader from the memory buffer
	memReader := io.NewSectionReader(
		&bytesReaderAt{fileData},
		0,
		int64(len(fileData)),
	)

	match3, err := xgparser.ParseXGFromReader(memReader)
	if err != nil {
		log.Fatalf("Error parsing from memory: %v\n", err)
	}
	fmt.Printf("Parsed: %s vs %s\n", match3.Metadata.Player1Name, match3.Metadata.Player2Name)
	fmt.Printf("Games: %d\n\n", len(match3.Games))

	// Example 4: Using deprecated ParseXGLight (backward compatibility)
	fmt.Println("=== Method 4: ParseXGLight (deprecated but still works) ===")
	match4, err := xgparser.ParseXGLight(xgFilename)
	if err != nil {
		log.Fatalf("Error parsing: %v\n", err)
	}
	fmt.Printf("Parsed: %s vs %s\n", match4.Metadata.Player1Name, match4.Metadata.Player2Name)
	fmt.Printf("Games: %d\n\n", len(match4.Games))

	fmt.Println("All parsing methods successful!")
	fmt.Println("\nThis demonstrates that XG files can be parsed from:")
	fmt.Println("  - Local filesystem (ParseXGFromFile)")
	fmt.Println("  - Network streams (ParseXGFromReader with net.Conn)")
	fmt.Println("  - Memory buffers (ParseXGFromReader with bytes.Reader)")
	fmt.Println("  - Any io.ReadSeeker source")
}

// bytesReaderAt wraps []byte to implement io.ReaderAt
type bytesReaderAt struct {
	data []byte
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}
