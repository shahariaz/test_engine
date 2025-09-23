package main

import (
	"jsonTodql/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	queryHandler := handlers.NewQueryHandler()

	api := router.Group("/api/v1")
	{
		api.GET("/health", queryHandler.HealthCheck)
		api.GET("/schema", queryHandler.GetSchema)
		api.POST("/convert", queryHandler.ConvertQuery)
		api.POST("/execute", queryHandler.ExecuteQuery)
		api.POST("/validate", queryHandler.ValidateQuery)
		api.POST("/analyze", queryHandler.AnalyzeComplexity)
	}

	log.Fatal(router.Run(":8090"))
}
