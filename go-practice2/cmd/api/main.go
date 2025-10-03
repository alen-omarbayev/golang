package main

import (
	"fmt"
	"go-practice2/internal/handlers"
	"go-practice2/internal/middleware"
	"net/http"
)

type general string

func (gn general) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Practice2 application")
}

func main() {
	router := http.NewServeMux()
	var gn general

	router.Handle("/", gn)
	router.Handle("/user", middleware.AuthMiddleware(http.HandlerFunc(handlers.UserHandler)))
	router.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "auth page")
	})

	server := http.Server{
		Addr:    ":8070",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server staring error: ", err)
	}
}
