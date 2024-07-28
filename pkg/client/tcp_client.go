package client

import (
	"fmt"
	"log"
	"net"
	"os"
)

type TcpClient struct {
	addr string
	log  *log.Logger
}

func NewTcpClient(addr string) TcpClient {
	logger := log.New(
		os.Stdout,
		fmt.Sprintf("Client (%s): ", addr),
		log.LUTC,
	)
	return TcpClient{addr: addr, log: logger}
}

func (c *TcpClient) Connect() error {
	c.log.Printf("Connecting client to %q\n", c.addr)

	dial, err := net.Dial("tcp", c.addr)
	if err != nil {
		c.log.Printf("Could not connect to %q\n", c.addr)
		return err
	}

	defer dial.Close()
	go c.readFromConnection()
	go c.writeToConnection()
	return nil
}

func (c *TcpClient) readFromConnection() {

}

func (c *TcpClient) writeToConnection() {

}
