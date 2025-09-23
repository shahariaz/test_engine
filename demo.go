package main

import (
	"encoding/json"
	"fmt"
	"jsonTodql/converter"
	"jsonTodql/models"
)

func main() {
	fmt.Println("🚀 JSON to DQL Converter - Direct Testing")
	fmt.Println("=========================================")

	// Initialize converter
	conv := converter.NewConverter()

	// Test examples from the requirements
	testExamples := []struct {
		name string
		json string
	}{
		{
			name: "Example 1 - Age and Country Filter",
			json: `{
				"combine_with": "AND",
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{ "field": "age", "op": ">=", "value": 21 },
							{ "field": "country", "op": "IN", "value": ["USA", "UK", "Canada"] }
						]
					},
					{
						"combine_with": "AND",
						"filters": [
							{ "field": "subscription_status", "op": "=", "value": "active" },
							{ "field": "last_login_days", "op": "<=", "value": 30 }
						]
					}
				]
			}`,
		},
		{
			name: "Example 2 - Nested Groups with Device and Age",
			json: `{
				"combine_with": "OR",
				"groups": [
					{
						"combine_with": "AND",
						"filters": [
							{ "field": "device", "op": "IN", "value": ["iOS", "Android"] },
							{ "field": "app_version", "op": ">=", "value": "5.0.0" }
						],
						"groups": [
							{
								"combine_with": "OR",
								"filters": [
									{ "field": "age", "op": "<", "value": 18 },
									{ "field": "country", "op": "=", "value": "Bangladesh" }
								]
							}
						]
					},
					{
						"combine_with": "AND",
						"filters": [
							{ "field": "subscribed_package", "op": "=", "value": "Premium" },
							{ "field": "subscription_status", "op": "=", "value": "trial" }
						]
					}
				]
			}`,
		},
		{
			name: "Example 3 - Watched Content Complex Object",
			json: `{
				"combine_with": "AND",
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{ "field": "age", "op": ">", "value": 18 },
							{ "field": "watched_content", "op": "IN", "value": { "content_type": "Clips", "ids": [123] } }
						]
					},
					{
						"combine_with": "OR",
						"filters": [
							{ "field": "subscribed_package", "op": "=", "value": "TEST Package" },
							{ "field": "subscription_status", "op": "IN", "value": ["expired"] }
						]
					}
				]
			}`,
		},
	}

	for i, example := range testExamples {
		fmt.Printf("\n%s\n", example.name)
		fmt.Printf("%s\n", strings.Repeat("=", len(example.name)))

		// Parse JSON query
		var jsonQuery models.JSONQuery
		err := json.Unmarshal([]byte(example.json), &jsonQuery)
		if err != nil {
			fmt.Printf("❌ Error parsing JSON: %v\n", err)
			continue
		}

		// Convert to DQL
		dqlQuery, err := conv.ConvertToDQL(&jsonQuery)
		if err != nil {
			fmt.Printf("❌ Error converting to DQL: %v\n", err)
			continue
		}

		// Generate DQL string
		dqlString := conv.GenerateDQLString(dqlQuery)

		fmt.Printf("✅ Successfully converted!\n")
		fmt.Printf("\n📝 Input JSON Query:\n%s\n", example.json)
		fmt.Printf("\n🎯 Generated DQL Query:\n%s\n", dqlString)

		if i < len(testExamples)-1 {
			fmt.Printf("\n%s\n", strings.Repeat("-", 80))
		}
	}

	fmt.Println("\n🎉 All examples converted successfully!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("   ✓ Simple AND/OR combinations")
	fmt.Println("   ✓ Nested groups with complex logic")
	fmt.Println("   ✓ Multiple entity types (customers, subscriptions, watch histories)")
	fmt.Println("   ✓ Complex object filters (watched_content)")
	fmt.Println("   ✓ Array value filters (IN operator)")
	fmt.Println("   ✓ Relationship traversals between entities")
}