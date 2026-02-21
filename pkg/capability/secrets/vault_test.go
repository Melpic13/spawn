package secrets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVaultResolverKV2(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"data":{"password":"secret-value"}}}`))
	}))
	defer srv.Close()

	resolver := VaultResolver{Address: srv.URL, Token: "test-token", Client: srv.Client()}
	val, err := resolver.Resolve(context.Background(), "vault://secret/data/app#password")
	if err != nil {
		t.Fatalf("resolve vault secret: %v", err)
	}
	if val != "secret-value" {
		t.Fatalf("expected secret-value, got %q", val)
	}
}
