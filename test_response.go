package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("🧪 Testing Updated API Response")
	fmt.Println("==============================")

	// Test query from your requirements
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
			{
				"combine_with": "AND",
				"filters": []map[string]interface{}{
					{"field": "subscription_status", "op": "=", "value": "active"},
					{"field": "last_login_days", "op": "<=", "value": 30},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(query)

	resp, err := http.Post("http://localhost:8080/api/v1/convert", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📊 Status Code: %d\n", resp.StatusCode)
	fmt.Printf("📝 Response:\n%s\n", string(body))

	// Parse and show just the DQL
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		if dql, ok := result["dql"].(string); ok {
			fmt.Printf("\n🎯 Clean DQL Query:\n%s\n", dql)
		}
	}
}
