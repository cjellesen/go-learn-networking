package server

import (
	"go-learn-networking/internal"
	"net"
)

type TcpServer struct {
	baseServer internal.Server
}

func NewTcpServer(addr net.Addr) *TcpServer {
	return &TcpServer{baseServer: internal.NewServer(addr)}
}

func (s *TcpServer) Start() error {
	s.baseServer.Logger.Printf("Starting server listening on %q\n", s.baseServer.AddrType.String())
	lsn, err := net.Listen(s.baseServer.AddrType.Network(), s.baseServer.AddrType.String())
	if err != nil {
		s.baseServer.Logger.Printf("Could not open a connection, failed with error:\n%q", err)
		return err
	}

	s.baseServer.Lsn = lsn
	go s.acceptConnections()
	s.baseServer.Logger.Println("Server is actively accepting connections")
	return nil
}

func (s *TcpServer) acceptConnections() {
	for {
		conn, err := s.baseServer.Lsn.Accept()
		if err != nil {
			s.baseServer.Logger.Printf(
				"Failed to accept incomming connection, failed with error:\n%q",
				err,
			)
		}

		go s.readFromConnection(conn)
	}
}

func (s *TcpServer) readFromConnection(conn net.Conn) {
	defer conn.Close()

}
