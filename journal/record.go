package journal

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrCorruptRecord  = errors.New("journal: corrupt record")
	ErrPartialTail    = errors.New("journal: partial tail record")
	ErrMidFileCorrupt = errors.New("journal: mid-file corruption")
)

const Magic = "WJRN"

const recordHeaderLen = 4 + 8 + 8 + 4 + 8

type Record struct {
	Sequence  uint64
	Version   int64
	Length    uint32
	Payload   []byte
	Checksum  [32]byte
	Timestamp int64
}

func (r Record) ChecksumBytes() []byte {
	return r.Checksum[:]
}

func (r Record) computeChecksum() [32]byte {
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.BigEndian, r.Sequence)
	_ = binary.Write(buf, binary.BigEndian, r.Version)
	_ = binary.Write(buf, binary.BigEndian, r.Length)
	_ = binary.Write(buf, binary.BigEndian, r.Timestamp)
	_, _ = buf.Write(r.Payload)
	return sha256.Sum256(buf.Bytes())
}

func (r Record) Validate() error {
	if int(r.Length) != len(r.Payload) {
		return fmt.Errorf("%w: length mismatch", ErrCorruptRecord)
	}
	if r.computeChecksum() != r.Checksum {
		return fmt.Errorf("%w: checksum mismatch", ErrCorruptRecord)
	}
	return nil
}

func (r Record) Encode() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, recordHeaderLen+len(r.Payload)+32))
	_, _ = buf.WriteString(Magic)
	_ = binary.Write(buf, binary.BigEndian, r.Sequence)
	_ = binary.Write(buf, binary.BigEndian, r.Version)
	_ = binary.Write(buf, binary.BigEndian, r.Length)
	_ = binary.Write(buf, binary.BigEndian, r.Timestamp)
	_, _ = buf.Write(r.Payload)
	_, _ = buf.Write(r.Checksum[:])
	return buf.Bytes()
}

func DecodeRecord(data []byte) (Record, int, error) {
	if len(data) < recordHeaderLen+32 {
		return Record{}, 0, ErrPartialTail
	}
	if string(data[:4]) != Magic {
		return Record{}, 0, fmt.Errorf("%w: bad magic", ErrCorruptRecord)
	}
	rec := Record{
		Sequence:  binary.BigEndian.Uint64(data[4:12]),
		Version:   int64(binary.BigEndian.Uint64(data[12:20])),
		Length:    binary.BigEndian.Uint32(data[20:24]),
		Timestamp: int64(binary.BigEndian.Uint64(data[24:32])),
	}
	if len(data) < recordHeaderLen+int(rec.Length)+32 {
		return Record{}, 0, ErrPartialTail
	}
	rec.Payload = append([]byte(nil), data[recordHeaderLen:recordHeaderLen+int(rec.Length)]...)
	copy(rec.Checksum[:], data[recordHeaderLen+int(rec.Length):recordHeaderLen+int(rec.Length)+32])
	if err := rec.Validate(); err != nil {
		return Record{}, 0, err
	}
	return rec, recordHeaderLen + int(rec.Length) + 32, nil
}
