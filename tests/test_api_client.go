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
	fmt.Println("🧪 Testing JSON to DQL Converter API")
	fmt.Println("=====================================")

	// Wait a moment for server to be ready
	time.Sleep(2 * time.Second)

	// Test 1: Health Check
	fmt.Println("\n1. Testing Health Check...")
	testHealthCheck()

	// Test 2: Schema Info
	fmt.Println("\n2. Testing Schema Information...")
	testSchemaInfo()

	// Test 3: Simple Query
	fmt.Println("\n3. Testing Simple Query Conversion...")
	testSimpleQuery()

	// Test 4: Complex Query
	fmt.Println("\n4. Testing Complex Query Conversion...")
	testComplexQuery()

	// Test 5: Watched Content Query
	fmt.Println("\n5. Testing Watched Content Query...")
	testWatchedContentQuery()

	fmt.Println("\n✅ All tests completed!")
}

func testHealthCheck() {
	resp, err := http.Get(baseURL + "/api/v1/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
}

func testSchemaInfo() {
	resp, err := http.Get(baseURL + "/api/v1/schema")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Status: %d\n", resp.StatusCode)
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if schema, ok := result["schema"].(map[string]interface{}); ok {
		if entityTypes, ok := schema["entity_types"].([]interface{}); ok {
			fmt.Printf("Available Entity Types: %v\n", entityTypes)
		}
	}
}

func testSimpleQuery() {
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

	result := makeAPICall(query)
	if result != nil {
		fmt.Println("✅ Simple query converted successfully")
		if dqlString, ok := result["dql_string"].(string); ok {
			fmt.Printf("Generated DQL:\n%s\n", dqlString)
		}
	}
}

func testComplexQuery() {
	query := map[string]interface{}{
		"combine_with": "OR",
		"groups": []map[string]interface{}{
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "device", "op": "IN", "value": []string{"iOS", "Android"}},
					{"field": "app_version", "op": ">=", "value": "5.0.0"},
				},
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "age", "op": "<", "value": 18},
							{"field": "country", "op": "=", "value": "Bangladesh"},
						},
					},
				},
			},
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "subscribed_package", "op": "=", "value": "Premium"},
					{"field": "subscription_status", "op": "=", "value": "trial"},
				},
			},
		},
	}

	result := makeAPICall(query)
	if result != nil {
		fmt.Println("✅ Complex query converted successfully")
		if dqlString, ok := result["dql_string"].(string); ok {
			fmt.Printf("Generated DQL:\n%s\n", dqlString)
		}
	}
}

func testWatchedContentQuery() {
	query := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{
						"field": "watched_content",
						"op":    "IN",
						"value": map[string]interface{}{
							"content_type": "Movie",
							"ids":          []int{111, 222, 333},
						},
					},
					{"field": "favorite_genres", "op": "IN", "value": []string{"Action", "Drama"}},
				},
			},
		},
	}

	result := makeAPICall(query)
	if result != nil {
		fmt.Println("✅ Watched content query converted successfully")
		if dqlString, ok := result["dql_string"].(string); ok {
			fmt.Printf("Generated DQL:\n%s\n", dqlString)
		}
	}
}

func makeAPICall(query map[string]interface{}) map[string]interface{} {
	jsonData, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("❌ Error marshaling JSON: %v\n", err)
		return nil
	}

	resp, err := http.Post(baseURL+"/api/v1/convert", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Error making API call: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ API Error (Status %d): %s\n", resp.StatusCode, string(body))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ Error unmarshaling response: %v\n", err)
		return nil
	}

	return result
}