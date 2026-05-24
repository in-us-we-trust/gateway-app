package main

import (
	_ "gateway-app/docs"
	"gateway-app/internal/client"
	"gateway-app/internal/config"
	"gateway-app/internal/handler"
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

	userService := service.NewUserService(
		userClient,
		detectorClient,
	)

	userHandler := handler.NewUserHandler(userService)

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

	api.Use(middleware.JWTAuth())

	{
		api.GET("/me/device", userHandler.GetMyDevice)
		api.POST("/me/device", userHandler.LinkMyDevice)
		api.DELETE("/me/device", userHandler.UnlinkMyDevice)
		api.GET("/me/device/events", userHandler.GetMyDeviceEvents)
	}

	log.Printf("starting server on :%s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
