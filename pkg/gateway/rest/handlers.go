package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/v1/agents", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"agents": []interface{}{}})
	})
}
