package server

import (
	"context"
	"errors"
	"go-learn-networking/internal"
	"io"
	"net"
	"time"
)

type TcpServer struct {
	baseServer internal.Server
}

func NewTcpServer(addr net.Addr) TcpServer {
	return TcpServer{baseServer: internal.NewServer(addr)}
}

func (s *TcpServer) Start(ctx context.Context) error {
	s.baseServer.Logger.Printf("Starting server listening on %q\n", s.baseServer.AddrType.String())
	lsn, err := net.Listen(s.baseServer.AddrType.Network(), s.baseServer.AddrType.String())
	if err != nil {
		s.baseServer.Logger.Printf("Could not open a connection, failed with error:\n%q", err)
		return err
	}

	// At the moment there is no graceful shutdown of the listener, which should be corrected
	s.baseServer.Lsn = lsn
	go s.accepLoop(ctx)
	s.baseServer.Logger.Println("Server is actively accepting connections")
	return nil
}

func (s *TcpServer) Terminate() error {
	return s.baseServer.Terminate()
}

func (s *TcpServer) accepLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.Terminate()
			s.baseServer.Logger.Println("The listener has been terminated")
		default:
			s.acceptConnections()
		}
	}
}

func (s *TcpServer) acceptConnections() {
	conn, err := s.baseServer.Lsn.Accept()
	if err != nil {
		s.baseServer.Logger.Printf(
			"Failed to accept incomming connection, failed with error:\n%q",
			err,
		)
		return
	}

	go s.processConnection(conn)
}

func (s *TcpServer) processConnection(conn net.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		conn.Close()
	}()
	initializationPacket, err := s.initializeConnection(conn)
	if err != nil {
		return err
	}

	reset_channel := make(chan time.Duration, 1)
	reset_interval := time.Millisecond * time.Duration(initializationPacket.PingIntervalMs)
	reset_channel <- reset_interval

	go s.Ping(ctx, conn, reset_channel)
	err = conn.SetDeadline(time.Now().Add(s.determineConnectionDeadline(reset_interval)))
	if err != nil {
		return err
	}

	go s.readFromConnection(conn, reset_channel)

	// If I remember correctly this should block
	<-ctx.Done()
	return nil
}

func (s *TcpServer) readFromConnection(conn net.Conn, reset_channel chan time.Duration) {
	for {
		payload, err := internal.Decode(conn)
		if err != nil {
			s.baseServer.Logger.Printf("Failed to decode message, failed with %s\n", err.Error())
		}

		// We dont need to send a ping if we have just read something from the client
		reset_channel <- 0
		err = conn.SetDeadline(time.Now().Add(s.determineConnectionDeadline(defaultPingInterval)))
		s.respond(payload, conn)
	}
}

func (s *TcpServer) respond(payload internal.Payload, conn net.Conn) {
	var response string
	payload_type := payload.GetType()
	switch payload_type {
	case internal.InitializationPacketType:
		response = "Hey Ho Sailor, go an InitializationPacketType after connection has been initialized, I don't know how to respond to this"
	case internal.BinaryType:
		response = "Give me all them bytes!"
	case internal.StringType:
		response = "Give me all them strings!"
	default:
		s.baseServer.Logger.Printf(
			"Received a payload of unknown type: %d, don't know how to respond \n",
			payload_type,
		)
		return
	}

	conn.Write([]byte(response))
}

func (s *TcpServer) determineConnectionDeadline(ping_timer time.Duration) time.Duration {
	if ping_timer < defaultPingInterval {
		return defaultPingInterval
	}

	return ping_timer + defaultPingInterval/2
}

func (s *TcpServer) initializeConnection(conn net.Conn) (internal.InitializationPacket, error) {
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

func (s *TcpServer) Ping(ctx context.Context, w io.Writer, reset chan time.Duration) {
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
