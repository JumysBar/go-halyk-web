package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	//создание нового инстанса сервера
	e := echo.New()

	// регистрация мидлварев для уровня root
	e.Use(middleware.Logger())

	//регистрация методов
	e.POST("/bindUser", bindUser)

	err := e.Start("127.0.0.1:8080")
	//исключаем ошибку ErrServerClosed, которая приходит при выключении сервера через вызов Echo.Shutdown
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err)
	}
}
