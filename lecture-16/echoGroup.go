package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// создание нового инстанса сервера
	e := echo.New()

	// регистрация мидлварев на уровне root
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// регистрация методов root
	e.GET("/health", health)
	e.GET("/panic", _panic)

	// создаём новую группу с префиксом /users и отдельным middleware
	userGroup := e.Group("/users")

	// регистрация мидлварев на уровне userGroup
	userGroup.Use(
		middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if username == "halyk" && password == "changeMe@123" {
				return true, nil
			}
			return false, nil
		}))

	// регистрация методов для userGroup
	userGroup.GET("/health", health)
	// а тут еще и middleware самописных накидываем в удобном для нам порядке
	userGroup.POST("/get/:name", genUser, timerMid, paramGetterMid)

	err := e.Start("127.0.0.1:8080")
	//исключаем ошибку ErrServerClosed, которая приходит при выключении сервера через вызов Echo.Shutdown
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err)
	}
}
