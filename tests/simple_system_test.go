package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestResult struct {
	Name        string
	Description string
	Success     bool
	Duration    time.Duration
	Response    interface{}
	Error       string
}

func runSimpleTest() {
	fmt.Println("🚀 **SIMPLE SYSTEM VALIDATION TEST** 🚀")
	fmt.Println("=====================================")

	var results []TestResult
	
	// Test 1: Simple customer query
	test1 := TestResult{
		Name:        "Basic Customer Query",
		Description: "Get all customers",
	}
	
	start := time.Now()
	query := map[string]interface{}{
		"customers": map[string]interface{}{
			"country": map[string]interface{}{
				"eq": "USA",
			},
		},
	}
	
	test1.Response, test1.Error = executeQuery(query)
	test1.Duration = time.Since(start)
	test1.Success = test1.Error == ""
	results = append(results, test1)
	
	// Test 2: Age filter
	test2 := TestResult{
		Name:        "Age Filter Query", 
		Description: "Find customers aged 25-35",
	}
	
	start = time.Now()
	query2 := map[string]interface{}{
		"customers": map[string]interface{}{
			"age": map[string]interface{}{
				"gte": 25,
				"lte": 35,
			},
		},
	}
	
	test2.Response, test2.Error = executeQuery(query2)
	test2.Duration = time.Since(start)
	test2.Success = test2.Error == ""
	results = append(results, test2)
	
	// Test 3: Movie content
	test3 := TestResult{
		Name:        "Content Query",
		Description: "Find movie content",
	}
	
	start = time.Now()
	query3 := map[string]interface{}{
		"contents": map[string]interface{}{
			"type": map[string]interface{}{
				"eq": "movie",
			},
		},
	}
	
	test3.Response, test3.Error = executeQuery(query3)
	test3.Duration = time.Since(start)
	test3.Success = test3.Error == ""
	results = append(results, test3)
	
	// Test 4: Watch history
	test4 := TestResult{
		Name:        "Watch History Query",
		Description: "Find watch histories",
	}
	
	start = time.Now()
	query4 := map[string]interface{}{
		"watch_histories": map[string]interface{}{
			"completion_percentage": map[string]interface{}{
				"gte": 50.0,
			},
		},
	}
	
	test4.Response, test4.Error = executeQuery(query4)
	test4.Duration = time.Since(start)
	test4.Success = test4.Error == ""
	results = append(results, test4)
	
	// Print results
	fmt.Println("\n📊 **TEST RESULTS**")
	fmt.Println("===================")
	
	successCount := 0
	for i, result := range results {
		fmt.Printf("\n📋 **Test %d: %s**\n", i+1, result.Name)
		fmt.Printf("📝 Description: %s\n", result.Description)
		fmt.Printf("⏱️ Duration: %v\n", result.Duration)
		
		if result.Success {
			fmt.Printf("✅ **PASSED**\n")
			if result.Response != nil {
				responseJSON, _ := json.MarshalIndent(result.Response, "", "  ")
				fmt.Printf("📤 Response: %s\n", string(responseJSON))
			}
			successCount++
		} else {
			fmt.Printf("❌ **FAILED**: %s\n", result.Error)
		}
		fmt.Println("─────────────────────────────────────────")
	}
	
	fmt.Printf("\n📊 **SUMMARY**\n")
	fmt.Printf("✅ Successful tests: %d/%d (%.1f%%)\n", successCount, len(results), float64(successCount)/float64(len(results))*100)
	
	if successCount == len(results) {
		fmt.Println("🎉 **ALL TESTS PASSED!** System is working correctly!")
	} else if successCount > 0 {
		fmt.Println("⚠️ **SOME TESTS PASSED.** System is partially working.")
	} else {
		fmt.Println("💥 **ALL TESTS FAILED.** System needs investigation.")
	}
}

func executeQuery(query map[string]interface{}) (interface{}, string) {
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
	runSimpleTest()
}