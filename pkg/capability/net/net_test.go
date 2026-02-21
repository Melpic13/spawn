package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"spawn.dev/pkg/capability"
)

func TestHTTPAllowAndDenyRules(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	cap := New([]string{"127.0.0.1"}, nil)
	resp, err := cap.Execute(context.Background(), &capability.Request{Action: "get", Params: map[string]interface{}{"url": srv.URL}})
	if err != nil {
		t.Fatalf("execute get: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected allowed request, got %#v", resp.Error)
	}

	cap = New([]string{"*"}, []string{"127.0.0.1"})
	resp, err = cap.Execute(context.Background(), &capability.Request{Action: "get", Params: map[string]interface{}{"url": srv.URL}})
	if err != nil {
		t.Fatalf("execute denied get: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected denied request")
	}
}
