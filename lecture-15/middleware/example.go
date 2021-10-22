package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("username").(string)
	fmt.Fprintln(w, "Super secret information for %s", user)
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
	final := AuthorizationCheckMiddleware(http.HandlerFunc(Handler))
	final = HTTPMethodsCheckMiddleware(final)
	final = TimerMiddleware(final)

	http.Handle("/", final)
	http.ListenAndServe(":8080", nil)
}
