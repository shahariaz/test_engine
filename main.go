package main

import (
	"fmt"
	"jsonTodql/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router
	router := gin.Default()
	
	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		
		c.Next()
	})
	
	// Initialize handlers
	queryHandler := handlers.NewQueryHandler()
	
	// API routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", queryHandler.HealthCheck)
		
		// Schema information
		api.GET("/schema", queryHandler.GetSchema)
		
		// Main conversion endpoint
		api.POST("/convert", queryHandler.ConvertQuery)
	}
	
	// Root endpoint with API information
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":     "JSON to DQL Converter",
			"version":     "1.0.0",
			"description": "Converts dynamic JSON queries to Dgraph DQL format",
			"endpoints": map[string]interface{}{
				"health":  "GET /api/v1/health - Service health check",
				"schema":  "GET /api/v1/schema - Get available fields and operators",
				"convert": "POST /api/v1/convert - Convert JSON query to DQL",
			},
			"example_usage": map[string]interface{}{
				"url":    "/api/v1/convert",
				"method": "POST",
				"body": map[string]interface{}{
					"combine_with": "AND",
					"groups": []map[string]interface{}{
						{
							"combine_with": "OR",
							"filters": []map[string]interface{}{
								{"field": "age", "op": ">=", "value": 18},
								{"field": "country", "op": "IN", "value": []string{"USA", "Canada"}},
							},
						},
					},
				},
			},
		})
	})
	
	// Start server
	port := ":8080"
	fmt.Printf("🚀 JSON to DQL Converter API starting on port %s\n", port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   GET  /                    - API information")
	fmt.Println("   GET  /api/v1/health      - Health check")
	fmt.Println("   GET  /api/v1/schema      - Schema information")
	fmt.Println("   POST /api/v1/convert     - Convert JSON to DQL")
	fmt.Println()
	
	log.Fatal(router.Run(port))
}