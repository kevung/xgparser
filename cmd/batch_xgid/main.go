//package batchxgid

//   batch_xgid - Batch process XGID files and output JSON
//   Copyright (C) 2025 Kevin Unger
//

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevung/xgparser/xgparser"
)

type BatchResult struct {
	Filename string                  `json:"filename"`
	Language string                  `json:"language"`
	Move     *xgparser.CheckerMove   `json:"move,omitempty"`
	Metadata *xgparser.MatchMetadata `json:"metadata,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

type BatchOutput struct {
	TotalFiles   int           `json:"total_files"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Results      []BatchResult `json:"results"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: batch_xgid <directory>")
		fmt.Println("Example: batch_xgid tmp/xgid")
		os.Exit(1)
	}

	rootDir := os.Args[1]
	output := BatchOutput{
		Results: make([]BatchResult, 0),
	}

	// Walk through all subdirectories
	err := filepath.Walk(rootDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		// Only process .txt files
		if !strings.HasSuffix(path, ".txt") {
			return nil
		}

		output.TotalFiles++

		// Determine language from directory path
		language := "unknown"
		rel, _ := filepath.Rel(rootDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) > 0 {
			language = parts[0]
		}

		// Parse the file
		move, metadata, err := xgparser.ParseXGIDFile(path)

		result := BatchResult{
			Filename: path,
			Language: language,
		}

		if err != nil {
			result.Error = err.Error()
			output.ErrorCount++
		} else {
			result.Move = move
			result.Metadata = metadata
			output.SuccessCount++
		}

		output.Results = append(output.Results, result)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Output JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
