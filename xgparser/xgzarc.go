//
//   xgzarc.go - XG zlib archive module (Go port)
//   Original Python version Copyright (C) 2013,2014 Michael Petch <mpetch@gnubg.org>
//   Go port Copyright (C) 2025 Kevin Unger
//
//   This library is free software; you can redistribute it and/or
//   modify it under the terms of the GNU Lesser General Public
//   License as published by the Free Software Foundation; either
//   version 2.1 of the License, or (at your option) any later version.
//
//   This library is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//   Lesser General Public License for more details.
//
//   You should have received a copy of the GNU Lesser General Public
//   License along with this library; if not, write to the Free Software
//   Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301
//   USA
//
//   This is a Go transcoding of the original Python xgdatatools library
//   by Michael Petch, available at https://github.com/oysteijo/xgdatatools
//

package xgparser

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const maxBufSize = 32768

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
		return fmt.Errorf("error extracting archive index: %v", err)
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
		// Decompress the segment
		r, err := zlib.NewReader(za.stream)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	} else {
		// Read uncompressed segment
		if numBytes == 0 {
			return nil, fmt.Errorf("numBytes must be specified for uncompressed segments")
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
		return nil, fmt.Errorf("error extracting archived file: %v", err)
	}

	// Verify CRC
	crc := crc32.ChecksumIEEE(data)
	if crc != filerec.CRC {
		return nil, fmt.Errorf("file CRC check failed - file corrupt")
	}

	return data, nil
}
