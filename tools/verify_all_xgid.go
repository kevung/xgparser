package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kevung/xgparser/xgparser"
)

func main() {
	languages := []string{"de", "en", "es", "fi", "fr", "gr", "it", "jp", "ru"}

	totalFiles := 0
	successFiles := 0
	failedFiles := 0
	var failures []string

	fmt.Println("=== XGID Parser Verification Across All Languages ===\n")

	for _, lang := range languages {
		langDir := filepath.Join("tmp/xgid", lang)

		// Check if directory exists
		if _, err := os.Stat(langDir); os.IsNotExist(err) {
			fmt.Printf("%-4s: Directory not found, skipping\n", lang)
			continue
		}

		// Find all XGID files
		pattern := filepath.Join(langDir, "XGID=*.txt")
		files, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("%-4s: Error finding files: %v\n", lang, err)
			continue
		}

		if len(files) == 0 {
			fmt.Printf("%-4s: No XGID files found\n", lang)
			continue
		}

		langSuccess := 0
		langFailed := 0
		langCubeDecisions := 0
		var langErrors []string

		for _, fpath := range files {
			totalFiles++

			// Detect file type first
			fileType, err := xgparser.DetectXGIDFileType(fpath)
			if err != nil {
				langFailed++
				failedFiles++
				errMsg := fmt.Sprintf("  %s: %v", filepath.Base(fpath), err)
				langErrors = append(langErrors, errMsg)
				failures = append(failures, fmt.Sprintf("[%s] %s: %v", lang, filepath.Base(fpath), err))
				continue
			}

			if fileType == "cube" {
				// Parse as cube decision
				cubeMove, _, err := xgparser.ParseXGIDCubeFile(fpath)
				if err != nil {
					langFailed++
					failedFiles++
					errMsg := fmt.Sprintf("  %s: cube parse error: %v", filepath.Base(fpath), err)
					langErrors = append(langErrors, errMsg)
					failures = append(failures, fmt.Sprintf("[%s] %s: cube parse error: %v", lang, filepath.Base(fpath), err))
					continue
				}
				if cubeMove == nil {
					langFailed++
					failedFiles++
					errMsg := fmt.Sprintf("  %s: cubeMove is nil", filepath.Base(fpath))
					langErrors = append(langErrors, errMsg)
					failures = append(failures, fmt.Sprintf("[%s] %s: cubeMove is nil", lang, filepath.Base(fpath)))
					continue
				}
				langCubeDecisions++
				langSuccess++
				successFiles++
				continue
			}

			// Parse as checker move
			move, metadata, err := xgparser.ParseXGIDFile(fpath)
			if err != nil {
				langFailed++
				failedFiles++
				errMsg := fmt.Sprintf("  %s: %v", filepath.Base(fpath), err)
				langErrors = append(langErrors, errMsg)
				failures = append(failures, fmt.Sprintf("[%s] %s: %v", lang, filepath.Base(fpath), err))
				continue
			}

			// Basic validation checks
			if move == nil {
				langFailed++
				failedFiles++
				errMsg := fmt.Sprintf("  %s: move is nil", filepath.Base(fpath))
				langErrors = append(langErrors, errMsg)
				failures = append(failures, fmt.Sprintf("[%s] %s: move is nil", lang, filepath.Base(fpath)))
				continue
			}

			// Check that position has valid checker counts (total should be <= 30)
			totalCheckers := 0
			for i := 0; i < 26; i++ {
				if move.Position.Checkers[i] > 0 {
					totalCheckers += int(move.Position.Checkers[i])
				} else if move.Position.Checkers[i] < 0 {
					totalCheckers += int(-move.Position.Checkers[i])
				}
			}

			if totalCheckers > 30 {
				langFailed++
				failedFiles++
				errMsg := fmt.Sprintf("  %s: invalid checker count (%d > 30)", filepath.Base(fpath), totalCheckers)
				langErrors = append(langErrors, errMsg)
				failures = append(failures, fmt.Sprintf("[%s] %s: invalid checker count %d", lang, filepath.Base(fpath), totalCheckers))
				continue
			}

			// Check that analysis entries were parsed
			if len(move.Analysis) == 0 {
				langFailed++
				failedFiles++
				errMsg := fmt.Sprintf("  %s: no analysis entries parsed", filepath.Base(fpath))
				langErrors = append(langErrors, errMsg)
				failures = append(failures, fmt.Sprintf("[%s] %s: no analysis entries", lang, filepath.Base(fpath)))
				continue
			}

			// Check that each analysis has a valid Move
			hasInvalidMove := false
			for i, analysis := range move.Analysis {
				if len(analysis.Move) == 0 {
					langFailed++
					failedFiles++
					errMsg := fmt.Sprintf("  %s: analysis[%d] has empty move", filepath.Base(fpath), i)
					langErrors = append(langErrors, errMsg)
					failures = append(failures, fmt.Sprintf("[%s] %s: analysis[%d] empty move", lang, filepath.Base(fpath), i))
					hasInvalidMove = true
					break
				}
			}

			if hasInvalidMove {
				continue
			}

			// Validation passed
			langSuccess++
			successFiles++

			// Show metadata for first file in each language
			if langSuccess == 1 {
				fmt.Printf("%-4s: ✓ Sample: %s\n", lang, filepath.Base(fpath))
				if metadata != nil {
					if metadata.Player1Name != "" && metadata.Player2Name != "" {
						fmt.Printf("      Players: %s vs %s\n", metadata.Player1Name, metadata.Player2Name)
					}
					if move.Position.Score[0] > 0 || move.Position.Score[1] > 0 || metadata.MatchLength > 0 {
						fmt.Printf("      Score: %d-%d /%d pts\n", move.Position.Score[0], move.Position.Score[1], metadata.MatchLength)
					}
				}
				fmt.Printf("      Position: %d checkers total, %d analysis entries\n", totalCheckers, len(move.Analysis))
			}
		}

		// Summary for this language
		if langFailed == 0 {
			if langCubeDecisions > 0 {
				fmt.Printf("%-4s: ✓ All %d files parsed (%d checker moves, %d cube decisions)\n",
					lang, langSuccess, langSuccess-langCubeDecisions, langCubeDecisions)
			} else {
				fmt.Printf("%-4s: ✓ All %d files parsed successfully\n", lang, langSuccess)
			}
		} else {
			fmt.Printf("%-4s: ⚠ %d success, %d failed (%.1f%% success)\n",
				lang, langSuccess, langFailed, 100.0*float64(langSuccess)/float64(langSuccess+langFailed))
			if len(langErrors) > 0 && len(langErrors) <= 3 {
				for _, errMsg := range langErrors {
					fmt.Println(errMsg)
				}
			} else if len(langErrors) > 3 {
				fmt.Printf("  (showing first 3 errors)\n")
				for i := 0; i < 3; i++ {
					fmt.Println(langErrors[i])
				}
			}
		}
		fmt.Println()
	}

	// Final summary
	fmt.Println("=== Overall Summary ===")
	fmt.Printf("Total files:    %d\n", totalFiles)
	fmt.Printf("Success:        %d (%.1f%%)\n", successFiles, 100.0*float64(successFiles)/float64(totalFiles))
	fmt.Printf("Failed:         %d (%.1f%%)\n", failedFiles, 100.0*float64(failedFiles)/float64(totalFiles))

	if failedFiles > 0 {
		fmt.Println("\n=== Failed Files ===")
		for i, failure := range failures {
			fmt.Printf("%3d. %s\n", i+1, failure)
		}
	}

	// Exit with error code if any failures
	if failedFiles > 0 {
		os.Exit(1)
	}
}
