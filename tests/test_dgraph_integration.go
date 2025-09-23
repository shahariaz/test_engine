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
	fmt.Println("🧪 Testing Dgraph Integration and /execute endpoint")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	// Test 1: Check if API server is running
	fmt.Println("\n1️⃣ Testing API Server Connection...")
	if !testHealthCheck() {
		fmt.Println("❌ API server not running. Please start with: go run main.go")
		return
	}

	// Test 2: Test /convert endpoint (should work without Dgraph)
	fmt.Println("\n2️⃣ Testing /convert endpoint...")
	testConvertEndpoint()

	// Test 3: Test /execute endpoint (requires Dgraph)
	fmt.Println("\n3️⃣ Testing /execute endpoint...")
	testExecuteEndpoint()

	fmt.Println("\n🎉 Testing complete!")
	fmt.Println("\n💡 Tips:")
	fmt.Println("   - If /execute fails, start Dgraph with: docker-compose up -d")
	fmt.Println("   - Use Ratel UI at http://localhost:8000 to explore data")
	fmt.Println("   - Check DGRAPH_SETUP_GUIDE.md for detailed instructions")
}

func testHealthCheck() bool {
	resp, err := http.Get(baseURL + "/api/v1/health")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ API server is running")
		return true
	}

	fmt.Printf("❌ Health check failed with status: %d\n", resp.StatusCode)
	return false
}

func testConvertEndpoint() {
	query := map[string]interface{}{
		"combine_with": "OR",
		"groups": []map[string]interface{}{
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "age", "op": "<", "value": 18},
					{"field": "country", "op": "=", "value": "Bangladesh"},
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

	result := makeAPICall("POST", "/api/v1/convert", query)
	if result["error"] != nil {
		fmt.Printf("❌ /convert failed: %v\n", result["error"])
		return
	}

	if dqlQueries, ok := result["dql_queries"].(map[string]interface{}); ok {
		if queries, ok := dqlQueries["queries"].([]interface{}); ok {
			fmt.Printf("✅ /convert successful - Generated %d DQL queries\n", len(queries))
			if len(queries) > 0 {
				if query, ok := queries[0].(map[string]interface{}); ok {
					fmt.Printf("   📝 Sample query: %s\n", truncateString(query["filter"].(string), 60))
				}
			}
		}
	} else {
		fmt.Println("✅ /convert returned response (format may have changed)")
	}
}

func testExecuteEndpoint() {
	query := map[string]interface{}{
		"combine_with": "AND",
		"groups": []map[string]interface{}{
			{
				"combine_with": "OR",
				"filters": []map[string]interface{}{
					{"field": "age", "op": "<", "value": 30},
					{"field": "country", "op": "=", "value": "Bangladesh"},
				},
			},
		},
	}

	result := makeAPICall("POST", "/api/v1/execute", query)
	
	if result["error"] != nil {
		errorMsg := result["error"].(string)
		if errorMsg == "Dgraph connection not available" {
			fmt.Println("⚠️  /execute endpoint requires Dgraph")
			fmt.Println("   💡 Start Dgraph with: docker-compose up -d")
			fmt.Println("   📚 See DGRAPH_SETUP_GUIDE.md for detailed instructions")
		} else {
			fmt.Printf("❌ /execute failed: %v\n", result["error"])
			if details, ok := result["details"]; ok {
				fmt.Printf("   Details: %v\n", details)
			}
		}
		return
	}

	if success, ok := result["success"].(bool); ok && success {
		fmt.Println("✅ /execute successful - Connected to Dgraph and executed query!")
		
		if data, ok := result["data"]; ok {
			fmt.Printf("   📊 Returned data: %v\n", summarizeData(data))
		}
		
		if queryInfo, ok := result["query_info"].(map[string]interface{}); ok {
			if queryTime, ok := queryInfo["query_time"].(string); ok {
				fmt.Printf("   ⏱️  Query time: %s\n", queryTime)
			}
			if stats, ok := queryInfo["stats"].(map[string]interface{}); ok {
				if resultCount, ok := stats["result_count"].(float64); ok {
					fmt.Printf("   📈 Result count: %.0f\n", resultCount)
				}
			}
		}
	} else {
		fmt.Println("❌ /execute returned unexpected format")
	}
}

func makeAPICall(method, endpoint string, data interface{}) map[string]interface{} {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("JSON marshal error: %v", err)}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Request creation error: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Request error: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Response read error: %v", err)}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("JSON unmarshal error: %v", err)}
	}

	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func summarizeData(data interface{}) string {
	if dataMap, ok := data.(map[string]interface{}); ok {
		count := 0
		var entityTypes []string
		for key, value := range dataMap {
			if array, ok := value.([]interface{}); ok {
				count += len(array)
				entityTypes = append(entityTypes, key)
			}
		}
		if len(entityTypes) > 0 {
			return fmt.Sprintf("%d records across %v", count, entityTypes)
		}
	}
	return "Unknown format"
}