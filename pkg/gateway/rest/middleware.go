package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIKeyMiddleware checks x-api-key against expected token.
func APIKeyMiddleware(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return next(c)
			}
			if c.Request().Header.Get("X-API-Key") != expected {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
			}
			return next(c)
		}
	}
}
