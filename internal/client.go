package internal

import (
	"fmt"
	"log"
	"net"
	"os"
)

type Client struct {
	AddrType net.Addr
	Logger   *log.Logger
}

func NewClient(addr net.Addr) Client {
	logger := log.New(
		os.Stdout,
		fmt.Sprintf("Constructing a new server of type: %s\n", addr.Network()),
		log.LUTC,
	)

	return Client{
		AddrType: addr,
		Logger:   logger,
	}
}
