//
//   xgzarc.go - XG zlib archive module
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
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

const maxBufSize = 32768

// MaxDecompressedSize caps the size of any segment read from an XG file:
// inflated zlib data, stored data and header blocks alike. It guards a
// reader of untrusted files against decompression bombs and oversized
// allocations. The largest segment of a real match seen on the BMAB europe
// corpus (33 343 files) is 6 259 200 bytes (about 6 MiB, a game file); the
// default of 128 MiB leaves a margin of about x21. Set it before parsing;
// it is read without locking.
var MaxDecompressedSize int64 = 128 << 20

// ErrDecompressionLimit is matched by errors.Is on a SizeLimitError.
var ErrDecompressionLimit = errors.New("xgparser: segment exceeds MaxDecompressedSize")

// SizeLimitError reports a segment larger than MaxDecompressedSize, or a
// size field that is negative.
type SizeLimitError struct {
	What  string // which segment: "zlib stream", "stored file", "GDF header", "thumbnail"
	Size  int64  // declared size, or Limit+1 when inflation went past the limit
	Limit int64
}

func (e *SizeLimitError) Error() string {
	return fmt.Sprintf("xgparser: %s of %d bytes exceeds the limit of %d bytes", e.What, e.Size, e.Limit)
}

// Is makes errors.Is(err, ErrDecompressionLimit) true.
func (e *SizeLimitError) Is(target error) bool { return target == ErrDecompressionLimit }

// checkSize rejects a declared size that is negative or above the limit.
func checkSize(what string, size int64) error {
	if size < 0 || size > MaxDecompressedSize {
		return &SizeLimitError{What: what, Size: size, Limit: MaxDecompressedSize}
	}
	return nil
}

// inflate decompresses one zlib stream from r, refusing to produce more than
// MaxDecompressedSize bytes.
func inflate(r io.Reader) ([]byte, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	limit := MaxDecompressedSize
	data, err := io.ReadAll(io.LimitReader(zr, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, &SizeLimitError{What: "zlib stream", Size: limit + 1, Limit: limit}
	}
	return data, nil
}

// ArchiveRecord represents the archive metadata
type ArchiveRecord struct {
	CRC                uint32
	FileCount          int32
	Version            int32
	RegistrySize       int32
	ArchiveSize        int32
	CompressedRegistry int32
	Reserved           [12]byte
}

// FileRecord represents a file in the archive
type FileRecord struct {
	Name             string
	Path             string
	OSize            int32
	CSize            int32
	Start            int32
	CRC              uint32
	Compressed       byte
	CompressionLevel byte
}

// ZlibArchive represents a zlib compressed archive
type ZlibArchive struct {
	ArcRec         ArchiveRecord
	ArcRegistry    []FileRecord
	StartOfArcData int64
	EndOfArcData   int64
	stream         io.ReadSeeker
}

// NewZlibArchive creates a new ZlibArchive from a stream
func NewZlibArchive(stream io.ReadSeeker) (*ZlibArchive, error) {
	za := &ZlibArchive{
		stream: stream,
	}

	err := za.getArchiveIndex()
	if err != nil {
		return nil, err
	}

	return za, nil
}

// getArchiveIndex reads the archive index
func (za *ZlibArchive) getArchiveIndex() error {
	currentPos, err := za.stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	defer za.stream.Seek(currentPos, io.SeekStart)

	// Read archive record at the end
	_, err = za.stream.Seek(-36, io.SeekEnd) // ArchiveRecord size = 36
	if err != nil {
		return err
	}

	za.EndOfArcData, _ = za.stream.Seek(0, io.SeekCurrent)

	err = binary.Read(za.stream, binary.LittleEndian, &za.ArcRec)
	if err != nil {
		return err
	}

	// Position at beginning of archive file index
	_, err = za.stream.Seek(-36-int64(za.ArcRec.RegistrySize), io.SeekEnd)
	if err != nil {
		return err
	}

	za.StartOfArcData, _ = za.stream.Seek(0, io.SeekCurrent)
	za.StartOfArcData -= int64(za.ArcRec.ArchiveSize)

	// Verify CRC
	crc, err := StreamCRC32(za.stream, za.EndOfArcData-za.StartOfArcData, za.StartOfArcData)
	if err != nil {
		return err
	}
	if crc != za.ArcRec.CRC {
		return fmt.Errorf("archive CRC check failed - file corrupt")
	}

	// Decompress index
	indexData, err := za.extractSegment(za.ArcRec.CompressedRegistry != 0, 0)
	if err != nil {
		return fmt.Errorf("error extracting archive index: %w", err)
	}

	// Read file records from index
	indexReader := bytes.NewReader(indexData)
	za.ArcRegistry = make([]FileRecord, za.ArcRec.FileCount)

	for i := int32(0); i < za.ArcRec.FileCount; i++ {
		var nameBytes [256]byte
		var pathBytes [256]byte

		err = binary.Read(indexReader, binary.LittleEndian, &nameBytes)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &pathBytes)
		if err != nil {
			return err
		}

		za.ArcRegistry[i].Name = DelphiShortStrToStr(nameBytes[:])
		za.ArcRegistry[i].Path = DelphiShortStrToStr(pathBytes[:])

		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].OSize)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].CSize)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].Start)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].CRC)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].Compressed)
		if err != nil {
			return err
		}
		err = binary.Read(indexReader, binary.LittleEndian, &za.ArcRegistry[i].CompressionLevel)
		if err != nil {
			return err
		}

		// Skip padding
		var padding [2]byte
		binary.Read(indexReader, binary.LittleEndian, &padding)
	}

	return nil
}

// extractSegment extracts a compressed or uncompressed segment
func (za *ZlibArchive) extractSegment(isCompressed bool, numBytes int32) ([]byte, error) {
	if isCompressed {
		return inflate(za.stream)
	} else {
		// Read uncompressed segment
		if numBytes == 0 {
			return nil, fmt.Errorf("numBytes must be specified for uncompressed segments")
		}
		if err := checkSize("stored file", int64(numBytes)); err != nil {
			return nil, err
		}

		data := make([]byte, numBytes)
		_, err := io.ReadFull(za.stream, data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}
}

// GetArchiveFile extracts a file from the archive
func (za *ZlibArchive) GetArchiveFile(filerec *FileRecord) ([]byte, error) {
	_, err := za.stream.Seek(int64(filerec.Start)+za.StartOfArcData, io.SeekStart)
	if err != nil {
		return nil, err
	}

	data, err := za.extractSegment(filerec.Compressed == 0, filerec.CSize)
	if err != nil {
		return nil, fmt.Errorf("error extracting archived file: %w", err)
	}

	// Verify CRC
	crc := crc32.ChecksumIEEE(data)
	if crc != filerec.CRC {
		return nil, fmt.Errorf("file CRC check failed - file corrupt")
	}

	return data, nil
}
