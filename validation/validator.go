package validation

import (
	"fmt"
	"jsonTodql/models"
	"reflect"
	"strconv"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationResult contains the validation results
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// QueryValidator validates JSON queries
type QueryValidator struct {
	schema           *models.SchemaInfo
	supportedOps     map[string]bool
	dataTypeValidators map[string]func(interface{}) bool
}

// NewQueryValidator creates a new query validator
func NewQueryValidator(schema *models.SchemaInfo) *QueryValidator {
	validator := &QueryValidator{
		schema: schema,
		supportedOps: map[string]bool{
			"=": true, ">=": true, "<=": true, ">": true, "<": true,
			"IN": true, "NOT_IN": true, "!=": true, "LIKE": true,
			"ILIKE": true, "REGEX": true, "BETWEEN": true,
			"IS_NULL": true, "IS_NOT_NULL": true, "STARTS_WITH": true,
			"ENDS_WITH": true, "CONTAINS": true,
		},
		dataTypeValidators: make(map[string]func(interface{}) bool),
	}
	
	validator.initDataTypeValidators()
	return validator
}

// Validate performs comprehensive validation on a JSON query
func (v *QueryValidator) Validate(query *models.JSONQuery) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Basic structure validation
	v.validateStructure(query, result)
	
	// Field existence validation
	v.validateFields(query.Groups, result)
	
	// Data type compatibility validation
	v.validateDataTypes(query.Groups, result)
	
	// Operator validation
	v.validateOperators(query.Groups, result)
	
	// Logical consistency validation
	v.validateLogicalConsistency(query, result)
	
	// Complex object validation
	v.validateComplexObjects(query.Groups, result)

	result.IsValid = len(result.Errors) == 0
	return result
}

// validateStructure validates the basic structure of the query
func (v *QueryValidator) validateStructure(query *models.JSONQuery, result *ValidationResult) {
	if query.CombineWith != "AND" && query.CombineWith != "OR" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "combine_with",
			Message: "combine_with must be either 'AND' or 'OR'",
			Code:    "INVALID_COMBINE_OPERATOR",
		})
	}

	if len(query.Groups) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "groups",
			Message: "at least one group is required",
			Code:    "MISSING_GROUPS",
		})
	}

	v.validateGroups(query.Groups, result, "groups")
}

// validateGroups recursively validates groups
func (v *QueryValidator) validateGroups(groups []models.Group, result *ValidationResult, path string) {
	for i, group := range groups {
		groupPath := fmt.Sprintf("%s[%d]", path, i)
		
		if group.CombineWith != "AND" && group.CombineWith != "OR" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("%s.combine_with", groupPath),
				Message: "group combine_with must be either 'AND' or 'OR'",
				Code:    "INVALID_GROUP_COMBINE_OPERATOR",
			})
		}

		if len(group.Filters) == 0 && len(group.Groups) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   groupPath,
				Message: "group must have either filters or nested groups",
				Code:    "EMPTY_GROUP",
			})
		}

		// Validate filters in this group
		for j, filter := range group.Filters {
			filterPath := fmt.Sprintf("%s.filters[%d]", groupPath, j)
			v.validateFilter(filter, result, filterPath)
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			v.validateGroups(group.Groups, result, fmt.Sprintf("%s.groups", groupPath))
		}
	}
}

// validateFilter validates an individual filter
func (v *QueryValidator) validateFilter(filter models.Filter, result *ValidationResult, path string) {
	if filter.Field == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.field", path),
			Message: "filter field cannot be empty",
			Code:    "MISSING_FIELD",
		})
	}

	if filter.Op == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.op", path),
			Message: "filter operator cannot be empty",
			Code:    "MISSING_OPERATOR",
		})
	}

	if filter.Value == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.value", path),
			Message: "filter value cannot be null",
			Code:    "MISSING_VALUE",
		})
	}
}

// validateFields validates that all fields exist in the schema
func (v *QueryValidator) validateFields(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			if _, exists := v.schema.FieldMappings[filter.Field]; !exists {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: fmt.Sprintf("field '%s' is not defined in the schema", filter.Field),
					Code:    "UNKNOWN_FIELD",
				})
			}
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			v.validateFields(group.Groups, result)
		}
	}
}

// validateDataTypes validates data type compatibility
func (v *QueryValidator) validateDataTypes(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			mappings, exists := v.schema.FieldMappings[filter.Field]
			if !exists {
				continue // Field validation will catch this
			}

			for _, mapping := range mappings {
				if !v.isValueCompatibleWithDataType(filter.Value, mapping.DataType, filter.Op) {
					result.Errors = append(result.Errors, ValidationError{
						Field:   filter.Field,
						Message: fmt.Sprintf("value type incompatible with field data type '%s'", mapping.DataType),
						Code:    "INCOMPATIBLE_DATA_TYPE",
					})
				}
			}
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			v.validateDataTypes(group.Groups, result)
		}
	}
}

// validateOperators validates that operators are supported and compatible
func (v *QueryValidator) validateOperators(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			if !v.supportedOps[filter.Op] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: fmt.Sprintf("operator '%s' is not supported", filter.Op),
					Code:    "UNSUPPORTED_OPERATOR",
				})
				continue
			}

			// Validate operator-specific requirements
			if err := v.validateOperatorRequirements(filter); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: err.Error(),
					Code:    "INVALID_OPERATOR_USAGE",
				})
			}
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			v.validateOperators(group.Groups, result)
		}
	}
}

// validateLogicalConsistency validates logical consistency
func (v *QueryValidator) validateLogicalConsistency(query *models.JSONQuery, result *ValidationResult) {
	// Check for contradictory conditions on the same field
	fieldConditions := v.extractFieldConditions(query.Groups)
	
	for field, conditions := range fieldConditions {
		if v.hasContradictoryConditions(conditions) {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("potentially contradictory conditions for field '%s'", field),
				Code:    "CONTRADICTORY_CONDITIONS",
			})
		}
	}
}

// validateComplexObjects validates complex object structures
func (v *QueryValidator) validateComplexObjects(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			if filter.Field == "watched_content" {
				v.validateWatchedContentObject(filter, result)
			}
		}

		// Recursively validate nested groups
		if len(group.Groups) > 0 {
			v.validateComplexObjects(group.Groups, result)
		}
	}
}

// Helper methods

func (v *QueryValidator) initDataTypeValidators() {
	v.dataTypeValidators["string"] = func(value interface{}) bool {
		_, ok := value.(string)
		return ok
	}
	
	v.dataTypeValidators["int"] = func(value interface{}) bool {
		switch value.(type) {
		case int, int64, float64:
			return true
		case string:
			_, err := strconv.Atoi(value.(string))
			return err == nil
		}
		return false
	}
	
	v.dataTypeValidators["float"] = func(value interface{}) bool {
		switch value.(type) {
		case float64, int, int64:
			return true
		case string:
			_, err := strconv.ParseFloat(value.(string), 64)
			return err == nil
		}
		return false
	}
	
	v.dataTypeValidators["bool"] = func(value interface{}) bool {
		_, ok := value.(bool)
		return ok
	}
	
	v.dataTypeValidators["datetime"] = func(value interface{}) bool {
		if str, ok := value.(string); ok {
			_, err := time.Parse(time.RFC3339, str)
			return err == nil
		}
		return false
	}
}

func (v *QueryValidator) isValueCompatibleWithDataType(value interface{}, dataType, operator string) bool {
	// Special cases for certain operators
	switch operator {
	case "IS_NULL", "IS_NOT_NULL":
		return true // These operators don't require a value
	case "IN", "NOT_IN":
		// For IN/NOT_IN, value should be an array
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if validator, exists := v.dataTypeValidators[dataType]; exists {
					if !validator(item) {
						return false
					}
				}
			}
			return true
		}
		return false
	case "BETWEEN":
		// For BETWEEN, value should be an array of 2 elements or an object with min/max
		if arr, ok := value.([]interface{}); ok {
			if len(arr) != 2 {
				return false
			}
			if validator, exists := v.dataTypeValidators[dataType]; exists {
				return validator(arr[0]) && validator(arr[1])
			}
		}
		if obj, ok := value.(map[string]interface{}); ok {
			min, hasMin := obj["min"]
			max, hasMax := obj["max"]
			if !hasMin || !hasMax {
				return false
			}
			if validator, exists := v.dataTypeValidators[dataType]; exists {
				return validator(min) && validator(max)
			}
		}
		return false
	default:
		// For other operators, validate the value directly
		if validator, exists := v.dataTypeValidators[dataType]; exists {
			return validator(value)
		}
	}
	
	return true // Default to true if no specific validator
}

func (v *QueryValidator) validateOperatorRequirements(filter models.Filter) error {
	switch filter.Op {
	case "IN", "NOT_IN":
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
			return fmt.Errorf("IN/NOT_IN operators require array value")
		}
	case "BETWEEN":
		switch v := filter.Value.(type) {
		case []interface{}:
			if len(v) != 2 {
				return fmt.Errorf("BETWEEN operator requires array with exactly 2 elements")
			}
		case map[string]interface{}:
			if _, hasMin := v["min"]; !hasMin {
				return fmt.Errorf("BETWEEN operator requires 'min' property")
			}
			if _, hasMax := v["max"]; !hasMax {
				return fmt.Errorf("BETWEEN operator requires 'max' property")
			}
		default:
			return fmt.Errorf("BETWEEN operator requires array or object with min/max")
		}
	case "REGEX":
		if _, ok := filter.Value.(string); !ok {
			return fmt.Errorf("REGEX operator requires string value")
		}
	}
	return nil
}

func (v *QueryValidator) extractFieldConditions(groups []models.Group) map[string][]models.Filter {
	conditions := make(map[string][]models.Filter)
	
	for _, group := range groups {
		for _, filter := range group.Filters {
			conditions[filter.Field] = append(conditions[filter.Field], filter)
		}
		
		// Recursively extract from nested groups
		nestedConditions := v.extractFieldConditions(group.Groups)
		for field, filters := range nestedConditions {
			conditions[field] = append(conditions[field], filters...)
		}
	}
	
	return conditions
}

func (v *QueryValidator) hasContradictoryConditions(filters []models.Filter) bool {
	// Simple contradiction detection
	// More sophisticated logic can be added here
	for i, filter1 := range filters {
		for j, filter2 := range filters {
			if i >= j {
				continue
			}
			
			// Check for obvious contradictions like field = 1 AND field = 2
			if filter1.Op == "=" && filter2.Op == "=" && filter1.Value != filter2.Value {
				return true
			}
		}
	}
	return false
}

func (v *QueryValidator) validateWatchedContentObject(filter models.Filter, result *ValidationResult) {
	obj, ok := filter.Value.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content",
			Message: "watched_content value must be an object",
			Code:    "INVALID_COMPLEX_OBJECT",
		})
		return
	}

	// Validate required properties
	if _, hasType := obj["content_type"]; !hasType {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content.content_type",
			Message: "content_type is required in watched_content object",
			Code:    "MISSING_REQUIRED_PROPERTY",
		})
	}

	if _, hasIds := obj["ids"]; !hasIds {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content.ids",
			Message: "ids is required in watched_content object",
			Code:    "MISSING_REQUIRED_PROPERTY",
		})
	}

	// Validate ids is an array
	if ids, ok := obj["ids"]; ok {
		if _, isArray := ids.([]interface{}); !isArray {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "watched_content.ids",
				Message: "ids must be an array",
				Code:    "INVALID_PROPERTY_TYPE",
			})
		}
	}
}