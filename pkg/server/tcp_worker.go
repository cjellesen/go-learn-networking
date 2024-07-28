package server

import (
	"context"
	"errors"
	"fmt"
	"go-learn-networking/internal"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
)

type TcpWorker struct {
	id        uuid.UUID
	log       *log.Logger
	connQueue chan net.Conn
}

func NewTcpWorker(connQueue chan net.Conn) TcpWorker {
	uuid := uuid.New()
	logger := log.New(
		os.Stdout,
		fmt.Sprintf("Woker (%s): ", uuid.String()),
		log.LUTC,
	)
	return TcpWorker{
		connQueue: connQueue,
		id:        uuid,
		log:       logger,
	}
}

func (t *TcpWorker) Start(ctx context.Context) {
	t.log.Printf("Worker is ready to process connections")
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-t.connQueue:
			err := t.ProcessConnection(conn)
			if err != nil {
				t.log.Printf("Failed to process connection, failed with error: %s\n", err)
			}
		default:
			continue
		}
	}
}

func (t *TcpWorker) ProcessConnection(conn net.Conn) error {
	t.log.Printf("Processing connection")
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		conn.Close()
	}()
	initializationPacket, err := t.initializeConnection(conn)
	if err != nil {
		return err
	}

	reset_channel := make(chan time.Duration, 1)
	reset_interval := time.Millisecond * time.Duration(initializationPacket.PingIntervalMs)
	reset_channel <- reset_interval

	go t.Ping(ctx, conn, reset_channel)
	err = conn.SetDeadline(time.Now().Add(t.determineConnectionDeadline(reset_interval)))
	if err != nil {
		return err
	}

	go t.readFromConnection(conn, reset_channel)

	// If I remember correctly this should block
	<-ctx.Done()
	return nil
}

func (s *TcpWorker) readFromConnection(conn net.Conn, reset_channel chan time.Duration) error {
	for {
		payload, err := internal.Decode(conn)
		if err != nil {
			return err
		}

		// We dont need to send a ping if we have just read something from the client
		reset_channel <- 0
		err = conn.SetDeadline(time.Now().Add(s.determineConnectionDeadline(defaultPingInterval)))
		s.respond(payload, conn)
	}
}

func (s *TcpWorker) respond(payload internal.Payload, conn net.Conn) {
	var response internal.String
	payload_type := payload.GetType()
	switch payload_type {
	case internal.InitializationPacketType:
		response = internal.String(
			"Hey Ho Sailor, got an InitializationPacketType after connection has been initialized, I don't know how to respond to this",
		)
	case internal.BinaryType:
		response = internal.String("Give me all them bytes!")
	case internal.StringType:
		response = internal.String("Give me all them strings!")
	default:
		s.log.Printf(
			"Received a payload of unknown type: %d, don't know how to respond \n",
			payload_type,
		)
		return
	}

	response.WriteTo(conn)
}

func (s *TcpWorker) determineConnectionDeadline(ping_timer time.Duration) time.Duration {
	if ping_timer < defaultPingInterval {
		return defaultPingInterval
	}

	return ping_timer + defaultPingInterval/2
}

func (s *TcpWorker) initializeConnection(conn net.Conn) (internal.InitializationPacket, error) {
	var initializationPacket internal.InitializationPacket
	for {
		payload, err := internal.Decode(conn)
		if err != nil {
			return initializationPacket, err
		}

		if payload.GetType() != internal.InitializationPacketType {
			return initializationPacket, errors.New(
				"The received packet is not of type InitializationPacketType",
			)
		}

		t, ok := payload.(*internal.InitializationPacket)
		if ok == true {
			return *t, nil
		}

		return initializationPacket, errors.New(
			"Could not cast the initial connection packet to type: InitializationPacket",
		)
	}
}

const defaultPingInterval = 30 * time.Second

func (s *TcpWorker) Ping(ctx context.Context, w io.Writer, reset chan time.Duration) {
	var interval time.Duration
	select {
	// If the context is done terminate the pinger
	case <-ctx.Done():
		return
	// Read out the interval from the reset channel
	case interval = <-reset:
	default:
		{
			// Default to the defaultPingInterval in case the value in the reset channel is invalid
			if interval <= 0 {
				interval = defaultPingInterval
			}
		}
	}

	timer := time.NewTimer(interval)
	// Once the method is about to exit drain the channel if it has been stopped
	defer func() {
		// timer.Stop() will return false if the timer has already been terminated, either directly by a call to
		// timer.Stop() by simply expiring. In the case that the timer has been terminated its channel is drained
		// to avoid leaking
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		// The select is a blocking operation and will not yield out of until one of the conditions has been met
		select {
		case <-ctx.Done():
			return

		// If a new reset is emitted via the reset channel read out the new value and update the interval.
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}

			if newInterval > 0 {
				interval = newInterval
			}

		// Read out the timer channel which will be available once the timer expires, if writing to the writer errors out abort
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}

		}
		// Once one of the conditions has happened reset the timer with the updated interval
		_ = timer.Reset(interval)
	}
}
