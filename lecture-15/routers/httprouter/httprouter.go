package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
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

func (s *MyRESTServer) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")
	// Валидация имени
	if strings.ContainsAny(name, "0123456789-=@!#$%^&*") {
		fmt.Fprintf(w, "Name contains invalid characters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Типа поиск юзера по имени
	if name == s.ServerUser.FirstName {
		fmt.Fprintf(w, "Name: %s. Age: %f. Married: %v", s.ServerUser.FirstName, s.ServerUser.Age, s.ServerUser.Married)
		return
	}

	// Ошибка
	fmt.Fprintf(w, "User %s not found", name)
	w.WriteHeader(http.StatusNotFound)
	return
}

func (s *MyRESTServer) GetObject(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "ObjID: %d. Name: %s. Meta: %s", s.Obj.ID, s.Obj.Name, s.Obj.Meta)
}

func (s *MyRESTServer) GetObjectName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Name: %s", s.Obj.Name)
}

func (s *MyRESTServer) ObjectHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	strId := p.ByName("id")

	// Валидация параметров
	id, err := strconv.Atoi(strId)
	if err != nil {
		fmt.Fprintf(w, "ID is not integer")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Типа поиск по ID
	if id != int(s.Obj.ID) {
		fmt.Fprintf(w, "Obj with ID %d not found", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	meta := r.FormValue("meta")
	s.Obj.Meta = meta
	fmt.Fprintln(w, "Success")
	return
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
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprintf(w, "Unknown resourse")
		w.WriteHeader(http.StatusNotFound)
		return
	})
	router.GET("/api/user/:name", server.GetUser)
	router.GET("/api/object", server.GetObject)
	router.GET("/api/object/name", server.GetObjectName)
	router.POST("/api/object/:id", server.ObjectHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
