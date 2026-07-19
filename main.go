package main

import (
	"fmt"
	"net/http"

	"github.com/leventkok/mlc-llm-monitoring/internal/handlers"
	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
)





func main(){
	store := storage.NewMemoryStore()
	authHandler := handlers.NewAuthHandler(store)


	http.HandleFunc("/health", handlers.Health)
	http.HandleFunc("/config", handlers.Config)
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/me", middleware.RequireAuth(authHandler.Me))
	http.HandleFunc("/auth/logout", authHandler.Logout)
	http.HandleFunc("/auth/refresh", middleware.RequireAuth(authHandler.Refresh))
	http.HandleFunc("/auth/change-password", middleware.RequireAuth(authHandler.ChangePassword))
	http.HandleFunc("/auth/login", authHandler.Login)
	fmt.Println("Server started: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}