package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type Object struct {
	ID   int64
	Name string
	Meta string
}

type User struct {
	FirstName string
	Age       float64
	Married   bool
}

type MyRESTServer struct {
	ServerUser *User
	Obj        *Object
}

func (s *MyRESTServer) GetUser(ctx *fasthttp.RequestCtx) {
	name := ctx.UserValue("name").(string)

	// Типа поиск юзера по имени
	if name == s.ServerUser.FirstName {
		ctx.WriteString(fmt.Sprintf("Name: %s. Age: %f. Married: %v", s.ServerUser.FirstName, s.ServerUser.Age, s.ServerUser.Married))
		return
	}

	// Ошибка
	ctx.WriteString(fmt.Sprintf("User %s not found", name))
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	return
}

func (s *MyRESTServer) GetObject(ctx *fasthttp.RequestCtx) {
	ctx.WriteString(fmt.Sprintf("ObjID: %d. Name: %s. Meta: %s", s.Obj.ID, s.Obj.Name, s.Obj.Meta))
}

func (s *MyRESTServer) GetObjectName(ctx *fasthttp.RequestCtx) {
	ctx.WriteString(fmt.Sprintf("Name: %s", s.Obj.Name))
}

func (s *MyRESTServer) ObjectHandler(ctx *fasthttp.RequestCtx) {
	strId := ctx.UserValue("id").(string)

	// Валидация параметров, но вероятность ошибки меньше, т.к. горилла позаботилась
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.WriteString(fmt.Sprintf("ID is not integer"))
		return
	}

	// Типа поиск по ID
	if id != int(s.Obj.ID) {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.WriteString(fmt.Sprintf("Obj with ID %d not found", id))
		return
	}

	meta := ctx.FormValue("meta")
	s.Obj.Meta = string(meta)
	ctx.WriteString("Success")
	return
}

func TimerMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		fmt.Printf("Request handling time: %v\n", time.Now().Sub(start))
	}
}

func AuthorizationCheckMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		user := string(ctx.FormValue("username"))
		if user == "" {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.WriteString(fmt.Sprintf("Parameter 'user' not found"))
			return
		}

		pass := string(ctx.FormValue("pass"))
		if pass == "" {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.WriteString(fmt.Sprintf("Parameter 'pass' not found"))
			return
		}

		if user != "SuperAdmin" || pass != "SuperSecretPassword" {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.WriteString(fmt.Sprintf("Incorrect user or password"))
			return
		}

		// Все в порядке

		ctx.SetUserValue("username", user)
		next(ctx)
	}
}

func main() {
	server := &MyRESTServer{
		ServerUser: &User{
			FirstName: "Vladimir",
			Age:       23,
			Married:   true,
		},
		Obj: &Object{
			ID:   1,
			Name: "SomeObject",
			Meta: "SomeMeta",
		},
	}
	r := fasthttprouter.New()
	r.GET("/api/user/:name", TimerMiddleware(AuthorizationCheckMiddleware(server.GetUser)))
	r.GET("/api/object", TimerMiddleware(AuthorizationCheckMiddleware(server.GetObject)))
	r.GET("/api/object/name", TimerMiddleware(AuthorizationCheckMiddleware(server.GetObjectName)))
	r.POST("/api/object/{id:[0-9]+}", TimerMiddleware(AuthorizationCheckMiddleware(server.ObjectHandler)))
	r.GET("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("Unknown resources")
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	})
	fasthttp.ListenAndServe(":8080", r.Handler)
}
