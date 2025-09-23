package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func runQuickValidation() {
	fmt.Println("🚀 **SYSTEM VALIDATION** 🚀")
	fmt.Println("============================")
	
	// Test 1: Simple customer query
	fmt.Println("\n📋 **Test 1: Basic Customer Query**")
	start := time.Now()
	query1 := map[string]interface{}{
		"customers": map[string]interface{}{
			"country": map[string]interface{}{
				"eq": "USA",
			},
		},
	}
	
	response1, err1 := executeAPIQuery(query1)
	duration1 := time.Since(start)
	
	if err1 != "" {
		fmt.Printf("❌ **FAILED**: %s\n", err1)
	} else {
		fmt.Printf("✅ **PASSED** (Duration: %v)\n", duration1)
		responseJSON, _ := json.MarshalIndent(response1, "", "  ")
		fmt.Printf("📤 Response: %s\n", string(responseJSON))
	}
	
	// Test 2: Content query
	fmt.Println("\n📋 **Test 2: Content Query**")
	start = time.Now()
	query2 := map[string]interface{}{
		"contents": map[string]interface{}{
			"type": map[string]interface{}{
				"eq": "movie",
			},
		},
	}
	
	response2, err2 := executeAPIQuery(query2)
	duration2 := time.Since(start)
	
	if err2 != "" {
		fmt.Printf("❌ **FAILED**: %s\n", err2)
	} else {
		fmt.Printf("✅ **PASSED** (Duration: %v)\n", duration2)
		responseJSON, _ := json.MarshalIndent(response2, "", "  ")
		fmt.Printf("📤 Response: %s\n", string(responseJSON))
	}
	
	// Test 3: Age filter
	fmt.Println("\n📋 **Test 3: Age Filter Query**")
	start = time.Now()
	query3 := map[string]interface{}{
		"customers": map[string]interface{}{
			"age": map[string]interface{}{
				"gte": 20,
				"lte": 40,
			},
		},
	}
	
	response3, err3 := executeAPIQuery(query3)
	duration3 := time.Since(start)
	
	if err3 != "" {
		fmt.Printf("❌ **FAILED**: %s\n", err3)
	} else {
		fmt.Printf("✅ **PASSED** (Duration: %v)\n", duration3)
		responseJSON, _ := json.MarshalIndent(response3, "", "  ")
		fmt.Printf("📤 Response: %s\n", string(responseJSON))
	}
	
	fmt.Println("\n🏁 **VALIDATION COMPLETE**")
}

func executeAPIQuery(query map[string]interface{}) (interface{}, string) {
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
	runQuickValidation()
}