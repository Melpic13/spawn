package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// JWTAuthorizer validates basic bearer format.
type JWTAuthorizer struct{}

// Authorize validates JWT header presence.
func (JWTAuthorizer) Authorize(r *http.Request) error {
	tok := r.Header.Get("Authorization")
	if tok == "" {
		return fmt.Errorf("authorize jwt: missing authorization header")
	}
	if !strings.HasPrefix(tok, "Bearer ") {
		return fmt.Errorf("authorize jwt: invalid bearer token")
	}
	return nil
}
