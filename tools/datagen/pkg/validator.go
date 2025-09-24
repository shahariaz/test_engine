package pkg

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Validator validates generated data
type Validator struct {
	logger *log.Logger
}

// NewValidator creates a new validator
func NewValidator(logger *log.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidationResult holds validation results
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// ValidateBatch validates a batch of data
func (v *Validator) ValidateBatch(data []map[string]interface{}) error {
	var allErrors []string

	for i, item := range data {
		result := v.ValidateItem(item)
		if !result.Valid {
			for _, err := range result.Errors {
				allErrors = append(allErrors, fmt.Sprintf("Item %d: %s", i+1, err))
			}
		}

		// Log warnings
		for _, warning := range result.Warnings {
			v.logger.Printf("Warning for item %d: %s", i+1, warning)
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(allErrors, "; "))
	}

	v.logger.Printf("Successfully validated batch of %d items", len(data))
	return nil
}

// ValidateItem validates a single data item
func (v *Validator) ValidateItem(item map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check for required dgraph.type
	if dgraphType, exists := item["dgraph.type"]; !exists {
		result.Valid = false
		result.Errors = append(result.Errors, "missing dgraph.type field")
	} else {
		// Validate dgraph.type format
		if typeSlice, ok := dgraphType.([]interface{}); ok && len(typeSlice) > 0 {
			if typeStr, ok := typeSlice[0].(string); ok {
				result = v.validateByType(item, typeStr, result)
			} else {
				result.Valid = false
				result.Errors = append(result.Errors, "dgraph.type must be a string")
			}
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, "dgraph.type must be a non-empty array")
		}
	}

	return result
}

// validateByType validates item based on its type
func (v *Validator) validateByType(item map[string]interface{}, entityType string, result ValidationResult) ValidationResult {
	switch entityType {
	case "chorki_customers":
		return v.validateCustomer(item, result)
	case "chorki_subscriptions":
		return v.validateSubscription(item, result)
	case "chorki_devices":
		return v.validateDevice(item, result)
	case "chorki_watch_histories":
		return v.validateWatchHistory(item, result)
	case "chorki_contents":
		return v.validateContent(item, result)
	default:
		result.Warnings = append(result.Warnings, fmt.Sprintf("unknown entity type: %s", entityType))
	}

	return result
}

// validateCustomer validates customer data
func (v *Validator) validateCustomer(item map[string]interface{}, result ValidationResult) ValidationResult {
	// Check required fields
	requiredFields := []string{
		"chorki_customers.id",
		"chorki_customers.name",
		"chorki_customers.email",
		"chorki_customers.age",
	}

	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate email format
	if email, exists := item["chorki_customers.email"]; exists {
		if emailStr, ok := email.(string); ok {
			if !strings.Contains(emailStr, "@") {
				result.Valid = false
				result.Errors = append(result.Errors, "invalid email format")
			}
		}
	}

	// Validate age range
	if age, exists := item["chorki_customers.age"]; exists {
		if ageVal, ok := age.(float64); ok {
			if ageVal < 13 || ageVal > 120 {
				result.Valid = false
				result.Errors = append(result.Errors, "age must be between 13 and 120")
			}
		} else if ageVal, ok := age.(int); ok {
			if ageVal < 13 || ageVal > 120 {
				result.Valid = false
				result.Errors = append(result.Errors, "age must be between 13 and 120")
			}
		}
	}

	return result
}

// validateSubscription validates subscription data
func (v *Validator) validateSubscription(item map[string]interface{}, result ValidationResult) ValidationResult {
	requiredFields := []string{
		"chorki_subscriptions.id",
		"chorki_subscriptions.package",
		"chorki_subscriptions.status",
	}

	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate status
	if status, exists := item["chorki_subscriptions.status"]; exists {
		if statusStr, ok := status.(string); ok {
			validStatuses := []string{"active", "inactive", "trial", "expired", "suspended", "cancelled"}
			isValid := false
			for _, validStatus := range validStatuses {
				if statusStr == validStatus {
					isValid = true
					break
				}
			}
			if !isValid {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("invalid status: %s", statusStr))
			}
		}
	}

	return result
}

// validateDevice validates device data
func (v *Validator) validateDevice(item map[string]interface{}, result ValidationResult) ValidationResult {
	requiredFields := []string{
		"chorki_devices.id",
		"chorki_devices.device_type",
	}

	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return result
}

// validateWatchHistory validates watch history data
func (v *Validator) validateWatchHistory(item map[string]interface{}, result ValidationResult) ValidationResult {
	requiredFields := []string{
		"chorki_watch_histories.id",
		"chorki_watch_histories.content_id",
	}

	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return result
}

// validateContent validates content data
func (v *Validator) validateContent(item map[string]interface{}, result ValidationResult) ValidationResult {
	requiredFields := []string{
		"chorki_contents.id",
		"chorki_contents.title",
		"chorki_contents.type",
	}

	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return result
}

// ValidateType validates that a value matches expected type
func (v *Validator) ValidateType(value interface{}, expectedType string) bool {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return expectedType == "nil"
	}

	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		_, ok := value.(int)
		if !ok {
			_, ok = value.(float64) // JSON numbers come as float64
		}
		return ok
	case "float":
		_, ok := value.(float64)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "array":
		return valueType.Kind() == reflect.Slice
	case "object":
		return valueType.Kind() == reflect.Map
	default:
		return false
	}
}
