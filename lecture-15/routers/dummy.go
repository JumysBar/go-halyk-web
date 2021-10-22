package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// 1. /api/user/$name - GET получить юзера
// 2. /api/object - GET - получить объект
// 3. /api/object/name - GET - получить имя объекта
// 4. /api/object/$id?meta=asdf - POST - изменить мету
func (s *MyRESTServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api") {
		fmt.Fprintf(w, "Incorrect request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

SWITCH:
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/user/"):
		// Проверка метода
		if r.Method != http.MethodGet {
			fmt.Fprintf(w, "Unsupported method")
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 4 { // Если url содержит меньше или больше частей, разделенных символом '/' - неподдерживаемый запрос
			break
		}
		name := parts[3]
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
	case strings.HasPrefix(r.URL.Path, "/api/object"):
		switch r.Method {
		case http.MethodGet:
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) == 3 {
				// Выдаем результат
				fmt.Fprintf(w, "ObjID: %d. Name: %s. Meta: %s", s.Obj.ID, s.Obj.Name, s.Obj.Meta)
				return
			}
			if len(parts) == 4 && parts[3] == "name" {
				// Выдаем результат
				fmt.Fprintf(w, "Name: %s", s.Obj.Name)
				return
			}
			// Ошибка
			fmt.Fprintf(w, "Incorrect request")
			w.WriteHeader(http.StatusBadRequest)
			return

		case http.MethodPost:
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) != 4 { // Если url содержит меньше или больше частей, разделенных символом '/' - неподдерживаемый запрос
				break SWITCH
			}

			// Валидация параметров
			id, err := strconv.Atoi(parts[3])
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

	}
	// Дефолтная ошибка
	fmt.Fprintf(w, "Incorrect request")
	w.WriteHeader(http.StatusBadRequest)
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
	http.ListenAndServe(":8080", server)

}
