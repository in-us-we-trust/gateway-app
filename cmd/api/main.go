package main

import (
	_ "gateway-app/docs"
	"gateway-app/internal/client"
	"gateway-app/internal/config"
	"gateway-app/internal/handlers"
	"gateway-app/internal/middleware"
	"gateway-app/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Secure Home Gateway API
// @version         1.0
// @description     IoT gateway for the Secure Home project — routes microcontroller events to the detection service.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	cfg := config.Load()

	detectorClient := client.NewHTTPDetectorClient(
		cfg.DetectorServiceURL,
	)

	userClient := client.NewHTTPUserClient(
		cfg.UserServiceURL,
	)

	authClient := client.NewHTTPAuthClient(
		cfg.AuthServiceURL,
	)

	deviceService := service.NewDeviceService(
		userClient,
		detectorClient,
	)

	authService := service.NewAuthService(
		userClient,
		authClient,
	)

	deviceHandler := handlers.NewDeviceHandler(deviceService)
	authHandler := handlers.NewAuthHandler(authService)

	router := gin.Default()

	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := router.Group("/api/v1")
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth())

	{
		protected.GET("/me/device", deviceHandler.GetMyDevice)
		protected.POST("/me/device", deviceHandler.LinkMyDevice)
		protected.DELETE("/me/device", deviceHandler.UnlinkMyDevice)
		protected.GET("/me/device/events", deviceHandler.GetMyDeviceEvents)
		protected.POST("/me/push-token", deviceHandler.RegisterPushToken)
		protected.DELETE("/me/push-token/:appDeviceId", deviceHandler.DeletePushToken)
	}

	log.Printf("starting server on :%s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
