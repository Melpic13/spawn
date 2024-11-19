package auth

import "net/http"

// Authorizer authorizes HTTP requests.
type Authorizer interface {
	Authorize(r *http.Request) error
}
