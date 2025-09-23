// Package validation provides comprehensive query validation capabilities for JSON to DQL conversion.
// This package ensures data integrity, type safety, and logical consistency of incoming queries
// before they are processed by the converter. It implements multi-layer validation including
// structural validation, schema compliance, data type compatibility, and business logic validation.
package validation

import (
	"fmt"
	"jsonTodql/models"
	"reflect"
	"strconv"
	"time"
)

// =============================================================================
// VALIDATION RESULT STRUCTURES
// =============================================================================

// ValidationError represents a specific validation failure with contextual information.
// This structure provides detailed error reporting to help developers identify and fix
// query issues quickly. The Code field enables programmatic error handling.
type ValidationError struct {
	Field   string `json:"field"`   // The field path where the error occurred
	Message string `json:"message"` // Human-readable error description
	Code    string `json:"code"`    // Machine-readable error code for programmatic handling
}

// ValidationResult contains comprehensive validation results including errors and warnings.
// This structure provides a complete picture of query validation status, enabling
// both strict validation (errors) and advisory feedback (warnings) for query optimization.
type ValidationResult struct {
	IsValid  bool              `json:"is_valid"`           // Overall validation status
	Errors   []ValidationError `json:"errors"`             // Critical validation failures
	Warnings []ValidationError `json:"warnings,omitempty"` // Non-critical recommendations
}

// =============================================================================
// CORE VALIDATOR STRUCTURE
// =============================================================================

// QueryValidator provides comprehensive validation capabilities for JSON query structures.
// This validator implements multi-layer validation including structural integrity,
// schema compliance, data type compatibility, operator validation, and business logic checks.
// It serves as the primary gatekeeper ensuring only valid, safe queries reach the converter.
type QueryValidator struct {
	schema             *models.SchemaInfo                // Schema definition for field validation
	supportedOps       map[string]bool                   // Supported query operators
	dataTypeValidators map[string]func(interface{}) bool // Data type validation functions
}

// =============================================================================
// CONSTRUCTOR AND INITIALIZATION
// =============================================================================

// NewQueryValidator creates a new query validator with comprehensive validation capabilities.
// This constructor initializes all validation components including operator support,
// data type validators, and schema integration for complete query validation.
func NewQueryValidator(schema *models.SchemaInfo) *QueryValidator {
	validator := &QueryValidator{
		schema: schema,
		// Define comprehensive operator support for all query operations
		supportedOps: map[string]bool{
			"=": true, ">=": true, "<=": true, ">": true, "<": true,
			"IN": true, "NOT_IN": true, "!=": true, "LIKE": true,
			"ILIKE": true, "REGEX": true, "BETWEEN": true,
			"IS_NULL": true, "IS_NOT_NULL": true, "STARTS_WITH": true,
			"ENDS_WITH": true, "CONTAINS": true,
		},
		dataTypeValidators: make(map[string]func(interface{}) bool),
	}

	// Initialize data type validation functions for type safety
	validator.initDataTypeValidators()
	return validator
}

// =============================================================================
// MAIN VALIDATION ENTRY POINT
// =============================================================================

// Validate performs comprehensive multi-layer validation on a JSON query structure.
// This method implements a systematic validation approach covering structural integrity,
// schema compliance, data type compatibility, operator validation, logical consistency,
// and business-specific validation rules. It provides detailed error reporting and
// advisory warnings to ensure query quality and execution safety.
func (v *QueryValidator) Validate(query *models.JSONQuery) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Layer 1: Basic structural validation (syntax and required fields)
	v.validateStructure(query, result)

	// Layer 2: Schema compliance validation (field existence and mapping)
	v.validateFields(query.Groups, result)

	// Layer 3: Data type compatibility validation (type safety)
	v.validateDataTypes(query.Groups, result)

	// Layer 4: Operator validation (supported operations and requirements)
	v.validateOperators(query.Groups, result)

	// Layer 5: Logical consistency validation (contradiction detection)
	v.validateLogicalConsistency(query, result)

	// Layer 6: Complex object validation (business-specific structures)
	v.validateComplexObjects(query.Groups, result)

	// Final validation status determination
	result.IsValid = len(result.Errors) == 0
	return result
}

// =============================================================================
// STRUCTURAL VALIDATION LAYER
// =============================================================================

// validateStructure validates the fundamental structure and syntax of the query.
// This method ensures the query has proper logical operators, required groups,
// and correct hierarchical structure before proceeding to deeper validation layers.
func (v *QueryValidator) validateStructure(query *models.JSONQuery, result *ValidationResult) {
	// Validate root-level logical operator
	if query.CombineWith != "AND" && query.CombineWith != "OR" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "combine_with",
			Message: "combine_with must be either 'AND' or 'OR'",
			Code:    "INVALID_COMBINE_OPERATOR",
		})
	}

	// Ensure at least one group exists for query execution
	if len(query.Groups) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "groups",
			Message: "at least one group is required",
			Code:    "MISSING_GROUPS",
		})
	}

	// Recursively validate all group structures
	v.validateGroups(query.Groups, result, "groups")
}

// validateGroups recursively validates the structure of group hierarchies.
// This method ensures proper nesting, logical operators, and content requirements
// for all groups in the query structure, maintaining validation context through path tracking.
func (v *QueryValidator) validateGroups(groups []models.Group, result *ValidationResult, path string) {
	for i, group := range groups {
		groupPath := fmt.Sprintf("%s[%d]", path, i)

		// Validate group-level logical operator
		if group.CombineWith != "AND" && group.CombineWith != "OR" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("%s.combine_with", groupPath),
				Message: "group combine_with must be either 'AND' or 'OR'",
				Code:    "INVALID_GROUP_COMBINE_OPERATOR",
			})
		}

		// Ensure group has content (either filters or nested groups)
		if len(group.Filters) == 0 && len(group.Groups) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   groupPath,
				Message: "group must have either filters or nested groups",
				Code:    "EMPTY_GROUP",
			})
		}

		// Validate individual filters within this group
		for j, filter := range group.Filters {
			filterPath := fmt.Sprintf("%s.filters[%d]", groupPath, j)
			v.validateFilter(filter, result, filterPath)
		}

		// Recursively validate nested group structures
		if len(group.Groups) > 0 {
			v.validateGroups(group.Groups, result, fmt.Sprintf("%s.groups", groupPath))
		}
	}
}

// validateFilter validates individual filter components for completeness and basic syntax.
// This method ensures each filter has required fields (field, operator, value) and
// provides specific error paths for precise debugging and error reporting.
func (v *QueryValidator) validateFilter(filter models.Filter, result *ValidationResult, path string) {
	// Validate field name presence
	if filter.Field == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.field", path),
			Message: "filter field cannot be empty",
			Code:    "MISSING_FIELD",
		})
	}

	// Validate operator presence
	if filter.Op == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.op", path),
			Message: "filter operator cannot be empty",
			Code:    "MISSING_OPERATOR",
		})
	}

	// Validate value presence (null values may be valid for some operators)
	if filter.Value == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("%s.value", path),
			Message: "filter value cannot be null",
			Code:    "MISSING_VALUE",
		})
	}
}

// =============================================================================
// SCHEMA COMPLIANCE VALIDATION LAYER
// =============================================================================

// validateFields validates that all referenced fields exist in the schema definition.
// This method ensures queries only reference valid, defined fields and provides
// clear error reporting for undefined field usage, preventing runtime failures.
func (v *QueryValidator) validateFields(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		// Validate each filter's field against schema
		for _, filter := range group.Filters {
			if _, exists := v.schema.FieldMappings[filter.Field]; !exists {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: fmt.Sprintf("field '%s' is not defined in the schema", filter.Field),
					Code:    "UNKNOWN_FIELD",
				})
			}
		}

		// Recursively validate nested group fields
		if len(group.Groups) > 0 {
			v.validateFields(group.Groups, result)
		}
	}
}

// =============================================================================
// DATA TYPE COMPATIBILITY VALIDATION LAYER
// =============================================================================

// validateDataTypes validates data type compatibility between filter values and schema definitions.
// This method ensures type safety by checking that filter values are compatible with
// their corresponding field data types, preventing type-related query execution errors.
func (v *QueryValidator) validateDataTypes(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			// Skip validation if field doesn't exist (handled by field validation)
			mappings, exists := v.schema.FieldMappings[filter.Field]
			if !exists {
				continue
			}

			// Validate value compatibility with all field mappings
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

		// Recursively validate nested group data types
		if len(group.Groups) > 0 {
			v.validateDataTypes(group.Groups, result)
		}
	}
}

// =============================================================================
// OPERATOR VALIDATION LAYER
// =============================================================================

// validateOperators validates that operators are supported and meet their specific requirements.
// This method ensures operators are recognized by the system and that their usage
// conforms to required patterns (e.g., IN requires arrays, BETWEEN requires ranges).
func (v *QueryValidator) validateOperators(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			// Check if operator is supported
			if !v.supportedOps[filter.Op] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: fmt.Sprintf("operator '%s' is not supported", filter.Op),
					Code:    "UNSUPPORTED_OPERATOR",
				})
				continue
			}

			// Validate operator-specific value requirements
			if err := v.validateOperatorRequirements(filter); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:   filter.Field,
					Message: err.Error(),
					Code:    "INVALID_OPERATOR_USAGE",
				})
			}
		}

		// Recursively validate nested group operators
		if len(group.Groups) > 0 {
			v.validateOperators(group.Groups, result)
		}
	}
}

// =============================================================================
// LOGICAL CONSISTENCY VALIDATION LAYER
// =============================================================================

// validateLogicalConsistency validates logical consistency and detects potential contradictions.
// This method analyzes the query for logical conflicts that might result in no matches
// or unexpected behavior, providing warnings for potential optimization opportunities.
func (v *QueryValidator) validateLogicalConsistency(query *models.JSONQuery, result *ValidationResult) {
	// Extract all conditions for each field across the entire query
	fieldConditions := v.extractFieldConditions(query.Groups)

	// Analyze each field's conditions for potential contradictions
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

// =============================================================================
// COMPLEX OBJECT VALIDATION LAYER
// =============================================================================

// validateComplexObjects validates complex object structures and business-specific requirements.
// This method handles specialized validation for complex data structures like watched_content
// that have specific structural and business logic requirements beyond basic type checking.
func (v *QueryValidator) validateComplexObjects(groups []models.Group, result *ValidationResult) {
	for _, group := range groups {
		for _, filter := range group.Filters {
			// Validate watched_content complex object structure
			if filter.Field == "watched_content" {
				v.validateWatchedContentObject(filter, result)
			}
			// Additional complex object validations can be added here
		}

		// Recursively validate nested group complex objects
		if len(group.Groups) > 0 {
			v.validateComplexObjects(group.Groups, result)
		}
	}
}

// =============================================================================
// DATA TYPE VALIDATION INITIALIZATION
// =============================================================================

// initDataTypeValidators initializes type-specific validation functions for all supported data types.
// This method sets up comprehensive type checking capabilities that handle Go's type system
// nuances and provide flexible type coercion for common data type conversions.
func (v *QueryValidator) initDataTypeValidators() {
	// String type validator - accepts native strings
	v.dataTypeValidators["string"] = func(value interface{}) bool {
		_, ok := value.(string)
		return ok
	}

	// Integer type validator - handles multiple numeric types and string conversion
	v.dataTypeValidators["int"] = func(value interface{}) bool {
		switch v := value.(type) {
		case int, int64, float64:
			return true
		case string:
			// Allow string-to-int conversion if the string represents a valid integer
			_, err := strconv.Atoi(v)
			return err == nil
		}
		return false
	}

	// Float type validator - handles numeric types and string conversion
	v.dataTypeValidators["float"] = func(value interface{}) bool {
		switch v := value.(type) {
		case float64, int, int64:
			return true
		case string:
			// Allow string-to-float conversion if the string represents a valid number
			_, err := strconv.ParseFloat(v, 64)
			return err == nil
		}
		return false
	}

	// Boolean type validator - strict boolean type checking
	v.dataTypeValidators["bool"] = func(value interface{}) bool {
		_, ok := value.(bool)
		return ok
	}

	// DateTime type validator - expects RFC3339 formatted strings
	v.dataTypeValidators["datetime"] = func(value interface{}) bool {
		if str, ok := value.(string); ok {
			// Validate against RFC3339 standard (ISO 8601)
			_, err := time.Parse(time.RFC3339, str)
			return err == nil
		}
		return false
	}
}

// =============================================================================
// VALUE COMPATIBILITY VALIDATION
// =============================================================================

// isValueCompatibleWithDataType determines if a filter value is compatible with the field's data type.
// This method implements sophisticated type checking that considers operator-specific requirements
// and handles special cases like arrays for IN operations and ranges for BETWEEN operations.
func (v *QueryValidator) isValueCompatibleWithDataType(value interface{}, dataType, operator string) bool {
	// Handle operators that don't require specific values
	switch operator {
	case "IS_NULL", "IS_NOT_NULL":
		return true // These operators check field presence, not value content
	case "IN", "NOT_IN":
		// IN/NOT_IN operations require array values with compatible element types
		if arr, ok := value.([]interface{}); ok {
			// Validate each array element against the field's data type
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
		// BETWEEN operations support both array [min, max] and object {min, max} formats
		if arr, ok := value.([]interface{}); ok {
			// Array format: exactly 2 elements required
			if len(arr) != 2 {
				return false
			}
			if validator, exists := v.dataTypeValidators[dataType]; exists {
				return validator(arr[0]) && validator(arr[1])
			}
		}
		if obj, ok := value.(map[string]interface{}); ok {
			// Object format: must have both min and max properties
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
		// Standard operators: validate single value directly
		if validator, exists := v.dataTypeValidators[dataType]; exists {
			return validator(value)
		}
	}

	// Default to true for unknown data types (permissive approach)
	return true
}

// =============================================================================
// OPERATOR REQUIREMENT VALIDATION
// =============================================================================

// validateOperatorRequirements validates operator-specific value format requirements.
// This method ensures that complex operators like IN, BETWEEN, and REGEX receive
// values in the correct format and structure for successful query execution.
func (v *QueryValidator) validateOperatorRequirements(filter models.Filter) error {
	switch filter.Op {
	case "IN", "NOT_IN":
		// IN/NOT_IN operations must receive array values
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
			return fmt.Errorf("IN/NOT_IN operators require array value")
		}
	case "BETWEEN":
		// BETWEEN operations support flexible value formats
		switch v := filter.Value.(type) {
		case []interface{}:
			// Array format: must contain exactly 2 elements [min, max]
			if len(v) != 2 {
				return fmt.Errorf("BETWEEN operator requires array with exactly 2 elements")
			}
		case map[string]interface{}:
			// Object format: must contain both min and max properties
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
		// REGEX operations require string pattern values
		if _, ok := filter.Value.(string); !ok {
			return fmt.Errorf("REGEX operator requires string value")
		}
	}
	return nil
}

// =============================================================================
// LOGICAL CONSISTENCY ANALYSIS
// =============================================================================

// extractFieldConditions extracts and groups all filter conditions by field name across the query.
// This method performs a comprehensive traversal of the query structure to collect all
// conditions applied to each field, enabling logical consistency analysis and contradiction detection.
func (v *QueryValidator) extractFieldConditions(groups []models.Group) map[string][]models.Filter {
	conditions := make(map[string][]models.Filter)

	for _, group := range groups {
		// Collect filters from the current group
		for _, filter := range group.Filters {
			conditions[filter.Field] = append(conditions[filter.Field], filter)
		}

		// Recursively collect from nested groups
		nestedConditions := v.extractFieldConditions(group.Groups)
		for field, filters := range nestedConditions {
			conditions[field] = append(conditions[field], filters...)
		}
	}

	return conditions
}

// hasContradictoryConditions analyzes a set of filters for logical contradictions.
// This method implements basic contradiction detection for obvious conflicts that
// would result in empty result sets. More sophisticated analysis can be added for
// complex logical scenarios and business rule validation.
func (v *QueryValidator) hasContradictoryConditions(filters []models.Filter) bool {
	// Detect simple equality contradictions (field = A AND field = B where A != B)
	for i, filter1 := range filters {
		for j, filter2 := range filters {
			if i >= j {
				continue // Avoid duplicate comparisons
			}

			// Check for obvious contradictions: same field with different equality values
			if filter1.Op == "=" && filter2.Op == "=" && filter1.Value != filter2.Value {
				return true
			}

			// Additional contradiction patterns can be added here:
			// - Range overlaps (field > 10 AND field < 5)
			// - Inclusion/exclusion conflicts (field IN [1,2] AND field NOT_IN [1,2])
			// - Null/non-null conflicts (field IS_NULL AND field = value)
		}
	}
	return false
}

// =============================================================================
// BUSINESS-SPECIFIC COMPLEX OBJECT VALIDATION
// =============================================================================

// validateWatchedContentObject validates the structure and content of watched_content objects.
// This method implements business-specific validation for the watched_content complex object,
// ensuring it contains required properties and follows the expected data structure for
// content filtering and user behavior analysis.
func (v *QueryValidator) validateWatchedContentObject(filter models.Filter, result *ValidationResult) {
	// Ensure the value is a proper object structure
	obj, ok := filter.Value.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content",
			Message: "watched_content value must be an object",
			Code:    "INVALID_COMPLEX_OBJECT",
		})
		return
	}

	// Validate required property: content_type
	if _, hasType := obj["content_type"]; !hasType {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content.content_type",
			Message: "content_type is required in watched_content object",
			Code:    "MISSING_REQUIRED_PROPERTY",
		})
	}

	// Validate required property: ids
	if _, hasIds := obj["ids"]; !hasIds {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "watched_content.ids",
			Message: "ids is required in watched_content object",
			Code:    "MISSING_REQUIRED_PROPERTY",
		})
	}

	// Validate that ids property is an array when present
	if ids, ok := obj["ids"]; ok {
		if _, isArray := ids.([]interface{}); !isArray {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "watched_content.ids",
				Message: "ids must be an array",
				Code:    "INVALID_PROPERTY_TYPE",
			})
		}
	}

	// Additional business validation can be added here:
	// - Validate content_type values against allowed types
	// - Validate ID format and ranges
	// - Cross-reference with content availability
}
