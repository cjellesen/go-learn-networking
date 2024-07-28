package server_test

import (
	"context"
	"go-learn-networking/pkg/server"
	"testing"
	"time"
)

func TestServerInitialization(t *testing.T) {
	server := server.NewTcpServer("127.0.0.1:4000", 5)
	ctx, cancel := context.WithCancel(context.Background())
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server")
	}
	time.Sleep(5 * time.Second)
	cancel()
}
