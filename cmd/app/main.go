package main

import (
	"context"
	"crud-go/internal/handler"
	"crud-go/internal/middleware"
	"crud-go/internal/repository"
	"crud-go/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	dbConn, connErr := repository.Connect()
	if connErr != nil {
		log.Fatal(connErr)
	}
	defer repository.Close(dbConn)

	basketRepository, basketRepositoryCreationErr := repository.NewBasketRepository(dbConn)
	if basketRepositoryCreationErr != nil {
		log.Fatal(basketRepositoryCreationErr)
	}
	itemRepository, itemRepositoryCreationErr := repository.NewItemRepository(dbConn)
	if itemRepositoryCreationErr != nil {
		log.Fatal(itemRepositoryCreationErr)
	}
	orderRepository, orderRepositoryCreationErr := repository.NewOrderRepository(dbConn)
	if orderRepositoryCreationErr != nil {
		log.Fatal(orderRepositoryCreationErr)
	}
	taskRepository, taskRepositoryCreationErr := repository.NewTaskRepository(dbConn)
	if taskRepositoryCreationErr != nil {
		log.Fatal(taskRepositoryCreationErr)
	}
	tokenRepository, tokenRepositoryCreationErr := repository.NewTokenRepository(dbConn)
	if tokenRepositoryCreationErr != nil {
		log.Fatal(tokenRepositoryCreationErr)
	}
	userRepository, userRepositoryCreationErr := repository.NewUserRepository(dbConn)
	if userRepositoryCreationErr != nil {
		log.Fatal(userRepositoryCreationErr)
	}

	accessLT, convErr := strconv.Atoi(os.Getenv("AUTH_ACCESS_LIFETIME"))
	refreshLT, convErr := strconv.Atoi(os.Getenv("AUTH_REFRESH_LIFETIME"))
	if convErr != nil {
		log.Fatal(convErr)
	}

	authService := service.NewAuthService(userRepository, tokenRepository, accessLT, refreshLT, os.Getenv("AUTH_ACCESS_KEY"), os.Getenv("AUTH_REFRESH_KEY"))
	customerService := service.NewCustomerService(&orderRepository, &taskRepository, basketRepository, itemRepository)
	managerService := service.NewManagerService(&orderRepository, &taskRepository, userRepository)
	workerService := service.NewWorkerService(&orderRepository, &taskRepository)

	authHandler := handler.NewAuthHandler(authService)
	customerHandler := handler.NewCustomerHandler(customerService)
	managerHandler := handler.NewManagerHandler(managerService)
	workerHandler := handler.NewWorkerHandler(workerService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	router := gin.Default()
	customerGroup := router.Group("/api/v1/customer")
	workerGroup := router.Group("api/v1/worker")
	managerGroup := router.Group("api/v1/manager")
	authGroup := router.Group("api/v1/auth")

	customerGroup.Use(authMiddleware.AuthMiddlewareFunc)
	workerGroup.Use(authMiddleware.AuthMiddlewareFunc)
	managerGroup.Use(authMiddleware.AuthMiddlewareFunc)

	customerGroup.GET("/orders", customerHandler.GetAllOrders)
	customerGroup.GET("/orders/:id", customerHandler.GetCustomerOrder)
	customerGroup.POST("/orders", customerHandler.CreateOrder)
	customerGroup.DELETE("orders/:id", customerHandler.DeleteOrder)
	customerGroup.GET("/items", customerHandler.GetItems)
	customerGroup.GET("/basket", customerHandler.GetBasket)
	customerGroup.POST("/basket", customerHandler.SaveToBasket)
	customerGroup.DELETE("/basket/:id", customerHandler.DeleteFromBasket)

	workerGroup.GET("/orders", workerHandler.GetOrdersByWorker)
	workerGroup.GET("/orders/:id", workerHandler.GetOrder)
	workerGroup.POST("/tasks/:id", workerHandler.SetTaskCompleted)

	managerGroup.GET("/orders", managerHandler.GetOrdersByManager)
	managerGroup.POST("/orders/:id", managerHandler.AssignWorkerToOrder)
	managerGroup.GET("/orders/:id", managerHandler.GetOrderByID)
	managerGroup.GET("/workers", managerHandler.GetAllWorkers)

	authGroup.POST("/newCustomer", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout)

	server := &http.Server{
		Addr:    ":" + os.Getenv("APP_PORT"),
		Handler: router,
	}

	go func() {
		log.Println("Server successfully started at port " + server.Addr)
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	server.Shutdown(ctx)
	log.Println("Server successfully shot down")

}
