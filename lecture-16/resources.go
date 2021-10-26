package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// User структура для общения
type User struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Children []User `json:"children,omitempty"`

	secret string
}

// genUser хендлер для генерации пользователя и ответа в виде XML или JSON
func genUser(c echo.Context) error {
	name := c.Param("name")
	secret := c.QueryParam("secret")

	u := User{
		Name: name + "-батя",
		Age:  35,
		Children: []User{
			{
				Name: name + "-младший",
				Age:  3,
			},
			{
				Name: name + "-средний",
				Age:  10,
			},
		},
		secret: secret,
	}

	return c.JSON(http.StatusOK, u)
	//return c.XML(http.StatusOK, u)
}

// health простейший хендлер, который говорит нам что сервер все еще жив
func health(c echo.Context) error {
	//возвращаем в качестве ответа статус 200 OK без содержимого
	return c.NoContent(http.StatusOK)
}

// _panic намеренный вызов паники
func _panic(c echo.Context) error {
	panic("PANIC-PANIC-PANIC")
}

// paramGetterMid - middleware для примера. Получаем данные из Query Params
func paramGetterMid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("QUERY PARAMS: %+v\n", c.QueryParams())
		time.Sleep(time.Second)
		return next(c)
	}
}

// timerMid - middleware для примера. Оборачиваем вызов и смотрим на его время выполнения.
func timerMid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		t := time.Now()
		_ = next(c)
		fmt.Println(time.Now().Sub(t))
		return nil
	}
}

// bindUser - пример сериализации и десириализации JSON объекта
func bindUser(c echo.Context) error {
	u := new(User)

	if err := c.Bind(u); err != nil {
		return fmt.Errorf("ow we fkn lost man (c) sing-sing")
	}

	fmt.Printf("Binded user: %+v", u)

	if err := validate(u); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"message": err.Error()},
		)
	}

	u.Name = "changed " + u.Name
	return c.JSON(http.StatusOK, u)
}

func validate(u *User) error {
	if u.Name == "" {
		return fmt.Errorf("user name must set")
	}

	if !(u.Age > 0) {
		return fmt.Errorf("user age must be positive number")
	}
	return nil
}
