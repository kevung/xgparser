package xgparser

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// Offsets inside the match header record, which opens the game file.
const (
	offJacoby             = 101
	offBeaver             = 102
	offElo1               = 104
	offExp2               = 124
	offCommentHeaderMatch = 576
	offCommentFooterMatch = 580
)

func fixtureSegments(t *testing.T, name string) []*Segment {
	t.Helper()
	segs, err := NewImport(filepath.Join("testdata", name)).GetFileSegments()
	if err != nil {
		t.Fatalf("segments of %s: %v", name, err)
	}
	return segs
}

func segmentOf(t *testing.T, segs []*Segment, typ int) *Segment {
	t.Helper()
	for _, s := range segs {
		if s.Type == typ {
			return s
		}
	}
	t.Fatalf("no segment of type %d", typ)
	return nil
}

// The GDF header's GameName reads "Played on <location>": it is the match's
// title, never the product version.
func TestParseXGProductVersionIsNotLocation(t *testing.T) {
	segs := fixtureSegments(t, "test.xg")

	m, err := ParseXG(segs)
	if err != nil {
		t.Fatal(err)
	}
	if m.Metadata.Location != "Paris, Fédération Française de Bridge" {
		t.Fatalf("Location = %q", m.Metadata.Location)
	}
	if m.Metadata.ProductVersion != "" {
		t.Errorf("ProductVersion = %q, want empty", m.Metadata.ProductVersion)
	}

	// Without a match header to overwrite it, the title used to leak through.
	gdfOnly, err := ParseXG([]*Segment{segmentOf(t, segs, SegmentGDFHdr)})
	if err != nil {
		t.Fatal(err)
	}
	if pv := gdfOnly.Metadata.ProductVersion; pv != "" {
		t.Errorf("ProductVersion from GDF header alone = %q, want empty", pv)
	}
}

func TestParseXGHeaderMetadata(t *testing.T) {
	m, err := ParseXGFromFile(filepath.Join("testdata", "test.xg"))
	if err != nil {
		t.Fatal(err)
	}
	md := m.Metadata
	if md.Player1Elo != 1600 || md.Player2Elo != 1600 {
		t.Errorf("Elo = %v/%v, want 1600/1600", md.Player1Elo, md.Player2Elo)
	}
	if md.Player1Experience != 0 || md.Player2Experience != 0 {
		t.Errorf("Experience = %v/%v, want 0/0", md.Player1Experience, md.Player2Experience)
	}
	if md.Transcriber != "Kévin Unger" {
		t.Errorf("Transcriber = %q", md.Transcriber)
	}
	if md.Jacoby || md.Beaver {
		t.Errorf("Jacoby/Beaver = %v/%v, want false/false", md.Jacoby, md.Beaver)
	}
	if md.MatchHeaderComment != "" || md.MatchFooterComment != "" {
		t.Errorf("match comments = %q/%q, want none", md.MatchHeaderComment, md.MatchFooterComment)
	}
}

// The fixtures state neither the session rules nor match comments, so the
// header record of a real file is patched to carry them.
func TestParseXGPatchedHeaderMetadata(t *testing.T) {
	segs := fixtureSegments(t, "match_with_comment.xg")
	comments := parseCommentSegment(segmentOf(t, segs, SegmentXGComment).Data)
	if len(comments) < 2 {
		t.Fatalf("fixture has %d comments, want at least 2", len(comments))
	}

	game := segmentOf(t, segs, SegmentXGGameFile)
	data := append([]byte(nil), game.Data...)
	data[offJacoby], data[offBeaver] = 1, 1
	binary.LittleEndian.PutUint64(data[offElo1:], 0x409F400000000000) // 2000.0
	binary.LittleEndian.PutUint32(data[offExp2:], 345)
	binary.LittleEndian.PutUint32(data[offCommentHeaderMatch:], 0)
	binary.LittleEndian.PutUint32(data[offCommentFooterMatch:], 1)
	game.Data = data

	m, err := ParseXG(segs)
	if err != nil {
		t.Fatal(err)
	}
	md := m.Metadata
	if !md.Jacoby || !md.Beaver {
		t.Errorf("Jacoby/Beaver = %v/%v, want true/true", md.Jacoby, md.Beaver)
	}
	if md.Player1Elo != 2000 || md.Player2Experience != 345 {
		t.Errorf("Player1Elo = %v, Player2Experience = %v", md.Player1Elo, md.Player2Experience)
	}
	if md.MatchHeaderComment != comments[0] || md.MatchFooterComment != comments[1] {
		t.Errorf("match comments = %q/%q, want %q/%q",
			md.MatchHeaderComment, md.MatchFooterComment, comments[0], comments[1])
	}
}

// bareContainer rebuilds name as a GDF header without thumbnail followed by
// the game file as a single zlib stream, with no archive: the layout of
// files found in the BMAB corpus.
func bareContainer(t *testing.T, name string) []byte {
	t.Helper()
	segs := fixtureSegments(t, name)
	hdr := append([]byte(nil), segmentOf(t, segs, SegmentGDFHdr).Data...)
	for i := 12; i < 24; i++ { // ThumbnailOffset, ThumbnailSize
		hdr[i] = 0
	}
	var buf bytes.Buffer
	buf.Write(hdr)
	zw := zlib.NewWriter(&buf)
	zw.Write(segmentOf(t, segs, SegmentXGGameFile).Data)
	zw.Close()
	return buf.Bytes()
}

func TestParseXGBareGameFileContainer(t *testing.T) {
	want, err := ParseXGFromFile(filepath.Join("testdata", "test.xg"))
	if err != nil {
		t.Fatal(err)
	}
	raw := bareContainer(t, "test.xg")
	path := filepath.Join(t.TempDir(), "bare.xg")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	fromFile, err := ParseXGFromFile(path)
	if err != nil {
		t.Fatalf("ParseXGFromFile: %v", err)
	}
	fromReader, err := ParseXGFromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseXGFromReader: %v", err)
	}
	for label, got := range map[string]*Match{"file": fromFile, "reader": fromReader} {
		if got.Metadata != want.Metadata {
			t.Errorf("%s: metadata = %+v, want %+v", label, got.Metadata, want.Metadata)
		}
		if len(got.Games) != len(want.Games) {
			t.Fatalf("%s: %d games, want %d", label, len(got.Games), len(want.Games))
		}
		for i := range want.Games {
			if len(got.Games[i].Moves) != len(want.Games[i].Moves) || got.Games[i].Winner != want.Games[i].Winner {
				t.Errorf("%s: game %d differs", label, i+1)
			}
		}
	}
}

// A truncated bare container is still refused: the fallback accepts only a
// stream that inflates completely and ends at end of file.
func TestParseXGTruncatedBareContainerRefused(t *testing.T) {
	raw := bareContainer(t, "test.xg")
	for _, cut := range []int{1, 100, len(raw) / 2} {
		if _, err := ParseXGFromReader(bytes.NewReader(raw[:len(raw)-cut])); err == nil {
			t.Errorf("cut %d bytes: parsed, want an error", cut)
		}
	}
}

// XG indexes comments by their position in the comment segment, so an empty
// comment still holds its index: the one after it must not move up.
func TestParseXGEmptyCommentKeepsIndices(t *testing.T) {
	segs := fixtureSegments(t, "match_with_comment.xg")
	commentSeg := segmentOf(t, segs, SegmentXGComment)
	original := parseCommentSegment(commentSeg.Data)
	if len(original) < 2 || original[0] == "" || original[1] == "" {
		t.Fatalf("fixture needs two non-empty comments, got %q", original)
	}
	// An empty comment at index 0 shifts every stored comment by one.
	commentSeg.Data = append([]byte("\r\n"), commentSeg.Data...)

	game := segmentOf(t, segs, SegmentXGGameFile)
	data := append([]byte(nil), game.Data...)
	binary.LittleEndian.PutUint32(data[offCommentHeaderMatch:], 0)
	binary.LittleEndian.PutUint32(data[offCommentFooterMatch:], 1)
	game.Data = data

	m, err := ParseXG(segs)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Metadata.MatchHeaderComment; got != "" {
		t.Errorf("comment 0 = %q, want empty", got)
	}
	if got := m.Metadata.MatchFooterComment; got != original[0] {
		t.Errorf("comment 1 = %q, want %q", got, original[0])
	}
}
