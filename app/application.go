package app

import (
	"fmt"
	"os"

	"authserver/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	api = gin.Default()
)

// StartApplication ...
func StartApplication() {
	//database.Seed()

	// Swagger 2.0 Meta Information
	docs.SwaggerInfo.Title = "Mobilesoft Authentication API"
	docs.SwaggerInfo.Description = "API for authenticating admins and clients"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	SetupRoutes()

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := api.Run(os.Getenv("HOST") + ":" + os.Getenv("PORT"))
	if err != nil {
		fmt.Println(err)
	}

}
