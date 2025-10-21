package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// simple helpers
func readLines(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    s := bufio.NewScanner(f)
    var lines []string
    for s.Scan() { lines = append(lines, s.Text()) }
    return lines, s.Err()
}

// find header indices
func findBoardBounds(lines []string) (int,int) {
    topRe := regexp.MustCompile(`\+13-14-15-16-17-18`)
    botRe := regexp.MustCompile(`\+12-11-10`) 
    top, bot := -1, -1
    for i,l := range lines {
        if top == -1 && topRe.FindStringIndex(l) != nil { top = i }
        if top != -1 && botRe.FindStringIndex(l) != nil { bot = i; break }
    }
    return top, bot
}

// map point numbers (1..24) to column positions using header lines
func mapPointColumns(topLine, botLine string) map[int]int {
    m := make(map[int]int)
    // find positions of labels 13..24 in topLine
    for p := 13; p <= 24; p++ {
        lbl := strconv.Itoa(p)
        idx := strings.Index(topLine, lbl)
        if idx >= 0 { m[p] = idx }
    }
    // bottom line labels 12..1
    for p := 12; p >= 1; p-- {
        lbl := strconv.Itoa(p)
        idx := strings.Index(botLine, lbl)
        if idx >= 0 { m[p] = idx }
    }
    return m
}

// count X and O for each point by scanning board rows at mapped columns
func countFromAscii(lines []string, top, bot int, colmap map[int]int) (map[int]int, map[int]int) {
    xCounts := make(map[int]int)
    oCounts := make(map[int]int)
    if top+1 >= bot { return xCounts, oCounts }
    boardLines := lines[top+1:bot]
    for point, col := range colmap {
        // check each board line for X/O at or near the column
        cntX, cntO := 0,0
        for _, l := range boardLines {
            // guard column index
            if col < 0 || col >= len(l) { continue }
            ch := l[col]
            if ch == 'X' { cntX++ }
            if ch == 'O' { cntO++ }
            // also check col+1 for two-digit alignment
            if col+1 < len(l) {
                ch2 := l[col+1]
                if ch2 == 'X' { cntX++ }
                if ch2 == 'O' { cntO++ }
            }
        }
        if cntX > 0 { xCounts[point] = cntX }
        if cntO > 0 { oCounts[point] = cntO }
    }
    return xCounts, oCounts
}

// decode XGID string into counts by current XGIDToPosition-like logic
func decodeXGIDpos(posstr string) ([26]int, [26]int) {
    var xarr [26]int
    var oarr [26]int
    for i := 0; i < len(posstr) && i < 26; i++ {
        c := posstr[i]
        if c == '-' { continue }
        if c >= 'A' && c <= 'Z' {
            count := int(c - 'A' + 1)
            // current assumption: index i maps to point (24-i), bars special
            if i < 24 {
                point := 24 - i
                xarr[point] = count
            } else if i == 24 {
                // bar pos - assume XGID[24] = opponent bar
                oarr[0] = count
            } else if i == 25 {
                xarr[25] = count
            }
        } else if c >= 'a' && c <= 'o' {
            count := int(c - 'a' + 1)
            if i < 24 {
                point := 24 - i
                oarr[point] = count
            } else if i == 24 {
                oarr[0] = count
            } else if i == 25 {
                xarr[25] = count
            }
        }
    }
    return xarr, oarr
}

func analyze(path string) {
    lines, err := readLines(path)
    if err != nil { fmt.Println(path, "err", err); return }

    // extract XGID string
    xgidRe := regexp.MustCompile(`^XGID=([^:]+)`) 
    posstr := ""
    for _, l := range lines { if m := xgidRe.FindStringSubmatch(l); m != nil { posstr = m[1]; break } }
    if posstr == "" { fmt.Println(path, "no xgid"); return }

    top, bot := findBoardBounds(lines)
    if top == -1 || bot == -1 { fmt.Println(path, "cannot find board"); return }
    colmap := mapPointColumns(lines[top], lines[bot])
    xAscii, oAscii := countFromAscii(lines, top, bot, colmap)

    fmt.Println("File:", path)
    fmt.Println("  XGID:", posstr)
    fmt.Println("  ASCII counts (X):", xAscii)
    fmt.Println("  ASCII counts (O):", oAscii)

    xgX, xgO := decodeXGIDpos(posstr)
    fmt.Print("  XGID decode (points nonzero):")
    for i:=1;i<=24;i++ { if xgX[i] != 0 || xgO[i] != 0 { fmt.Printf(" %d(X:%d,O:%d)", i, xgX[i], xgO[i]) } }
    fmt.Println()
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
    for _, p := range paths { analyze(p) }
}
