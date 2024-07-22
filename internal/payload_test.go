package internal

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
)

func TestStringPayload(t *testing.T) {
	msg := String("this is a test")

	var buffer bytes.Buffer
	w := io.Writer(&buffer)
	msg.WriteTo(w)

	r := io.Reader(&buffer)
	decodedMsg, err := Decode(r)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("failed to decode the test string, failed with error: %q", err)
	}

	if !reflect.DeepEqual(msg.Bytes(), decodedMsg.Bytes()) {
		t.Fatalf(
			"the decoded message is different from the original message, expected:\n %q\ngot:%q",
			msg.String(),
			decodedMsg.String(),
		)
	}
}

func TestBinaryPayload(t *testing.T) {
	msg := Binary([]byte{1, 2, 3, 4, 5})

	var buffer bytes.Buffer
	w := io.Writer(&buffer)
	msg.WriteTo(w)

	r := io.Reader(&buffer)
	decodedMsg, err := Decode(r)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("failed to decode the test string, failed with error: %q", err)
	}

	if !reflect.DeepEqual(msg.Bytes(), decodedMsg.Bytes()) {
		t.Fatalf("the decoded message is different from the original message")
	}
}
func TestPayloadsTransfers(t *testing.T) {
	b1 := Binary("Clear is better than clever")
	b2 := Binary("Don't panic")
	s1 := String("Errors are values.")
	payloads := []Payload{&b1, &b2, &s1}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		for _, p := range payloads {
			_, err := p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		actual, err := Decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("Value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T] %[1]q", actual)
	}
}
