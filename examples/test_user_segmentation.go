package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TestUserSegmentationEndpoint demonstrates the difference between /execute and /userSegmentation
func main() {
	baseURL := "http://localhost:8090/api/v1"

	// Sample query that would return related entities in /execute but only user data in /userSegmentation
	testQuery := map[string]interface{}{
		"combine_with": "OR", // This will be forced to AND in userSegmentation
		"groups": []map[string]interface{}{
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "age", "op": ">=", "value": 25},
					{"field": "country", "op": "IN", "value": []string{"USA", "UK", "Canada"}},
					{"field": "subscription_status", "op": "=", "value": "active"},
				},
			},
		},
	}

	fmt.Println("🧪 Testing User Segmentation vs Execute Endpoint")
	fmt.Println(strings.Repeat("=", 60))

	// Test /execute endpoint
	fmt.Println("\n1️⃣  Testing /execute endpoint...")
	executeResponse := testEndpoint(baseURL+"/execute", testQuery)
	fmt.Printf("Execute Response Structure: %s\n", getResponseStructure(executeResponse))

	// Test /userSegmentation endpoint
	fmt.Println("\n2️⃣  Testing /userSegmentation endpoint...")
	segmentResponse := testEndpoint(baseURL+"/userSegmentation", testQuery)
	fmt.Printf("UserSegmentation Response Structure: %s\n", getResponseStructure(segmentResponse))

	// Compare differences
	fmt.Println("\n📊 Key Differences:")
	fmt.Println("   /execute: Returns full entity data with relationships")
	fmt.Println("   /userSegmentation: Returns ONLY user node data")
	fmt.Println("   /userSegmentation: Forces 'combine_with' to 'AND'")
	fmt.Println("   /userSegmentation: Simplified response with users array at top level")
}

func testEndpoint(url string, query map[string]interface{}) map[string]interface{} {
	jsonData, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return nil
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error calling %s: %v\n", url, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return nil
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	return result
}

func getResponseStructure(response map[string]interface{}) string {
	if response == nil {
		return "No response"
	}

	structure := make([]string, 0)
	for key := range response {
		structure = append(structure, key)
	}

	structureJSON, _ := json.MarshalIndent(structure, "", "  ")
	return string(structureJSON)
}
