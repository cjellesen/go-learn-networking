package server_test

import (
	"context"
	"go-learn-networking/internal"
	"go-learn-networking/pkg/server"
	"net"
	"testing"
	"time"
)

func TestServerInitialization(t *testing.T) {
	server := server.NewTcpServer("127.0.0.1:5000", 5)
	ctx, cancel := context.WithCancel(context.Background())
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server")
	}
	time.Sleep(3 * time.Second)
	cancel()
}

func TestServerWithClient(t *testing.T) {
	server := server.NewTcpServer("127.0.0.1:2000", 5)
	ctx, cancel := context.WithCancel(context.Background())
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server")
	}

	client_conn, err := net.Dial("tcp", "127.0.0.1:2000")
	defer client_conn.Close()
	if err != nil {
		t.Fatalf("Failed to start client")
	}

	init_packet := internal.InitializationPacket{PingIntervalMs: 300, NRetries: 10}
	init_packet.WriteTo(client_conn)
	msg := internal.String("this is a test")
	msg.WriteTo(client_conn)

	time.Sleep(5 * time.Second)
	payload, err := internal.Decode(client_conn)
	if err != nil {
		t.Fatalf("Failed to decode message from server, failed with error: %s\n", err)
	}

	t.Logf(payload.String())
	time.Sleep(5 * time.Second)
	cancel()
}
