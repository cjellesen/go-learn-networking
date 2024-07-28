package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
)

type TcpServer struct {
	addr      string
	logger    *log.Logger
	lsn       net.Listener
	MaxConn   int
	connQueue chan net.Conn
}

func NewTcpServer(addr string, maxConn int) TcpServer {
	logger := log.New(
		os.Stdout,
		fmt.Sprintf("Server (%s): ", addr),
		log.LUTC,
	)
	return TcpServer{
		addr:      addr,
		logger:    logger,
		MaxConn:   maxConn,
		connQueue: make(chan net.Conn, maxConn),
	}
}

func (s *TcpServer) Start(ctx context.Context) error {
	s.logger.Printf("Starting server listening on %q\n", s.addr)
	lsn, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Printf("Could not open a connection, failed with error:\n%q", err)
		return err
	}

	s.spinUpWorkers(ctx)

	s.lsn = lsn
	go s.acceptConnections()
	go s.terminationLoop(ctx)
	s.logger.Println("Server is actively accepting connections")
	return nil
}

func (s *TcpServer) spinUpWorkers(ctx context.Context) {
	for range s.MaxConn {
		worker := NewTcpWorker(s.connQueue)
		go worker.Start(ctx)
	}
}

func (s *TcpServer) terminationLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.terminate()
			return
		default:
			continue
		}
	}
}

// Any connection that is being worked by the TcpWorker(s) will not be shutdown by this. These connection will be
// termated by the worker itself when either the work is done or the connection time out runs out
func (s *TcpServer) terminate() {
	if err := s.lsn.Close(); err != nil {
		s.logger.Printf("Failed to close the listener, failed with error:\n%s", err)
	}

	close(s.connQueue)
	for conn := range s.connQueue {
		if err := conn.Close(); err != nil {
			s.logger.Printf("Failed to close the enqueued connection, failed with error:\n%s", err)
		}
	}
}

func (s *TcpServer) acceptConnections() {
	conn, err := s.lsn.Accept()
	if err != nil {
		s.logger.Printf(
			"Failed to accept incomming connection, failed with error:\n%q",
			err,
		)
		return
	}

	s.connQueue <- conn
}
