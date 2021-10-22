package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

func (s *MyRESTServer) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

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

func (s *MyRESTServer) GetObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ObjID: %d. Name: %s. Meta: %s", s.Obj.ID, s.Obj.Name, s.Obj.Meta)
}

func (s *MyRESTServer) GetObjectName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Name: %s", s.Obj.Name)
}

func (s *MyRESTServer) ObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId := vars["id"]

	// Валидация параметров, но вероятность ошибки меньше, т.к. горилла позаботилась
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

func TimerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(rw, r)
		fmt.Printf("Request handling time: %v\n", time.Now().Sub(start))
	})
}

func HTTPMethodsCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			fmt.Fprintln(w, "Unsupported method")
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthorizationCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.FormValue("username")
		if user == "" {
			fmt.Fprintln(w, "Parameter 'user' not found")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		pass := r.FormValue("pass")
		if pass == "" {
			fmt.Fprintln(w, "Parameter 'pass' not found")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if user != "SuperAdmin" || pass != "SuperSecretPassword" {
			fmt.Fprintln(w, "Incorrect user or password")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Все в порядке

		ctx := context.WithValue(r.Context(), "username", user)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
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
	r := mux.NewRouter()
	r.HandleFunc("/api/user/{name:[a-zA-Z]+}", server.GetUser).Methods("GET")
	r.HandleFunc("/api/object", server.GetObject).Methods("GET")
	r.HandleFunc("/api/object/name", server.GetObjectName).Methods("GET")
	r.HandleFunc("/api/object/{id:[0-9]+}", server.ObjectHandler).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Unknown resourse")
		w.WriteHeader(http.StatusNotFound)
		return
	})
	r.Use(TimerMiddleware, HTTPMethodsCheckMiddleware, AuthorizationCheckMiddleware)
	http.ListenAndServe(":8080", r)
}
