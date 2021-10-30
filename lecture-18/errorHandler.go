package main

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func main() {
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		// to skip multiple calls of error handling
		if c.Response().Committed {
			return
		}
		// may return JSON serialize error
		handleError(err, c)
		// Call the default handler to return the HTTP response
		e.DefaultHTTPErrorHandler(err, c)
	}

	v1 := e.Group("/v1")

	v1.GET("/eof", eof)
	v1.POST("/buffer", buffer)
	v1.POST("/ok", ok)

	_ = e.Start(":8080")
}

// handleError - central error handler
func handleError(serviceError error, c echo.Context) {

	// handle errors from usecases layer
	if jsonErr := setResponseFromUseCaseError(serviceError, c); jsonErr != nil {
		log.Error().Err(jsonErr).Msg("JSON serializing")
	}
}

func setResponseFromUseCaseError(serviceError error, c echo.Context) error {
	var jsonErr error

	if errors.Is(serviceError, io.EOF) {
		jsonErr = c.JSON(http.StatusInsufficientStorage, "error: "+serviceError.Error())
	} else if errors.Is(serviceError, io.ErrShortBuffer) {
		jsonErr = c.JSON(http.StatusInsufficientStorage, "error: "+serviceError.Error())
	}

	return jsonErr
}

func eof(c echo.Context) error {
	return io.EOF
}

func buffer(c echo.Context) error {
	return echo.NewHTTPError(http.StatusBadRequest, io.ErrShortBuffer)
}

func ok(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
