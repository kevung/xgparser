//
//   xgimport.go - XG import module
//   Copyright (C) 2025 Kévin Unger
//
//   Released under the MIT License; see LICENSE at the repository root.
//
//   The .xg / .xgp format was first documented publicly by Michael Petch
//   <mpetch@gnubg.org> in the Python xgdatatools library; that description
//   is credited here, the code below is this repository's own.
//

package xgparser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	SegmentGDFHdr = iota
	SegmentGDFImage
	SegmentXGGameHdr
	SegmentXGGameFile
	SegmentXGRollouts
	SegmentXGComment
	SegmentZlibArcIdx
	SegmentXGUnknown
)

const XGGameHdrLen = 556

var SegmentExtensions = []string{
	"_gdh.bin",
	".jpg",
	"_gamehdr.bin",
	"_gamefile.bin",
	"_rollouts.bin",
	"_comments.bin",
	"_idx.bin",
	"",
}

var XGFileMap = map[string]int{
	"temp.xgi": SegmentXGGameHdr,
	"temp.xgr": SegmentXGRollouts,
	"temp.xgc": SegmentXGComment,
	"temp.xg":  SegmentXGGameFile,
}

// Segment represents a file segment
type Segment struct {
	Type     int
	Data     []byte
	Filename string
}

// Import handles XG file import
type Import struct {
	Filename string
}

// NewImport creates a new Import
func NewImport(filename string) *Import {
	return &Import{Filename: filename}
}

// GetFileSegments extracts all segments from the XG file
func (imp *Import) GetFileSegments() ([]*Segment, error) {
	file, err := os.Open(imp.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readSegments(file, true)
}

// readSegments splits an XG stream into its segments: the GDF header, the
// optional thumbnail, then the files of the zlib archive. checkMagic rejects
// a game file without the "DMLI" signature.
//
// Some files carry no archive at all: after the GDF header (with no
// thumbnail) comes the game file as a single zlib stream, without the
// archive index and trailer. Such files are complete; read as an archive,
// their last bytes give a nonsense trailer, hence a seek error or a failed
// archive CRC. They are recognised after the archive read fails, and only
// when the stream inflates to a game file ending exactly at end of file.
func readSegments(r io.ReadSeeker, checkMagic bool) ([]*Segment, error) {
	gdfHeader := &GameDataFormatHdrRecord{}
	if err := gdfHeader.FromStream(r); err != nil {
		return nil, fmt.Errorf("not a game data format file: %v", err)
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if err := checkSize("GDF header", int64(gdfHeader.HeaderSize)); err != nil {
		return nil, err
	}
	gdfData := make([]byte, gdfHeader.HeaderSize)
	if _, err := io.ReadFull(r, gdfData); err != nil {
		return nil, err
	}
	segments := []*Segment{{Type: SegmentGDFHdr, Data: gdfData}}

	if gdfHeader.ThumbnailSize > 0 {
		r.Seek(gdfHeader.ThumbnailOffset, io.SeekCurrent)
		if err := checkSize("thumbnail", int64(gdfHeader.ThumbnailSize)); err != nil {
			return nil, err
		}
		imgData := make([]byte, gdfHeader.ThumbnailSize)
		if _, err := io.ReadFull(r, imgData); err != nil {
			return nil, err
		}
		segments = append(segments, &Segment{Type: SegmentGDFImage, Data: imgData})
	}

	dataStart, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	archiveObj, err := NewZlibArchive(r)
	if err != nil {
		if gdfHeader.ThumbnailSize == 0 {
			data, ok, bareErr := bareGameFile(r, dataStart)
			if bareErr != nil {
				return nil, bareErr
			}
			if ok {
				return append(segments, &Segment{
					Type:     SegmentXGGameFile,
					Data:     data,
					Filename: "temp.xg",
				}), nil
			}
		}
		return nil, err
	}

	for _, fileRec := range archiveObj.ArcRegistry {
		data, err := archiveObj.GetArchiveFile(&fileRec)
		if err != nil {
			return nil, err
		}

		segmentType := XGFileMap[fileRec.Name]
		if checkMagic && segmentType == SegmentXGGameFile && !hasGameFileMagic(data) && len(data) > XGGameHdrLen+4 {
			return nil, fmt.Errorf("not a valid XG gamefile")
		}

		segments = append(segments, &Segment{
			Type:     segmentType,
			Data:     data,
			Filename: fileRec.Name,
		})
	}

	return segments, nil
}

// gameFileRecordSize is the fixed size of every record of an XG game file.
const gameFileRecordSize = 2560

// hasGameFileMagic reports whether data carries the game file signature.
func hasGameFileMagic(data []byte) bool {
	return len(data) >= XGGameHdrLen+4 && string(data[XGGameHdrLen:XGGameHdrLen+4]) == "DMLI"
}

// bareGameFile reads, from offset start, a game file stored as one zlib
// stream with no archive around it. It accepts the stream only if it ends
// exactly at end of file and inflates to whole records with the game file
// signature, so a genuinely corrupt archive still reports its own error.
// A stream inflating past MaxDecompressedSize is reported as such.
func bareGameFile(r io.ReadSeeker, start int64) ([]byte, bool, error) {
	if _, err := r.Seek(start, io.SeekStart); err != nil {
		return nil, false, nil
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, false, nil
	}
	br := bytes.NewReader(raw)
	data, err := inflate(br)
	if errors.Is(err, ErrDecompressionLimit) {
		return nil, false, err
	}
	if err != nil || br.Len() != 0 {
		return nil, false, nil
	}
	if len(data) == 0 || len(data)%gameFileRecordSize != 0 || !hasGameFileMagic(data) {
		return nil, false, nil
	}
	return data, true, nil
}

// ParseGameFile parses the game file segment and returns records
func ParseGameFile(data []byte, version int32) ([]interface{}, error) {
	reader := bytes.NewReader(data)
	var records []interface{}

	for {
		rec := &GameFileRecord{}
		err := rec.FromStream(reader, version)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Check if we're at the end
			if reader.Len() == 0 {
				break
			}
			return nil, err
		}

		if rec.Record != nil {
			records = append(records, rec.Record)

			// Update version if this is a HeaderMatchEntry
			if hme, ok := rec.Record.(*HeaderMatchEntry); ok {
				version = hme.Version
			}
		}
	}

	return records, nil
}
