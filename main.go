// Package main is the entry point for the user management API.
package main

import (
	"log"
	"net/http"
	"userapi/db"
	"userapi/handler"
)

// main initializes the database, creates the HTTP handler, registers routes,
// and starts the API server on port 8080.
func main() {
	database, err := db.Open("users.db")
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer database.Close()

	h := &handler.Handler{DB: database}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", h.ListUsers)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users/{id}", h.GetUser)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
	mux.HandleFunc("POST /login", h.Login)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
