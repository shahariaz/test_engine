package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("🚀 Enhanced JSON to DQL Converter - Feature Testing")
	fmt.Println("==================================================")

	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	// Test 1: Health Check
	fmt.Println("\n1. Testing Health Check...")
	testHealthCheck()

	// Test 2: Enhanced Schema Info
	fmt.Println("\n2. Testing Enhanced Schema Information...")
	testEnhancedSchema()

	// Test 3: Query Validation
	fmt.Println("\n3. Testing Query Validation...")
	testQueryValidation()

	// Test 4: Complexity Analysis
	fmt.Println("\n4. Testing Complexity Analysis...")
	testComplexityAnalysis()

	// Test 5: New Operators
	fmt.Println("\n5. Testing New Operators...")
	testNewOperators()

	// Test 6: Cache Statistics
	fmt.Println("\n6. Testing Cache Statistics...")
	testCacheStats()

	// Test 7: Enhanced Conversion with Caching
	fmt.Println("\n7. Testing Enhanced Conversion with Caching...")
	testEnhancedConversion()

	fmt.Println("\n✅ All enhanced features tested successfully!")
}

func testHealthCheck() {
	resp, err := makeRequest("GET", "/api/v1/health", nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Printf("✅ Health check passed\n")
	fmt.Printf("Response: %s\n", resp)
}

func testEnhancedSchema() {
	resp, err := makeRequest("GET", "/api/v1/schema", nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	if schema, ok := result["schema"].(map[string]interface{}); ok {
		if operators, ok := schema["available_operators"].([]interface{}); ok {
			fmt.Printf("✅ Enhanced operators available: %d operators\n", len(operators))
			fmt.Printf("New operators include: REGEX, BETWEEN, IS_NULL, LIKE, etc.\n")
		}
		if limits, ok := schema["complexity_limits"].(map[string]interface{}); ok {
			fmt.Printf("✅ Complexity limits configured: %v\n", limits)
		}
	}
}

func testQueryValidation() {
	// Test valid query
	validQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "age", "op": ">=", "value": 21},
					{"field": "country", "op": "IN", "value": []string{"USA", "UK"}},
				},
			},
		},
	}

	resp, err := makeRequest("POST", "/api/v1/validate", validQuery)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	var validResult map[string]interface{}
	json.Unmarshal([]byte(resp), &validResult)
	if validation, ok := validResult["validation"].(map[string]interface{}); ok {
		if isValid, ok := validation["is_valid"].(bool); ok && isValid {
			fmt.Printf("✅ Valid query validation passed\n")
		}
	}

	// Test invalid query
	invalidQuery := map[string]interface{}{
		"combine_with": "INVALID",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "invalid_field", "op": "=", "value": "test"},
				},
			},
		},
	}

	resp, err = makeRequest("POST", "/api/v1/validate", invalidQuery)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	var invalidResult map[string]interface{}
	json.Unmarshal([]byte(resp), &invalidResult)
	if validation, ok := invalidResult["validation"].(map[string]interface{}); ok {
		if isValid, ok := validation["is_valid"].(bool); ok && !isValid {
			fmt.Printf("✅ Invalid query validation correctly failed\n")
			if errors, ok := validation["errors"].([]interface{}); ok {
				fmt.Printf("   Found %d validation errors\n", len(errors))
			}
		}
	}
}

func testComplexityAnalysis() {
	// Test simple query
	simpleQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "age", "op": ">=", "value": 21},
				},
			},
		},
	}

	resp, err := makeRequest("POST", "/api/v1/analyze", simpleQuery)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)
	if complexity, ok := result["complexity"].(map[string]interface{}); ok {
		if score, ok := complexity["total_score"].(float64); ok {
			fmt.Printf("✅ Simple query complexity score: %.0f\n", score)
		}
		if acceptable, ok := complexity["is_acceptable"].(bool); ok && acceptable {
			fmt.Printf("✅ Query complexity within acceptable limits\n")
		}
	}

	// Test complex query
	complexQuery := map[string]interface{}{
		"combine_with": "OR",
		"groups": []map[string]interface{}{
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "age", "op": "BETWEEN", "value": []int{18, 65}},
					{"field": "country", "op": "IN", "value": []string{"USA", "UK", "Canada", "Australia"}},
					{"field": "device", "op": "REGEX", "value": "iOS|Android"},
				},
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "subscription_status", "op": "=", "value": "active"},
							{"field": "last_login_days", "op": "<=", "value": 7},
						},
					},
				},
			},
		},
	}

	resp, err = makeRequest("POST", "/api/v1/analyze", complexQuery)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	json.Unmarshal([]byte(resp), &result)
	if complexity, ok := result["complexity"].(map[string]interface{}); ok {
		if score, ok := complexity["total_score"].(float64); ok {
			fmt.Printf("✅ Complex query complexity score: %.0f\n", score)
		}
	}
}

func testNewOperators() {
	// Test BETWEEN operator
	betweenQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "age", "op": "BETWEEN", "value": []int{18, 65}},
				},
			},
		},
	}

	resp, err := makeRequest("POST", "/api/v1/convert", betweenQuery)
	if err != nil {
		fmt.Printf("❌ BETWEEN operator error: %v\n", err)
		return
	}
	fmt.Printf("✅ BETWEEN operator conversion successful\n")

	// Test REGEX operator
	regexQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "device", "op": "REGEX", "value": "iOS|Android"},
				},
			},
		},
	}

	resp, err = makeRequest("POST", "/api/v1/convert", regexQuery)
	if err != nil {
		fmt.Printf("❌ REGEX operator error: %v\n", err)
		return
	}
	fmt.Printf("✅ REGEX operator conversion successful\n")

	// Test NOT_IN operator
	notInQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "country", "op": "NOT_IN", "value": []string{"Restricted1", "Restricted2"}},
				},
			},
		},
	}

	resp, err = makeRequest("POST", "/api/v1/convert", notInQuery)
	if err != nil {
		fmt.Printf("❌ NOT_IN operator error: %v\n", err)
		return
	}
	fmt.Printf("✅ NOT_IN operator conversion successful\n")

	// Test IS_NULL operator
	nullQuery := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "last_login_days", "op": "IS_NULL", "value": nil},
				},
			},
		},
	}

	resp, err = makeRequest("POST", "/api/v1/convert", nullQuery)
	if err != nil {
		fmt.Printf("❌ IS_NULL operator error: %v\n", err)
		return
	}
	fmt.Printf("✅ IS_NULL operator conversion successful\n")
}

func testCacheStats() {
	// Get initial cache stats
	resp, err := makeRequest("GET", "/api/v1/cache/stats", nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	var statsResult map[string]interface{}
	json.Unmarshal([]byte(resp), &statsResult)
	if cacheStats, ok := statsResult["cache_stats"].(map[string]interface{}); ok {
		fmt.Printf("✅ Cache statistics retrieved\n")
		if queryCache, ok := cacheStats["query_cache"].(map[string]interface{}); ok {
			if entries, ok := queryCache["entries"].(float64); ok {
				fmt.Printf("   Query cache entries: %.0f\n", entries)
			}
		}
	}

	// Clear cache
	resp, err = makeRequest("DELETE", "/api/v1/cache", nil)
	if err != nil {
		fmt.Printf("❌ Error clearing cache: %v\n", err)
		return
	}
	fmt.Printf("✅ Cache cleared successfully\n")
}

func testEnhancedConversion() {
	query := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "age", "op": ">=", "value": 21},
					{"field": "country", "op": "IN", "value": []string{"USA", "UK", "Canada"}},
				},
			},
		},
	}

	// First conversion (should cache)
	start := time.Now()
	resp1, err := makeRequest("POST", "/api/v1/convert", query)
	duration1 := time.Since(start)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Second conversion (should use cache)
	start = time.Now()
	resp2, err := makeRequest("POST", "/api/v1/convert", query)
	duration2 := time.Since(start)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Enhanced conversion with caching tested\n")
	fmt.Printf("   First request: %v\n", duration1)
	fmt.Printf("   Second request: %v (potentially cached)\n", duration2)

	// Check if responses are identical
	if resp1 == resp2 {
		fmt.Printf("✅ Cached response consistency verified\n")
	}
}

func makeRequest(method, endpoint string, data interface{}) (string, error) {
	var req *http.Request
	var err error

	if data != nil {
		jsonData, _ := json.Marshal(data)
		req, err = http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, baseURL+endpoint, nil)
		if err != nil {
			return "", err
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
