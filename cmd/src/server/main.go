package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-management-service/internal/config"
	"order-management-service/internal/controller"
	"order-management-service/internal/middleware"
	"order-management-service/internal/repository"
	"order-management-service/internal/routes"
	"order-management-service/internal/service"
	"order-management-service/internal/service/rabbitmq"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Router   *mux.Router
	DB       *config.Database
	Config   *config.Config
	Logger   *zap.Logger
	Broker   service.MessageBroker
	Consumer rabbitmq.Consumer
}

func main() {
	server()
}

func server() {
	app := initializeApp()
	defer func() {
		if app.Broker != nil {
			app.Broker.Close()
		}
	}()

	// Start RabbitMQ Consumer
	if app.Consumer != nil {
		if err := app.Consumer.Consume(context.Background()); err != nil {
			app.Logger.Error("Failed to start consumer", zap.Error(err))
		}
	}

	addr := fmt.Sprintf(":%s", app.Config.AppPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		app.Logger.Info("Server starting", zap.String("port", app.Config.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Could not start server", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	app.Logger.Info("Server exiting")
}

func initializeApp() *App {
	cfg := config.LoadConfig()

	// Logger
	logger, _ := zap.NewProduction()

	// Database
	dbService := &config.DBService{}
	dsn := config.GetDSN(cfg)
	db, err := dbService.EstablishPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	dbObj := &config.Database{Instance: db}

	// Auto Migrate
	if cfg.EnableMigration {
		repository.RunMigrations(dbObj.Instance)
	}

	// RabbitMQ
	broker, err := rabbitmq.NewRabbitMQBroker(cfg)
	if err != nil {
		logger.Warn("RabbitMQ not connected", zap.Error(err))
	}

	// Repositories (Inject Logger)
	userRepo := repository.NewUserRepository(dbObj.Instance, logger)
	productRepo := repository.NewProductRepository(dbObj.Instance, logger)
	orderRepo := repository.NewOrderRepository(dbObj.Instance, logger)

	// Services (Inject Logger)
	userSvc := service.NewUserService(userRepo, cfg, logger)
	productSvc := service.NewProductService(productRepo, logger)
	commSvc := service.NewCommunicationService(cfg, logger)
	addressSvc := service.NewAddressService(userRepo, logger)
	activitySvc := service.NewOrderActivityService(userRepo, orderRepo, commSvc, logger)
	orderSvc := service.NewOrderService(orderRepo, productRepo, userRepo, broker, logger)

	// Controllers (Inject Logger)
	userCtrl := controller.NewUserController(userSvc, logger)
	productCtrl := controller.NewProductController(productSvc, logger)
	addressCtrl := controller.NewAddressController(addressSvc, logger)
	orderCtrl := controller.NewOrderController(orderSvc, logger)

	consumer, err := rabbitmq.NewRabbitMQConsumer(cfg, logger, activitySvc)
	if err != nil {
		logger.Warn("RabbitMQ Consumer not connected", zap.Error(err))
	}

	// Router & Middlewares
	r := mux.NewRouter()
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.SecurityHeaders)

	rateLimiter := middleware.NewRateLimiter()
	r.Use(rateLimiter.Middleware)

	r.Use(CORS)

	AppRoutes(r, userCtrl, productCtrl, orderCtrl, addressCtrl, userRepo, cfg)

	return &App{
		Router:   r,
		DB:       dbObj,
		Config:   cfg,
		Logger:   logger,
		Broker:   broker,
		Consumer: consumer,
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AppRoutes(r *mux.Router, userCtrl controller.UserController, productCtrl controller.ProductController, orderCtrl controller.OrderController, addrCtrl controller.AddressController, userRepo repository.UserRepository, cfg *config.Config) {
	api := r.PathPrefix("/api/v1").Subrouter()

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, userRepo)

	routes.RegisterUserRoutes(api, userCtrl, addrCtrl, authMiddleware)
	routes.RegisterProductRoutes(api, productCtrl)
	routes.RegisterOrderRoutes(api, orderCtrl, authMiddleware)
}
