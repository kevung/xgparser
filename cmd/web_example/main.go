//package webexample

//   web_example.go - Example of using XG parser in a web application
//   Copyright (C) 2025 Kevin Unger
//
//   This example demonstrates how to use the parser in HTTP handlers
//   for uploading and analyzing XG files via a web interface.
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kevung/xgparser/xgparser"
)

// MatchSummary is a simplified match summary for API responses
type MatchSummary struct {
	Player1     string `json:"player1"`
	Player2     string `json:"player2"`
	Event       string `json:"event"`
	Location    string `json:"location"`
	MatchLength int32  `json:"match_length"`
	NumGames    int    `json:"num_games"`
	TotalMoves  int    `json:"total_moves"`
}

// uploadHandler handles XG file uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (10 MB max)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, _, err := r.FormFile("xgfile")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file into memory
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Create a ReadSeeker from the data
	reader := io.NewSectionReader(bytes.NewReader(fileData), 0, int64(len(fileData)))

	// Parse the XG file
	match, err := xgparser.ParseXGFromReader(reader)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse XG file: %v", err), http.StatusBadRequest)
		return
	}

	// Create summary
	totalMoves := 0
	for _, game := range match.Games {
		totalMoves += len(game.Moves)
	}

	summary := MatchSummary{
		Player1:     match.Metadata.Player1Name,
		Player2:     match.Metadata.Player2Name,
		Event:       match.Metadata.Event,
		Location:    match.Metadata.Location,
		MatchLength: match.Metadata.MatchLength,
		NumGames:    len(match.Games),
		TotalMoves:  totalMoves,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// fullMatchHandler returns the complete match as JSON
func fullMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file
	file, _, err := r.FormFile("xgfile")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read into memory
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Parse
	reader := io.NewSectionReader(bytes.NewReader(fileData), 0, int64(len(fileData)))
	match, err := xgparser.ParseXGFromReader(reader)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse XG file: %v", err), http.StatusBadRequest)
		return
	}

	// Return full match as JSON
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := match.ToJSON()
	w.Write(jsonData)
}

// homeHandler serves a simple HTML form
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>XG Match Analyzer</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; }
        form { margin: 20px 0; }
        input[type="file"] { margin: 10px 0; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        button:hover { background: #0056b3; }
        pre { background: #f4f4f4; padding: 15px; overflow-x: auto; }
        .section { border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>XG Match Analyzer</h1>
    <p>Upload eXtremeGammon files to analyze.</p>
    
    <div class="section">
        <h2>Binary Match Files (.xg)</h2>
        
        <h3>Quick Summary</h3>
        <form action="/upload" method="post" enctype="multipart/form-data" target="summary">
            <input type="file" name="xgfile" accept=".xg" required>
            <button type="submit">Get Summary</button>
        </form>
        <iframe name="summary" style="width:100%; height:200px; border:1px solid #ccc;"></iframe>

        <h3>Full Match JSON</h3>
        <form action="/full" method="post" enctype="multipart/form-data" target="full">
            <input type="file" name="xgfile" accept=".xg" required>
            <button type="submit">Get Full Match</button>
        </form>
        <iframe name="full" style="width:100%; height:400px; border:1px solid #ccc;"></iframe>
    </div>

    <div class="section">
        <h2>Text Position Files (.txt)</h2>
        <p>Supports English, French, German, and Japanese text positions</p>
        
        <form action="/text" method="post" enctype="multipart/form-data" target="textpos">
            <input type="file" name="textfile" accept=".txt" required>
            <button type="submit">Parse Text Position</button>
        </form>
        <iframe name="textpos" style="width:100%; height:500px; border:1px solid #ccc;"></iframe>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// textPositionHandler handles XG text position uploads
func textPositionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file
	file, _, err := r.FormFile("textfile")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse text position
	pos, err := xgparser.ParseXGTextPosition(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse text position: %v", err), http.StatusBadRequest)
		return
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pos.ToJSON())
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/full", fullMatchHandler)
	http.HandleFunc("/text", textPositionHandler)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Upload XG files to analyze matches via web interface")
	fmt.Println("  - /upload : Quick summary of XG match files")
	fmt.Println("  - /full   : Full match analysis of XG files")
	fmt.Println("  - /text   : Parse XG text positions (EN, FR, DE, JP)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
