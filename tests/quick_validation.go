package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🚀 **QUICK SYSTEM TEST** 🚀")
	fmt.Println("===========================")

	// Test simple queries that have been working
	tests := []struct {
		name        string
		description string
		query       map[string]interface{}
	}{
		{
			name:        "Simple Country Filter",
			description: "Find customers from USA",
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "country",
								"op":    "=",
								"value": "USA",
							},
						},
					},
				},
			},
		},
		{
			name:        "Age Range Query",
			description: "Find customers aged 25-35",
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "age",
								"op":    ">=",
								"value": 25,
							},
							{
								"field": "age",
								"op":    "<=",
								"value": 35,
							},
						},
					},
				},
			},
		},
		{
			name:        "Movie Content Type",
			description: "Find movie content",
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "content_type",
								"op":    "=",
								"value": "movie",
							},
						},
					},
				},
			},
		},
	}

	successCount := 0
	totalTime := time.Duration(0)

	for i, test := range tests {
		fmt.Printf("\n📋 **Test %d: %s**\n", i+1, test.name)
		fmt.Printf("📝 Description: %s\n", test.description)

		start := time.Now()
		
		// Make API request
		success, responseBody := makeAPIRequest(test.query)
		
		duration := time.Since(start)
		totalTime += duration
		
		fmt.Printf("⏱️ Duration: %v\n", duration)
		
		if success {
			fmt.Printf("✅ **PASSED**\n")
			successCount++
			
			// Parse and show basic stats
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(responseBody), &response); err == nil {
				if data, ok := response["data"].(map[string]interface{}); ok {
					for key, value := range data {
						if arr, ok := value.([]interface{}); ok {
							fmt.Printf("📊 Found %d %s\n", len(arr), key)
						}
					}
				}
			}
		} else {
			fmt.Printf("❌ **FAILED**\n")
			fmt.Printf("📤 Error: %s\n", responseBody)
		}
		
		fmt.Println("─────────────────────────────────────────")
	}

	// Final summary
	fmt.Printf("\n📊 **FINAL SUMMARY**\n")
	fmt.Printf("====================\n")
	fmt.Printf("✅ Successful tests: %d/%d (%.1f%%)\n", successCount, len(tests), float64(successCount)/float64(len(tests))*100)
	fmt.Printf("⏱️ Total execution time: %v\n", totalTime)
	fmt.Printf("📈 Average query time: %v\n", totalTime/time.Duration(len(tests)))
	
	if successCount == len(tests) {
		fmt.Printf("🎉 **ALL TESTS PASSED!** System is working perfectly.\n")
	} else if successCount > 0 {
		fmt.Printf("⚠️ **SOME TESTS PASSED.** System is partially working.\n")
	} else {
		fmt.Printf("❌ **ALL TESTS FAILED.** System needs investigation.\n")
	}
}

func makeAPIRequest(query map[string]interface{}) (bool, string) {
	// Convert query to JSON
	jsonData, err := json.Marshal(query)
	if err != nil {
		return false, fmt.Sprintf("Failed to marshal JSON: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "http://localhost:8090/api/v1/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("Failed to read response: %v", err)
	}

	if resp.StatusCode == 200 {
		return true, string(body)
	} else {
		return false, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
	}
}