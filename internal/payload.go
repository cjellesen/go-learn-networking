package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// The protocol will be [VERSION|TYPE|HEADER_SIZE|DATA]
var VERSION uint8 = 1

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize = 10 << 20
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")
var ErrVersionMisMatch = errors.New("payload does not match expected version")
var ErrUnexpectedType = errors.New("could not decode packet - unexpected payload type encountered")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

func Decode(r io.Reader) (Payload, error) {
	var version uint8
	err := binary.Read(r, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}

	if version != VERSION {
		return nil, ErrVersionMisMatch
	}

	var typ uint8
	err = binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, ErrUnexpectedType
	}

	_, err = payload.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

type Binary []byte

func (b Binary) Bytes() []byte  { return b }
func (b Binary) String() string { return string(b) }

func (b Binary) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	err = binary.Write(w, binary.BigEndian, BinaryType)
	if err != nil {
		return 0, err
	}

	n += 1
	err = binary.Write(w, binary.BigEndian, uint32(len(b)))
	if err != nil {
		return n, err
	}

	n += 4
	o, err := w.Write(b)
	return n + int64(o), nil
}

// The Decoder will be responsible for parsing the VERSION and TYPE and selecting the correct type to unmarshall to
func (b *Binary) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}

	var n int64 = 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*b = make([]byte, size)
	o, err := r.Read(*b)
	return n + int64(o), err
}

type String string

func (s String) Bytes() []byte  { return []byte(s) }
func (s String) String() string { return string(s) }

func (s String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	err = binary.Write(w, binary.BigEndian, StringType)
	if err != nil {
		return 0, err
	}

	n += 1
	err = binary.Write(w, binary.BigEndian, uint32(len(s)))
	if err != nil {
		return n, err
	}

	n += 4
	o, err := w.Write([]byte(s))
	return n + int64(o), nil
}

func (s *String) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}

	var n int64 = 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}

	*s = String(buf)
	return n + int64(o), nil
}
