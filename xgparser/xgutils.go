//
//   xgutils.go - XG utilities module (Go port)
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
	"encoding/binary"
	"hash/crc32"
	"io"
	"time"
	"unicode/utf16"
)

// StreamCRC32 computes CRC32 on a reader
func StreamCRC32(r io.ReadSeeker, numBytes int64, startPos int64) (uint32, error) {
	currentPos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	if startPos >= 0 {
		_, err = r.Seek(startPos, io.SeekStart)
		if err != nil {
			return 0, err
		}
	}

	crc := crc32.NewIEEE()
	if numBytes > 0 {
		_, err = io.CopyN(crc, r, numBytes)
	} else {
		_, err = io.Copy(crc, r)
	}
	if err != nil {
		return 0, err
	}

	// Restore position
	_, err = r.Seek(currentPos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return crc.Sum32(), nil
}

// UTF16IntArrayToString converts array of uint16 to string
func UTF16IntArrayToString(data []uint16) string {
	// Find null terminator
	length := 0
	for i, v := range data {
		if v == 0 {
			length = i
			break
		}
	}
	if length == 0 {
		return ""
	}

	runes := utf16.Decode(data[:length])
	return string(runes)
}

// DelphiDateTimeConv converts Delphi datetime (float64) to Go time
func DelphiDateTimeConv(delphiDateTime float64) time.Time {
	// Delphi datetime is days since Dec 30, 1899
	baseDate := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

	days := int64(delphiDateTime)
	fraction := delphiDateTime - float64(days)
	seconds := int64(fraction * 86400)

	return baseDate.AddDate(0, 0, int(days)).Add(time.Duration(seconds) * time.Second)
}

// DelphiShortStrToStr converts Delphi shortstring to Go string
func DelphiShortStrToStr(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	length := int(data[0])
	if length >= len(data) {
		length = len(data) - 1
	}
	return string(data[1 : length+1])
}

// ReadUTF16Array reads an array of uint16 values
func ReadUTF16Array(r io.Reader, count int) ([]uint16, error) {
	result := make([]uint16, count)
	err := binary.Read(r, binary.LittleEndian, &result)
	return result, err
}
