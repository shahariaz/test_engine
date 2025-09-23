package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("🧪 Testing the specific query that had issues")
	fmt.Println("============================================")

	// Your specific query
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

	fmt.Println("📝 Input JSON Query:")
	jsonBytes, _ := json.MarshalIndent(query, "", "  ")
	fmt.Println(string(jsonBytes))

	// Convert the query
	fmt.Println("\n🔄 Converting to DQL...")
	result := makeAPICall("POST", "http://localhost:8080/api/v1/convert", query)

	if result != nil {
		if dql, ok := result["dql"].(string); ok {
			fmt.Printf("\n🎯 Generated DQL:\n%s\n", dql)

			// Analyze what entities are involved
			fmt.Println("\n📊 Analysis:")
			if containsString(dql, "customers(func: type(chorki_customers))") {
				fmt.Println("✅ Found chorki_customers query")
			}
			if containsString(dql, "subscriptions(func: type(chorki_subscriptions))") {
				fmt.Println("✅ Found chorki_subscriptions query")
			} else {
				fmt.Println("❌ Missing chorki_subscriptions query")
			}

			// Count the number of separate queries
			queryCount := countQueries(dql)
			fmt.Printf("📈 Total number of entity queries: %d\n", queryCount)

			if queryCount >= 2 {
				fmt.Println("✅ Multiple entity queries generated correctly!")
			} else {
				fmt.Println("❌ Expected multiple entity queries but only found one")
			}
		}
	}

	// Also test the validation
	fmt.Println("\n🔍 Testing validation...")
	validationResult := makeAPICall("POST", "http://localhost:8080/api/v1/validate", query)
	if validationResult != nil {
		if validation, ok := validationResult["validation"].(map[string]interface{}); ok {
			if isValid, ok := validation["is_valid"].(bool); ok {
				if isValid {
					fmt.Println("✅ Query validation passed")
				} else {
					fmt.Println("❌ Query validation failed")
					if errors, ok := validation["errors"].([]interface{}); ok {
						fmt.Printf("   Errors: %v\n", errors)
					}
				}
			}
		}
	}

	// Test complexity analysis
	fmt.Println("\n📊 Testing complexity analysis...")
	complexityResult := makeAPICall("POST", "http://localhost:8080/api/v1/analyze", query)
	if complexityResult != nil {
		if complexity, ok := complexityResult["complexity"].(map[string]interface{}); ok {
			if score, ok := complexity["total_score"].(float64); ok {
				fmt.Printf("📈 Complexity score: %.0f\n", score)
			}
			if acceptable, ok := complexity["is_acceptable"].(bool); ok {
				if acceptable {
					fmt.Println("✅ Complexity within acceptable limits")
				} else {
					fmt.Println("❌ Complexity exceeds limits")
				}
			}
		}
	}
}

func makeAPICall(method, url string, data interface{}) map[string]interface{} {
	jsonData, _ := json.Marshal(data)

	var req *http.Request
	var err error

	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		return nil
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API Error (Status %d): %s\n", resp.StatusCode, string(body))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ Error parsing response: %v\n", err)
		return nil
	}

	return result
}

func containsString(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text[:len(substr)] == substr ||
			containsString(text[1:], substr))
}

func countQueries(dql string) int {
	count := 0
	queries := []string{
		"customers(func:",
		"subscriptions(func:",
		"watch_histories(func:",
		"contents(func:",
		"devices(func:",
	}

	for _, queryType := range queries {
		if containsString(dql, queryType) {
			count++
		}
	}

	return count
}
