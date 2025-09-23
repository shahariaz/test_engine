package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	// Test basic customer query using JSON format
	testQueries := []map[string]interface{}{
		{
			"name": "Active Customers",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "is_active", "op": "=", "value": true},
						},
					},
				},
			},
		},
		{
			"name": "Customers by Country",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "country", "op": "=", "value": "USA"},
							{"field": "country", "op": "=", "value": "Canada"},
						},
					},
				},
			},
		},
		{
			"name": "Age Range Query",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 25},
							{"field": "age", "op": "<=", "value": 35},
						},
					},
				},
			},
		},
	}

	for _, test := range testQueries {
		fmt.Printf("\n=== Testing: %s ===\n", test["name"])
		
		// Convert to JSON
		jsonData, err := json.Marshal(test["query"])
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			continue
		}

		// Make HTTP request
		resp, err := http.Post("http://localhost:8090/api/v1/execute", 
			"application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// Read response
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			continue
		}

		// Pretty print response
		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("Response:\n%s\n", prettyJSON)
	}
}