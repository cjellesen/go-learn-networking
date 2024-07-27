package client

import (
	"go-learn-networking/internal"
	"net"
)

type TcpClient struct {
	baseClient internal.Client
}

func NewTcpClient(addr net.Addr) TcpClient {
	return TcpClient{baseClient: internal.NewClient(addr)}
}

func (c *TcpClient) Connect() error {
	c.baseClient.Logger.Printf("Connecting client to %q\n", c.baseClient.AddrType.String())

	dial, err := net.Dial(c.baseClient.AddrType.Network(), c.baseClient.AddrType.String())
	if err != nil {
		c.baseClient.Logger.Printf("Could not connect to %q\n", c.baseClient.AddrType.String())
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
