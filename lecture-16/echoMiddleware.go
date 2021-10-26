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

	// регистрация мидлварев
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//регистрация методов
	e.GET("/health", health)
	e.GET("/panic", _panic)

	// а тут еще и middleware самописных накидываем в удобном для нам порядке
	e.POST("/users/get/:name", genUser, timerMid, paramGetterMid)

	err := e.Start("127.0.0.1:8080")
	//исключаем ошибку ErrServerClosed, которая приходит при выключении сервера через вызов Echo.Shutdown
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err)
	}
}
