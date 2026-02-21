//go:build integration

package e2e

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	restgw "spawn.dev/pkg/gateway/rest"
)

func TestRESTGatewayHealthEndpoint(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ephemeral port: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close ephemeral listener: %v", err)
	}

	server := restgw.New(addr)
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("start rest gateway: %v", err)
	}
	defer func() {
		_ = server.Stop(context.Background())
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(5 * time.Second)
	url := fmt.Sprintf("http://%s/healthz", addr)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("health endpoint did not return 200 in time, last err: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
