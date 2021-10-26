package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	//создание нового инстанса сервера
	e := echo.New()

	//регистрация методов
	e.GET("/health", health)
	e.POST("/health", health)
	e.PATCH("/health", health)

	err := e.Start("127.0.0.1:8080")
	//исключаем ошибку ErrServerClosed, которая приходит при выключении сервера через вызов Echo.Shutdown
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err)
	}
}
