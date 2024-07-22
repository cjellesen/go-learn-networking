package internal

import (
	"fmt"
	"log"
	"net"
	"os"
)

type Server struct {
	AddrType net.Addr
	Logger   *log.Logger
	Lsn      net.Listener
}

func NewServer(addr net.Addr) Server {
	logger := log.New(
		os.Stdout,
		fmt.Sprintf("Constructing a new server of type: %s\n", addr.Network()),
		log.LUTC,
	)
	return Server{
		AddrType: addr,
		Logger:   logger,
	}
}
