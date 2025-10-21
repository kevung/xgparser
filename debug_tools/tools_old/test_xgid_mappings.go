package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
)

func readLines(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    s := bufio.NewScanner(f)
    var lines []string
    for s.Scan() { lines = append(lines, s.Text()) }
    return lines, s.Err()
}

func extractXGID(lines []string) string {
    re := regexp.MustCompile(`^XGID=([^:]+)`)
    for _, l := range lines {
        if m := re.FindStringSubmatch(l); m != nil { return m[1] }
    }
    return ""
}

// Extract simple ASCII counts similar to earlier tool (counts of 'X'/'O' per point by scanning columns)
func extractAsciiCounts(lines []string) (map[int]int, map[int]int) {
    // find header lines with point labels
    var topIdx, botIdx int = -1, -1
    for i,l := range lines {
        if topIdx == -1 && strings.Contains(l, "+13-14-15-16-17-18") { topIdx = i }
        if topIdx != -1 && strings.Contains(l, "+12-11-10--9--8--7") { botIdx = i; break }
    }
    xcounts := make(map[int]int)
    ocnts := make(map[int]int)
    if topIdx == -1 || botIdx == -1 { return xcounts, ocnts }
    top := lines[topIdx]
    bot := lines[botIdx]
    // locate numeric labels positions
    positions := map[int]int{}
    for p:=13;p<=24;p++ {
        s := fmt.Sprintf("%d", p)
        idx := strings.Index(top, s)
        if idx != -1 { positions[p] = idx }
    }
    for p:=12;p>=1;p-- {
        s := fmt.Sprintf("%d", p)
        idx := strings.Index(bot, s)
        if idx != -1 { positions[p] = idx }
    }
    // scan board lines between top and bot
    for _, l := range lines[topIdx+1:botIdx] {
        for p, col := range positions {
            if col < 0 || col >= len(l) { continue }
            ch := l[col]
            if ch == 'X' { xcounts[p]++ }
            if ch == 'O' { ocnts[p]++ }
        }
    }
    return xcounts, ocnts
}

// mapping functions
func mapA(i int) int { return i+1 }       // index->point i+1
func mapB(i int) int { return 24 - i }    // reversed
func mapC(i int) int { return i }         // index->point i (0-based mapping)

func decodeWithMap(posstr string, mapper func(int) int) (map[int]int, map[int]int) {
    xgX := make(map[int]int)
    xgO := make(map[int]int)
    for i := 0; i < len(posstr) && i < 26; i++ {
        c := posstr[i]
        if c == '-' { continue }
        if c >= 'A' && c <= 'Z' {
            count := int(c - 'A' + 1)
            if i < 24 { p := mapper(i); xgX[p] = count }
            // bars ignored for mapping test
        } else if c >= 'a' && c <= 'o' {
            count := int(c - 'a' + 1)
            if i < 24 { p := mapper(i); xgO[p] = count }
        }
    }
    return xgX, xgO
}

func scoreMatch(ascii map[int]int, xgid map[int]int) int {
    score := 0
    for p, cnt := range xgid {
        if ascii[p] > 0 && ascii[p] == cnt { score += 2 }
        if ascii[p] > 0 && ascii[p] != cnt { score += 0 }
    }
    return score
}

func analyze(path string) {
    lines, err := readLines(path)
    if err != nil { fmt.Println(path, "err", err); return }
    posstr := extractXGID(lines)
    if posstr == "" { fmt.Println(path, "no xgid"); return }
    xAscii, oAscii := extractAsciiCounts(lines)
    xA, oA := decodeWithMap(posstr, mapA)
    xB, oB := decodeWithMap(posstr, mapB)
    xC, oC := decodeWithMap(posstr, mapC)
    scoreA := scoreMatch(xAscii, xA) + scoreMatch(oAscii, oA)
    scoreB := scoreMatch(xAscii, xB) + scoreMatch(oAscii, oB)
    scoreC := scoreMatch(xAscii, xC) + scoreMatch(oAscii, oC)
    fmt.Println("File:", path)
    fmt.Println("  scores: A(i+1)", scoreA, "B(24-i)", scoreB, "C(i)", scoreC)
    fmt.Println("  ascii X sample:", xAscii)
    fmt.Println("  xgid A sample:", xA)
    fmt.Println("  xgid B sample:", xB)
    fmt.Println("  xgid C sample:", xC)
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
