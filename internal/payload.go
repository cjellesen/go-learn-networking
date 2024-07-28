package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// The protocol will be [VERSION|TYPE|HEADER_SIZE|DATA]
var VERSION uint8 = 1

type PayloadType uint8

const (
	BinaryType PayloadType = iota + 1
	StringType
	InitializationPacketType
)

const MaxPayloadSize = 10 << 20

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")
var ErrVersionMisMatch = errors.New("payload does not match expected version")
var ErrUnexpectedType = errors.New("could not decode packet - unexpected payload type encountered")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
	GetType() PayloadType
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

	var typ PayloadType
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
	case InitializationPacketType:
		payload = new(InitializationPacket)
	default:
		return nil, ErrUnexpectedType
	}

	_, err = payload.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

type InitializationPacket struct {
	PingIntervalMs uint16
	NRetries       uint16
}

func (i *InitializationPacket) Bytes() []byte {
	payload := make([]byte, 0, 8)
	payload = binary.BigEndian.AppendUint16(payload, i.PingIntervalMs)
	payload = binary.BigEndian.AppendUint16(payload, i.NRetries)
	return payload
}

func (i *InitializationPacket) String() string {
	return fmt.Sprintf(
		"Initialization packet, ping interval %d, n retries %d",
		i.PingIntervalMs,
		i.NRetries,
	)
}

func (i *InitializationPacket) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	err = binary.Write(w, binary.BigEndian, InitializationPacketType)
	if err != nil {
		return n, err
	}

	n += 1

	payload := i.Bytes()
	err = binary.Write(w, binary.BigEndian, uint32(len(payload)))
	if err != nil {
		return n, err
	}

	n += 4
	o, err := w.Write(payload)
	return n + int64(o), nil
}

func (i *InitializationPacket) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}

	var n int64 = 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	var pingInterval uint16
	err = binary.Read(r, binary.BigEndian, &pingInterval)
	if err != nil {
		return n, errors.New("failed to parse ping interval")
	}

	n += 2
	var nRetries uint16
	err = binary.Read(r, binary.BigEndian, &nRetries)
	if err != nil {
		return n, errors.New("failed to parse number of retries before failure")
	}

	n += 2
	i.PingIntervalMs = pingInterval
	i.NRetries = nRetries
	return n, nil
}

func (i InitializationPacket) GetType() PayloadType {
	return InitializationPacketType
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
	if err != nil {
		return n, err
	}
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
	if err != nil {
		return n, err
	}
	return n + int64(o), err
}

func (i Binary) GetType() PayloadType {
	return BinaryType
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
	if err != nil {
		return n, err
	}
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

func (i String) GetType() PayloadType {
	return StringType
}
