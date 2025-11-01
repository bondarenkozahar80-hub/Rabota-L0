package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"order-service/internal/domain"
	cache "order-service/internal/repository/ cache"
	"os"
	"os/signal"
	"syscall"
	"time"

	" order-service/internal/config"
	" order-service/internal/delivery/kafka"
	" order-service/internal/handler/http "
	" order-service/internal/repository/ cache"
	" order-service/internal/repository/postgres"
	" order-service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.Load ("./configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	orderRepo, err := postgres.NewOrderRepo(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer orderRepo.Close()

	// Initialize cache
	cache := cache.NewInMemoryCache()

	// Restore cache from database on startup
	ctx := context.Background()
	orders, err := orderRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Warning: Failed to restore cache from database: %v", err)
	} else {
		cacheMap := make(map[string]*domain.Order)
		for _, order := range orders {
			cacheMap[order.OrderUID] = order
		}
		cache.Restore(cacheMap)
		log.Printf("Restored %d orders to cache", len(orders))
	}

	// Initialize services
	orderService := service.NewOrderService(orderRepo, cache)

	// Initialize Kafka consumer
	kafkaConsumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, orderService)

	// Start Kafka consumer in background
	go kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Initialize HTTP handler
	orderHandler := http.NewOrderHandler(orderService)

	// Setup HTTP router
	router := mux.NewRouter()
	router.HandleFunc("/order/{id}", orderHandler.GetOrderByUID).Methods("GET")
	router.HandleFunc("/", orderHandler.ServeStatic).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
