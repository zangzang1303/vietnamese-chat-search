package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// ============================================================================
// CÁC HÀM TIỆN ÍCH ĐỌC / GHI NHỊ PHÂN THEO CHUẨN LITTLE-ENDIAN
// ============================================================================

func WriteUint16(w io.Writer, v uint16) error {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func ReadUint16(r io.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf[:]), nil
}

func WriteUint32(w io.Writer, v uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func ReadUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func WriteUint64(w io.Writer, v uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func ReadUint64(r io.Reader) (uint64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func WriteFloat64(w io.Writer, v float64) error {
	bits := math.Float64bits(v)
	return WriteUint64(w, bits)
}

func ReadFloat64(r io.Reader) (float64, error) {
	bits, err := ReadUint64(r)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// WriteString ghi chuỗi UTF-8 dưới dạng: [Length uint16] + [Bytes chuỗi]
func WriteString(w io.Writer, s string) error {
	b := []byte(s)
	if len(b) > math.MaxUint16 {
		return fmt.Errorf("độ dài chuỗi vượt quá giới hạn uint16 (%d bytes)", len(b))
	}
	if err := WriteUint16(w, uint16(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

// ReadString đọc chuỗi UTF-8 theo định dạng tương ứng
func ReadString(r io.Reader) (string, error) {
	l, err := ReadUint16(r)
	if err != nil {
		return "", err
	}
	if l == 0 {
		return "", nil
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// WriteBytes ghi mảng byte với độ dài uint32
func WriteBytes(w io.Writer, b []byte) error {
	if err := WriteUint32(w, uint32(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

// ReadBytes đọc mảng byte với độ dài uint32
func ReadBytes(r io.Reader) ([]byte, error) {
	l, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}
	if l == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
