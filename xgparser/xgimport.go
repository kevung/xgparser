package xgparser

import (
	"bytes"
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

	var segments []*Segment

	// Read and extract the Game Data Format Header
	gdfHeader := &GameDataFormatHdrRecord{}
	err = gdfHeader.FromStream(file)
	if err != nil {
		return nil, fmt.Errorf("not a game data format file: %v", err)
	}

	// Read the full GDF header segment
	file.Seek(0, io.SeekStart)
	gdfData := make([]byte, gdfHeader.HeaderSize)
	_, err = io.ReadFull(file, gdfData)
	if err != nil {
		return nil, err
	}

	segments = append(segments, &Segment{
		Type: SegmentGDFHdr,
		Data: gdfData,
	})

	// Extract thumbnail if present
	if gdfHeader.ThumbnailSize > 0 {
		file.Seek(gdfHeader.ThumbnailOffset, io.SeekCurrent)
		imgData := make([]byte, gdfHeader.ThumbnailSize)
		_, err = io.ReadFull(file, imgData)
		if err != nil {
			return nil, err
		}

		segments = append(segments, &Segment{
			Type: SegmentGDFImage,
			Data: imgData,
		})
	}

	// Get archive object
	archiveObj, err := NewZlibArchive(file)
	if err != nil {
		return nil, err
	}

	// Process all files in the archive
	for _, fileRec := range archiveObj.ArcRegistry {
		data, err := archiveObj.GetArchiveFile(&fileRec)
		if err != nil {
			return nil, err
		}

		segmentType := XGFileMap[fileRec.Name]

		// Verify magic number for game file
		if segmentType == SegmentXGGameFile {
			if len(data) > XGGameHdrLen+4 {
				magicBytes := data[XGGameHdrLen : XGGameHdrLen+4]
				magic := string(magicBytes)
				if magic != "DMLI" {
					return nil, fmt.Errorf("not a valid XG gamefile")
				}
			}
		}

		segments = append(segments, &Segment{
			Type:     segmentType,
			Data:     data,
			Filename: fileRec.Name,
		})
	}

	return segments, nil
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
