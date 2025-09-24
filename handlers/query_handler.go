package handlers

import (
	"context"
	"fmt"
	"jsonTodql/config"
	"jsonTodql/converter"
	"jsonTodql/dgraph"
	"jsonTodql/models"
	"jsonTodql/validation"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// QueryHandler handles the JSON to DQL conversion API
type QueryHandler struct {
	converter    *converter.Converter
	validator    *validation.QueryValidator
	dgraphClient *dgraph.Client
}

// NewQueryHandler creates a new query handler
func NewQueryHandler() *QueryHandler {
	schema := config.GetSchemaConfig()

	// Initialize Dgraph client
	dgraphClient, err := dgraph.NewClient(dgraph.DefaultConfig())
	if err != nil {
		// Log error but don't fail - allow converter to work without Dgraph
		fmt.Printf("⚠️ Warning: Could not connect to Dgraph: %v\n", err)
		fmt.Println("💡 To use /execute endpoint, start Dgraph with: docker-compose up -d")
	}

	return &QueryHandler{
		converter:    converter.NewConverter(),
		validator:    validation.NewQueryValidator(schema),
		dgraphClient: dgraphClient,
	}
}

// ConvertQuery handles POST /convert endpoint
func (h *QueryHandler) ConvertQuery(c *gin.Context) {
	var jsonQuery models.JSONQuery

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Enhanced validation using the new validator
	validationResult := h.validator.Validate(&jsonQuery)
	if !validationResult.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Query validation failed",
			"validation": validationResult,
		})
		return
	}

	// fmt.Println("query", jsonQuery)

	// Convert to DQL
	dqlQuery, err := h.converter.ConvertToDQL(&jsonQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to convert query",
			"details": err.Error(),
		})
		return
	}

	// Generate DQL string
	dqlString := h.converter.GenerateDQLString(dqlQuery)

	// Return response with validation warnings if any
	response := gin.H{
		"dql": dqlString,
	}

	if len(validationResult.Warnings) > 0 {
		response["warnings"] = validationResult.Warnings
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

// validateQuery performs basic validation on the JSON query
func (h *QueryHandler) validateQuery(query *models.JSONQuery) error {
	if query.CombineWith != "AND" && query.CombineWith != "OR" {
		return gin.Error{Err: fmt.Errorf("combine_with must be 'AND' or 'OR'"), Type: gin.ErrorTypePublic}
	}

	if len(query.Groups) == 0 {
		return gin.Error{Err: fmt.Errorf("at least one group is required"), Type: gin.ErrorTypePublic}
	}

	return h.validateGroups(query.Groups)
}

// validateGroups recursively validates groups
func (h *QueryHandler) validateGroups(groups []models.Group) error {
	for _, group := range groups {
		if group.CombineWith != "AND" && group.CombineWith != "OR" {
			return gin.Error{Err: fmt.Errorf("group combine_with must be 'AND' or 'OR'"), Type: gin.ErrorTypePublic}
		}

		if len(group.Filters) == 0 && len(group.Groups) == 0 {
			return gin.Error{Err: fmt.Errorf("group must have either filters or nested groups"), Type: gin.ErrorTypePublic}
		}

		// Validate filters
		for _, filter := range group.Filters {
			if filter.Field == "" {
				return gin.Error{Err: fmt.Errorf("filter field cannot be empty"), Type: gin.ErrorTypePublic}
			}
			if filter.Op == "" {
				return gin.Error{Err: fmt.Errorf("filter operator cannot be empty"), Type: gin.ErrorTypePublic}
			}
			if filter.Value == nil {
				return gin.Error{Err: fmt.Errorf("filter value cannot be null"), Type: gin.ErrorTypePublic}
			}
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			if err := h.validateGroups(group.Groups); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetSchema handles GET /schema endpoint - returns available fields and operators
func (h *QueryHandler) GetSchema(c *gin.Context) {
	schema := map[string]interface{}{
		"available_operators": []string{
			"=", ">=", "<=", ">", "<", "IN", "NOT_IN", "!=",
			"LIKE", "ILIKE", "REGEX", "BETWEEN", "IS_NULL", "IS_NOT_NULL",
			"STARTS_WITH", "ENDS_WITH", "CONTAINS",
		},
		"combine_operators": []string{"AND", "OR"},
		"available_fields": map[string][]string{
			"customer_fields": {
				"age", "country", "device", "app_version", "last_login_days",
				"email", "name",
			},
			"subscription_fields": {
				"subscription_status", "subscribed_package", "package", "status",
			},
			"content_fields": {
				"watched_content", "favorite_genres", "content_type", "genre", "title",
			},
			"device_fields": {
				"device_type", "os_version",
			},
		},
		"entity_types": []string{
			"chorki_customers", "chorki_subscriptions", "chorki_watch_histories",
			"chorki_contents", "chorki_devices",
		},
		"complexity_limits": h.converter.GetComplexityLimits(),
		"example_queries":   h.getExampleQueries(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"schema":  schema,
	})
}

// getExampleQueries returns example JSON queries
func (h *QueryHandler) getExampleQueries() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Simple customer filter",
			"description": "Find customers older than 18 from specific countries",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 18},
							{"field": "country", "op": "IN", "value": []string{"USA", "UK", "Canada"}},
						},
					},
				},
			},
		},
		{
			"name":        "Complex subscription filter",
			"description": "Find premium subscribers with recent activity",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "subscription_status", "op": "=", "value": "active"},
							{"field": "subscribed_package", "op": "=", "value": "Premium"},
						},
					},
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "last_login_days", "op": "<=", "value": 7},
							{"field": "app_version", "op": ">=", "value": "5.0.0"},
						},
					},
				},
			},
		},
	}
}

// HealthCheck handles GET /health endpoint
func (h *QueryHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "json-to-dql-converter",
		"version": "1.0.0",
	})
}

// AnalyzeComplexity handles POST /analyze endpoint
func (h *QueryHandler) AnalyzeComplexity(c *gin.Context) {
	var jsonQuery models.JSONQuery

	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	complexityScore := h.converter.AnalyzeComplexity(&jsonQuery)

	c.JSON(http.StatusOK, gin.H{
		"complexity": complexityScore,
	})
}

// ValidateQuery handles POST /validate endpoint
func (h *QueryHandler) ValidateQuery(c *gin.Context) {
	var jsonQuery models.JSONQuery

	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	validationResult := h.validator.Validate(&jsonQuery)

	c.JSON(http.StatusOK, gin.H{
		"validation": validationResult,
	})
}

// ExecuteQuery handles POST /execute endpoint - converts JSON to DQL and executes against Dgraph
func (h *QueryHandler) ExecuteQuery(c *gin.Context) {
	var jsonQuery models.JSONQuery

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Check if Dgraph client is available
	if h.dgraphClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Dgraph connection not available",
			"message": "Please start Dgraph using: docker-compose up -d",
			"tip":     "Use /convert endpoint to generate DQL without executing",
		})
		return
	}

	// Check Dgraph connection
	if !h.dgraphClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Dgraph connection lost",
			"message": "Please check Dgraph server status",
		})
		return
	}

	// Enhanced validation
	validationResult := h.validator.Validate(&jsonQuery)
	if !validationResult.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Query validation failed",
			"validation": validationResult,
		})
		return
	}

	// Convert JSON to DQL
	dqlQuery, err := h.converter.ConvertToDQL(&jsonQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DQL conversion failed",
			"details": err.Error(),
		})
		return
	}

	// Generate DQL string
	dqlString := h.converter.GenerateDQLString(dqlQuery)

	// Execute DQL query against Dgraph
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.dgraphClient.ExecuteDQL(ctx, dqlString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Query execution failed",
			"details":   err.Error(),
			"dql_query": dqlString,
		})
		return
	}

	// Get execution statistics
	stats := h.dgraphClient.GetExecutionStats(response)

	// Return successful response with data and metadata
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Data,
		"query_info": gin.H{
			"dql":        dqlString,
			"query_time": response.QueryTime,
			"stats":      stats,
		},
		"validation": validationResult,
	})
}

// UserSegmentation handles POST /userSegmentation endpoint - executes user segmentation queries
// Returns only users that match ALL specified conditions with only user node data
func (h *QueryHandler) UserSegmentation(c *gin.Context) {
	var jsonQuery models.JSONQuery

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Check if Dgraph client is available
	if h.dgraphClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Dgraph connection not available",
			"message": "Please start Dgraph using: docker-compose up -d",
			"tip":     "Use /convert endpoint to generate DQL without executing",
		})
		return
	}

	// Check Dgraph connection
	if !h.dgraphClient.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Dgraph connection lost",
			"message": "Please check Dgraph server status",
		})
		return
	}

	// Enhanced validation
	validationResult := h.validator.Validate(&jsonQuery)
	if !validationResult.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Query validation failed",
			"validation": validationResult,
		})
		return
	}

	// Force combine_with to "AND" for user segmentation to ensure ALL conditions match
	originalCombineWith := jsonQuery.CombineWith
	jsonQuery.CombineWith = "AND"

	// Convert JSON to DQL
	dqlQuery, err := h.converter.ConvertToDQL(&jsonQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DQL conversion failed",
			"details": err.Error(),
		})
		return
	}

	// Generate user-only DQL string (only user fields, no related entities)
	dqlString := h.generateUserOnlyDQL(dqlQuery)

	// Execute DQL query against Dgraph
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.dgraphClient.ExecuteDQL(ctx, dqlString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "User segmentation query execution failed",
			"details":   err.Error(),
			"dql_query": dqlString,
		})
		return
	}

	// Get execution statistics
	stats := h.dgraphClient.GetExecutionStats(response)

	// Extract and filter user data only
	userData := h.extractUserDataOnly(response.Data)
	userCount := h.countUsers(userData)

	// Return successful response with user-only segmentation data
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   userData,
		"segment_info": gin.H{
			"total_users":           userCount,
			"query_type":            "user_segmentation",
			"data_scope":            "user_details_only",
			"forced_combine_with":   "AND",
			"original_combine_with": originalCombineWith,
		},
		"query_info": gin.H{
			"dql":        dqlString,
			"query_time": response.QueryTime,
			"stats":      stats,
		},
		"validation": validationResult,
	})
}

// generateUserOnlyDQL creates a DQL query that returns only user node data
func (h *QueryHandler) generateUserOnlyDQL(dqlQuery *models.DQLQuery) string {
	if dqlQuery == nil || len(dqlQuery.Queries) == 0 {
		return ""
	}

	// Take the first query and modify it to select only user fields
	baseQuery := dqlQuery.Queries[0]

	// Define user-only fields
	userFields := []string{
		"uid",
		"chorki_customers.id",
		"chorki_customers.name",
		"chorki_customers.email",
		"chorki_customers.age",
		"chorki_customers.country",
		"chorki_customers.device",
		"chorki_customers.app_version",
		"chorki_customers.last_login_days",
		"chorki_customers.is_active",
		"chorki_customers.city",
	}

	// Build the user-only DQL query
	dqlString := fmt.Sprintf("{\n  %s(func: type(chorki_customers))", baseQuery.Name)

	// Add filter conditions if they exist, but remove @cascade for user segmentation
	if baseQuery.Filter != "" {
		// Remove @cascade from the filter as it's too restrictive for user segmentation
		cleanFilter := strings.Replace(baseQuery.Filter, " @cascade", "", -1)
		dqlString += " " + cleanFilter
	}

	dqlString += " {\n"

	// Add only user fields
	for _, field := range userFields {
		dqlString += "    " + field + "\n"
	}

	dqlString += "  }\n}"

	return dqlString
}

// extractUserDataOnly filters response to return only user node data
func (h *QueryHandler) extractUserDataOnly(data interface{}) interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Look for user data in the response
		for _, value := range dataMap {
			if userList, ok := value.([]interface{}); ok {
				// Filter each user object to contain only user fields
				filteredUsers := make([]interface{}, 0)
				for _, user := range userList {
					if userObj, ok := user.(map[string]interface{}); ok {
						filteredUser := h.filterUserFields(userObj)
						filteredUsers = append(filteredUsers, filteredUser)
					}
				}
				return filteredUsers
			}
		}
	}
	return []interface{}{} // Return empty array if no users found
}

// filterUserFields keeps only user-specific fields in the response
func (h *QueryHandler) filterUserFields(userObj map[string]interface{}) map[string]interface{} {
	allowedFields := map[string]bool{
		"uid":                              true,
		"chorki_customers.id":              true,
		"chorki_customers.name":            true,
		"chorki_customers.email":           true,
		"chorki_customers.age":             true,
		"chorki_customers.country":         true,
		"chorki_customers.device":          true,
		"chorki_customers.app_version":     true,
		"chorki_customers.last_login_days": true,
		"chorki_customers.is_active":       true,
		"chorki_customers.city":            true,
	}

	filtered := make(map[string]interface{})
	for key, value := range userObj {
		if allowedFields[key] {
			filtered[key] = value
		}
	}
	return filtered
}

// countUsers counts the number of users in the response
func (h *QueryHandler) countUsers(userData interface{}) int {
	if userList, ok := userData.([]interface{}); ok {
		return len(userList)
	}
	return 0
}
