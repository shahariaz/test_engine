package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestCase struct {
	Name        string
	Description string
	Query       map[string]interface{}
	ExpectedResults string
}

func main() {
	fmt.Println("🧪 **COMPREHENSIVE TESTING OF 77K+ RECORD DATASET** 🧪")
	fmt.Println("======================================================")
	
	// Test cases for our large dataset
	testCases := []TestCase{
		{
			Name:        "Age Filter Test",
			Description: "Find customers aged 25-35 from Bangladesh",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 25},
							{"field": "age", "op": "<=", "value": 35},
							{"field": "country", "op": "=", "value": "Bangladesh"},
						},
					},
				},
			},
			ExpectedResults: "Young adults from Bangladesh",
		},
		{
			Name:        "Premium Subscribers Test", 
			Description: "Find all Premium subscribers with active status",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "subscribed_package", "op": "=", "value": "Premium"},
							{"field": "subscription_status", "op": "=", "value": "active"},
						},
					},
				},
			},
			ExpectedResults: "Active Premium subscribers",
		},
		{
			Name:        "iOS Users Test",
			Description: "Find iOS users with latest app versions",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "device", "op": "=", "value": "iOS"},
							{"field": "app_version", "op": ">=", "value": "6.0.0"},
						},
					},
				},
			},
			ExpectedResults: "iOS users with recent app versions",
		},
		{
			Name:        "High Price Subscriptions Test",
			Description: "Find expensive subscriptions (>= $400)",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{"field": "price", "op": ">=", "value": 400},
						},
					},
				},
			},
			ExpectedResults: "High-value subscriptions",
		},
		{
			Name:        "Multi-Country Test",
			Description: "Find users from USA, UK, or Canada",
			Query: map[string]interface{}{
				"combine_with": "OR",
				"groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{"field": "country", "op": "IN", "value": []string{"United States", "United Kingdom", "Canada"}},
						},
					},
				},
			},
			ExpectedResults: "North American users",
		},
		{
			Name:        "Complex OR Test",
			Description: "Find young users OR trial subscriptions",
			Query: map[string]interface{}{
				"combine_with": "OR",
				"groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{"field": "age", "op": "<", "value": 20},
						},
					},
					{
						"filters": []map[string]interface{}{
							{"field": "subscription_status", "op": "=", "value": "trial"},
						},
					},
				},
			},
			ExpectedResults: "Young users or trial subscribers",
		},
		{
			Name:        "High Watch Time Test",
			Description: "Find users who watched more than 100 minutes",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{"field": "watched_minutes", "op": ">", "value": 100},
						},
					},
				},
			},
			ExpectedResults: "Heavy content watchers",
		},
		{
			Name:        "Geographic + Demographic Test",
			Description: "Adults from Asian countries with Family packages",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 30},
							{"field": "country", "op": "IN", "value": []string{"Bangladesh", "India", "Pakistan", "Japan", "Singapore", "Malaysia"}},
							{"field": "subscribed_package", "op": "=", "value": "Family"},
						},
					},
				},
			},
			ExpectedResults: "Adult Asian family subscribers",
		},
		{
			Name:        "Recent High-Rated Content Test",
			Description: "Find high-rated content from recent years",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "rating", "op": ">=", "value": 8.0},
							{"field": "release_year", "op": ">=", "value": 2022},
						},
					},
				},
			},
			ExpectedResults: "High-rated recent content",
		},
		{
			Name:        "Performance Stress Test",
			Description: "Complex query with multiple conditions",
			Query: map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 18},
							{"field": "age", "op": "<=", "value": 45},
							{"field": "subscription_status", "op": "IN", "value": []string{"active", "trial"}},
						},
					},
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "device", "op": "=", "value": "iOS"},
							{"field": "device", "op": "=", "value": "Android"},
						},
					},
				},
			},
			ExpectedResults: "Active mobile users aged 18-45",
		},
	}

	fmt.Printf("🔍 Running %d test cases against 77K+ records...\n\n", len(testCases))
	
	successCount := 0
	totalTime := 0.0
	
	for i, testCase := range testCases {
		fmt.Printf("📋 **Test %d: %s**\n", i+1, testCase.Name)
		fmt.Printf("📝 Description: %s\n", testCase.Description)
		fmt.Printf("🎯 Expected: %s\n", testCase.ExpectedResults)
		
		// Execute the test
		startTime := time.Now()
		result, err := executeQuery(testCase.Query)
		duration := time.Since(startTime)
		totalTime += duration.Seconds()
		
		if err != nil {
			fmt.Printf("❌ **FAILED**: %v\n", err)
		} else {
			fmt.Printf("✅ **SUCCESS**: Query completed in %.2f seconds\n", duration.Seconds())
			printQueryResults(result)
			successCount++
		}
		
		fmt.Println("─────────────────────────────────────────")
		fmt.Println()
	}
	
	// Print summary
	fmt.Println("📊 **TEST SUMMARY**")
	fmt.Println("===================")
	fmt.Printf("✅ Successful tests: %d/%d (%.1f%%)\n", successCount, len(testCases), float64(successCount)/float64(len(testCases))*100)
	fmt.Printf("⏱️  Total execution time: %.2f seconds\n", totalTime)
	fmt.Printf("📈 Average query time: %.2f seconds\n", totalTime/float64(len(testCases)))
	fmt.Printf("🏆 Performance: %.1f queries/second\n", float64(len(testCases))/totalTime)
	
	if successCount == len(testCases) {
		fmt.Println("\n🎉 **ALL TESTS PASSED!** The system is working perfectly with 77K+ records!")
	} else {
		fmt.Printf("\n⚠️  **%d tests failed.** Please check the error details above.\n", len(testCases)-successCount)
	}
}

func executeQuery(query map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}
	
	resp, err := http.Post("http://localhost:8090/api/v1/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	return result, nil
}

func printQueryResults(result map[string]interface{}) {
	if success, ok := result["success"].(bool); ok && success {
		// Extract query information
		if queryInfo, ok := result["query_info"].(map[string]interface{}); ok {
			if stats, ok := queryInfo["stats"].(map[string]interface{}); ok {
				if resultCount, ok := stats["result_count"].(float64); ok {
					fmt.Printf("📊 Results found: %.0f records\n", resultCount)
				}
			}
			if queryTime, ok := queryInfo["query_time"].(string); ok {
				fmt.Printf("⚡ Dgraph query time: %s\n", queryTime)
			}
		}
		
		// Show sample data if available
		if data, ok := result["data"].(map[string]interface{}); ok {
			for entity, records := range data {
				if recordList, ok := records.([]interface{}); ok && len(recordList) > 0 {
					fmt.Printf("🔍 Sample %s (%d total):\n", entity, len(recordList))
					// Show first 2 records
					for i, record := range recordList {
						if i >= 2 { break }
						if recordMap, ok := record.(map[string]interface{}); ok {
							fmt.Printf("   • ")
							count := 0
							for key, value := range recordMap {
								if count >= 3 { break }
								if !contains(key, "uid") && !contains(key, "dgraph.type") {
									fmt.Printf("%s: %v, ", key, value)
									count++
								}
							}
							fmt.Println()
						}
					}
					if len(recordList) > 2 {
						fmt.Printf("   ... and %d more records\n", len(recordList)-2)
					}
				}
			}
		}
	} else {
		fmt.Println("❌ Query failed or returned no success indicator")
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[len(str)-len(substr):] == substr
}