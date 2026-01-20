package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p2nova-vpn/internal/config"
	"p2nova-vpn/internal/handler"
	"p2nova-vpn/internal/middleware"
	"p2nova-vpn/internal/repository"
	"p2nova-vpn/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository()
	serverRepo := repository.NewServerRepository()

	// Initialize services
	wgService := service.NewWireguardService(cfg)
	vpnService := service.NewVPNService(sessionRepo, wgService, cfg)
	serverService := service.NewServerService(serverRepo, cfg)

	// Initialize handlers
	h := handler.NewHandler(vpnService, serverService)

	// Setup router
	r := mux.NewRouter()

	// Middleware
	r.Use(middleware.CORS)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	// Routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/servers", h.GetServers).Methods("GET")
	api.HandleFunc("/vpn/server", h.SelectServer).Methods("POST")
	api.HandleFunc("/vpn/connect", h.Connect).Methods("POST")
	api.HandleFunc("/vpn/disconnect", h.Disconnect).Methods("POST")
	api.HandleFunc("/vpn/speed", h.GetSpeed).Methods("GET")
	api.HandleFunc("/vpn/status", h.GetStatus).Methods("GET")
	api.HandleFunc("/health", h.Health).Methods("GET")

	// Server setup
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("p2Nova VPN API starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
