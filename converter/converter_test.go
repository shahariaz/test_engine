package converter

import (
	"encoding/json"
	"jsonTodql/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ConvertToDQL(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name           string
		input          string
		expectedQueries int
		expectedFields []string
	}{
		{
			name: "Simple customer age filter",
			input: `{
				"combine_with": "AND",
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{"field": "age", "op": ">=", "value": 21},
							{"field": "country", "op": "IN", "value": ["USA", "UK", "Canada"]}
						]
					}
				]
			}`,
			expectedQueries: 1,
			expectedFields:  []string{"chorki_customers"},
		},
		{
			name: "Complex multi-entity filter",
			input: `{
				"combine_with": "AND",
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{"field": "age", "op": ">=", "value": 21},
							{"field": "subscription_status", "op": "=", "value": "active"}
						]
					}
				]
			}`,
			expectedQueries: 2,
			expectedFields:  []string{"chorki_customers", "chorki_subscriptions"},
		},
		{
			name: "Watched content filter",
			input: `{
				"combine_with": "AND",
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{"field": "watched_content", "op": "IN", "value": {"content_type": "Movie", "ids": [111, 222, 333]}}
						]
					}
				]
			}`,
			expectedQueries: 1,
			expectedFields:  []string{"chorki_watch_histories"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonQuery models.JSONQuery
			err := json.Unmarshal([]byte(tt.input), &jsonQuery)
			require.NoError(t, err)

			dqlQuery, err := converter.ConvertToDQL(&jsonQuery)
			require.NoError(t, err)
			require.NotNil(t, dqlQuery)

			assert.Equal(t, tt.expectedQueries, len(dqlQuery.Queries))

			// Check that expected entity types are included
			entityTypes := make(map[string]bool)
			for _, query := range dqlQuery.Queries {
				entityTypes[query.Type] = true
			}

			for _, expectedField := range tt.expectedFields {
				assert.True(t, entityTypes[expectedField], "Expected entity type %s not found", expectedField)
			}
		})
	}
}

func TestConverter_GenerateDQLString(t *testing.T) {
	converter := NewConverter()

	// Test with a simple query
	input := `{
		"combine_with": "AND",
		"groups": [
			{
				"combine_with": "OR",
				"filters": [
					{"field": "age", "op": ">=", "value": 21},
					{"field": "country", "op": "=", "value": "USA"}
				]
			}
		]
	}`

	var jsonQuery models.JSONQuery
	err := json.Unmarshal([]byte(input), &jsonQuery)
	require.NoError(t, err)

	dqlQuery, err := converter.ConvertToDQL(&jsonQuery)
	require.NoError(t, err)

	dqlString := converter.GenerateDQLString(dqlQuery)

	// Basic checks for DQL format
	assert.Contains(t, dqlString, "{")
	assert.Contains(t, dqlString, "}")
	assert.Contains(t, dqlString, "func:")
	assert.Contains(t, dqlString, "@filter")
	assert.Contains(t, dqlString, "uid")

	t.Logf("Generated DQL:\n%s", dqlString)
}

func TestConverter_BuildFilterCondition(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name           string
		field          string
		op             string
		value          interface{}
		entityType     string
		expectedResult bool
	}{
		{
			name:           "Age greater than",
			field:          "age",
			op:             ">=",
			value:          21,
			entityType:     "chorki_customers",
			expectedResult: true,
		},
		{
			name:           "Country equals",
			field:          "country",
			op:             "=",
			value:          "USA",
			entityType:     "chorki_customers",
			expectedResult: true,
		},
		{
			name:           "Country IN array",
			field:          "country",
			op:             "IN",
			value:          []interface{}{"USA", "UK", "Canada"},
			entityType:     "chorki_customers",
			expectedResult: true,
		},
		{
			name:           "Invalid field",
			field:          "invalid_field",
			op:             "=",
			value:          "test",
			entityType:     "chorki_customers",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := models.Filter{
				Field: tt.field,
				Op:    tt.op,
				Value: tt.value,
			}

			result := converter.buildFilterCondition(filter, tt.entityType)

			if tt.expectedResult {
				assert.NotEmpty(t, result, "Expected non-empty result for valid filter")
				t.Logf("Generated condition: %s", result)
			} else {
				assert.Empty(t, result, "Expected empty result for invalid filter")
			}
		})
	}
}

func TestConverter_ComplexObjectFilter(t *testing.T) {
	converter := NewConverter()

	input := `{
		"combine_with": "AND",
		"groups": [
			{
				"combine_with": "OR",
				"filters": [
					{
						"field": "watched_content",
						"op": "IN",
						"value": {
							"content_type": "Movie",
							"ids": [111, 222, 333]
						}
					}
				]
			}
		]
	}`

	var jsonQuery models.JSONQuery
	err := json.Unmarshal([]byte(input), &jsonQuery)
	require.NoError(t, err)

	dqlQuery, err := converter.ConvertToDQL(&jsonQuery)
	require.NoError(t, err)
	require.NotNil(t, dqlQuery)

	dqlString := converter.GenerateDQLString(dqlQuery)

	// Check for complex object handling
	assert.Contains(t, dqlString, "uid_in")
	assert.Contains(t, dqlString, "111, 222, 333")

	t.Logf("Generated DQL for complex object:\n%s", dqlString)
}

func TestConverter_NestedGroups(t *testing.T) {
	converter := NewConverter()

	input := `{
		"combine_with": "OR",
		"groups": [
			{
				"combine_with": "AND",
				"filters": [
					{"field": "device", "op": "IN", "value": ["iOS", "Android"]},
					{"field": "app_version", "op": ">=", "value": "5.0.0"}
				],
				"groups": [
					{
						"combine_with": "OR",
						"filters": [
							{"field": "age", "op": "<", "value": 18},
							{"field": "country", "op": "=", "value": "Bangladesh"}
						]
					}
				]
			}
		]
	}`

	var jsonQuery models.JSONQuery
	err := json.Unmarshal([]byte(input), &jsonQuery)
	require.NoError(t, err)

	dqlQuery, err := converter.ConvertToDQL(&jsonQuery)
	require.NoError(t, err)
	require.NotNil(t, dqlQuery)

	dqlString := converter.GenerateDQLString(dqlQuery)

	// Check for nested conditions
	assert.Contains(t, dqlString, "AND")
	assert.Contains(t, dqlString, "OR")

	t.Logf("Generated DQL for nested groups:\n%s", dqlString)
}