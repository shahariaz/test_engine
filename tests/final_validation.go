package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func runCompleteValidation() {
	fmt.Println("🚀 **COMPLETE SYSTEM VALIDATION** 🚀")
	fmt.Println("====================================")
	
	tests := []struct {
		name        string
		description string
		query       map[string]interface{}
	}{
		{
			name:        "Customer by Country",
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
			name:        "Customer by Age Range",
			description: "Find customers aged 20-40",
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "age",
								"op":    ">=",
								"value": 20,
							},
							{
								"field": "age",
								"op":    "<=",
								"value": 40,
							},
						},
					},
				},
			},
		},
		{
			name:        "Movie Content",
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
		{
			name:        "Active Customers by Email",
			description: "Find customers with example.com emails", 
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "email",
								"op":    "CONTAINS",
								"value": "example.com",
							},
						},
					},
				},
			},
		},
		{
			name:        "Customer by Name",
			description: "Find customers named John",
			query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{
								"field": "name",
								"op":    "CONTAINS",
								"value": "John",
							},
						},
					},
				},
			},
		},
	}
	
	successCount := 0
	totalDuration := time.Duration(0)
	
	for i, test := range tests {
		fmt.Printf("\n📋 **Test %d: %s**\n", i+1, test.name)
		fmt.Printf("📝 Description: %s\n", test.description)
		
		start := time.Now()
		response, err := executeValidQuery(test.query)
		duration := time.Since(start)
		totalDuration += duration
		
		fmt.Printf("⏱️ Duration: %v\n", duration)
		
		if err != "" {
			fmt.Printf("❌ **FAILED**: %s\n", err)
		} else {
			fmt.Printf("✅ **PASSED**\n")
			responseJSON, _ := json.MarshalIndent(response, "", "  ")
			fmt.Printf("📤 Response: %s\n", string(responseJSON))
			successCount++
		}
		fmt.Println("─────────────────────────────────────────")
	}
	
	fmt.Printf("\n📊 **FINAL SUMMARY**\n")
	fmt.Printf("====================\n")
	fmt.Printf("✅ Successful tests: %d/%d (%.1f%%)\n", successCount, len(tests), float64(successCount)/float64(len(tests))*100)
	fmt.Printf("⏱️ Total execution time: %v\n", totalDuration)
	fmt.Printf("📈 Average query time: %v\n", totalDuration/time.Duration(len(tests)))
	fmt.Printf("🏆 Performance: %.1f queries/second\n", float64(len(tests))/totalDuration.Seconds())
	
	if successCount == len(tests) {
		fmt.Println("🎉 **ALL TESTS PASSED!** System is working correctly!")
		fmt.Println("💪 The JSON→DQL→Dgraph pipeline is fully functional!")
	} else if successCount > 0 {
		fmt.Println("⚠️ **SOME TESTS PASSED.** System is partially working.")
		fmt.Printf("🔧 %d tests need investigation.\n", len(tests)-successCount)
	} else {
		fmt.Println("💥 **ALL TESTS FAILED.** System needs immediate attention.")
	}
}

func executeValidQuery(query map[string]interface{}) (interface{}, string) {
	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Sprintf("Query marshaling error: %v", err)
	}
	
	// Make API request
	resp, err := http.Post("http://localhost:8090/api/v1/execute", "application/json", bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, fmt.Sprintf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Sprintf("Response reading error: %v", err)
	}
	
	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var response interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Sprintf("Response parsing error: %v", err)
	}
	
	return response, ""
}

func main() {
	runCompleteValidation()
}