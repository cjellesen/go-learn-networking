package server_test

import (
	"context"
	"go-learn-networking/internal"
	"go-learn-networking/pkg/server"
	"testing"
	"time"
)

func TestServerInitialization(t *testing.T) {
	addr := internal.NewAddrType("tcp", "127.0.0.1:4000")
	server := server.NewTcpServer(addr)

	ctx, cancel := context.WithCancel(context.Background())
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server")
	}
	time.Sleep(5 * time.Second)
	cancel()
}
