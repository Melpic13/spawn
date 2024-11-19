package auth

import (
	"fmt"
	"net/http"
)

// APIKeyAuthorizer validates API keys.
type APIKeyAuthorizer struct {
	Header string
	Key    string
}

// Authorize validates the configured API key.
func (a APIKeyAuthorizer) Authorize(r *http.Request) error {
	header := a.Header
	if header == "" {
		header = "X-API-Key"
	}
	if a.Key == "" {
		return nil
	}
	if r.Header.Get(header) != a.Key {
		return fmt.Errorf("authorize api key: invalid key")
	}
	return nil
}
