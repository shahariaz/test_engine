package converter

import (
	"fmt"
	"jsonTodql/analyzer"
	"jsonTodql/config"
	"jsonTodql/models"
	"jsonTodql/utils"
	"strconv"
	"strings"
)

// Converter handles the conversion from JSON query to DQL
type Converter struct {
	schema             *models.SchemaInfo
	operators          map[string]string
	complexityAnalyzer *analyzer.ComplexityAnalyzer
	versionFields      map[string]string
	reversePredicates  map[string]string
}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	schema := config.GetSchemaConfig()
	return &Converter{
		schema:             schema,
		operators:          config.GetOperatorMappings(),
		complexityAnalyzer: analyzer.NewComplexityAnalyzer(schema.FieldMappings),
		versionFields:      config.GetVersionFields(),
		reversePredicates:  config.GetReversePredicates(),
	}
}

// ConvertToDQL converts a JSON query to DQL format with complexity analysis
func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.DQLQuery, error) {
	// Analyze query complexity first
	complexityScore := c.complexityAnalyzer.AnalyzeComplexity(jsonQuery)
	if !complexityScore.IsAcceptable {
		return nil, fmt.Errorf("query too complex: %s", complexityScore.Warning)
	}

	// Get all entity types involved in the query
	involvedEntities := c.getInvolvedEntityTypes(jsonQuery)

	// Generate queries for each involved entity type
	var queries []models.EntityQuery
	for _, entityType := range involvedEntities {
		filter, err := c.buildFilterForEntity(jsonQuery, entityType)
		if err != nil {
			return nil, fmt.Errorf("error building filter for %s: %v", entityType, err)
		}

		if filter != "" {
			query := models.EntityQuery{
				Name:     c.getQueryName(entityType),
				Type:     entityType,
				Function: fmt.Sprintf("type(%s)", entityType),
				Filter:   filter,
				Fields:   c.buildFieldsSelection(entityType),
			}
			queries = append(queries, query)
		}
	}

	return &models.DQLQuery{Queries: queries}, nil
}

// getPrimaryEntity determines the primary entity type based on field frequency
func (c *Converter) getPrimaryEntity(jsonQuery *models.JSONQuery) string {
	entityCount := make(map[string]int)

	// Count field occurrences for each entity type
	c.countEntityTypesFromGroups(jsonQuery.Groups, entityCount)

	// Find the entity with the most fields
	var primaryEntity string
	maxCount := 0

	// Prioritize customers as the default primary entity
	if count, exists := entityCount["chorki_customers"]; exists && count > 0 {
		primaryEntity = "chorki_customers"
		maxCount = count
	}

	// Check if any other entity has significantly more fields
	for entityType, count := range entityCount {
		if count > maxCount {
			primaryEntity = entityType
			maxCount = count
		}
	}

	// Default to customers if no clear primary entity
	if primaryEntity == "" {
		primaryEntity = "chorki_customers"
	}

	return primaryEntity
}

// countEntityTypesFromGroups recursively counts entity types from groups
func (c *Converter) countEntityTypesFromGroups(groups []models.Group, entityCount map[string]int) {
	for _, group := range groups {
		// Count filters in this group
		for _, filter := range group.Filters {
			if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {
				for _, mapping := range mappings {
					entityCount[mapping.EntityType]++
				}
			}
		}

		// Recursively count nested groups
		if len(group.Groups) > 0 {
			c.countEntityTypesFromGroups(group.Groups, entityCount)
		}
	}
}

// getInvolvedEntityTypes determines which entity types are referenced in the query
func (c *Converter) getInvolvedEntityTypes(jsonQuery *models.JSONQuery) []string {
	entityTypeMap := make(map[string]bool)

	// Recursively check all groups and filters
	c.collectEntityTypesFromGroups(jsonQuery.Groups, entityTypeMap)

	var result []string
	for entityType := range entityTypeMap {
		result = append(result, entityType)
	}

	// If no entities found, default to customers
	if len(result) == 0 {
		result = append(result, "chorki_customers")
	}

	return result
}

// collectEntityTypesFromGroups recursively collects entity types from groups
func (c *Converter) collectEntityTypesFromGroups(groups []models.Group, entityTypeMap map[string]bool) {
	for _, group := range groups {
		// Check filters in this group
		for _, filter := range group.Filters {
			if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {
				for _, mapping := range mappings {
					entityTypeMap[mapping.EntityType] = true
				}
			}
		}

		// Recursively check nested groups
		if len(group.Groups) > 0 {
			c.collectEntityTypesFromGroups(group.Groups, entityTypeMap)
		}
	}
}

// buildFilterForEntity builds the DQL filter string for a specific entity type
func (c *Converter) buildFilterForEntity(jsonQuery *models.JSONQuery, entityType string) (string, error) {
	filter := c.buildGroupsFilter(jsonQuery.Groups, jsonQuery.CombineWith, entityType)
	if filter == "" {
		return "", nil
	}
	return fmt.Sprintf("@filter(%s)", filter), nil
}

// buildGroupsFilter builds filter string for a list of groups
func (c *Converter) buildGroupsFilter(groups []models.Group, combineWith, entityType string) string {
	var conditions []string

	for _, group := range groups {
		condition := c.buildGroupFilter(group, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	operator := " AND "
	if strings.ToUpper(combineWith) == "OR" {
		operator = " OR "
	}

	return "(" + strings.Join(conditions, operator) + ")"
}

// buildGroupFilter builds filter string for a single group
func (c *Converter) buildGroupFilter(group models.Group, entityType string) string {
	var conditions []string

	// Build conditions from filters
	for _, filter := range group.Filters {
		condition := c.buildFilterCondition(filter, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	// Apply subscription filter optimization if applicable
	if entityType == "chorki_subscriptions" {
		conditions = c.optimizeSubscriptionFilters(conditions, group)
	}

	// Build conditions from nested groups
	if len(group.Groups) > 0 {
		nestedCondition := c.buildGroupsFilter(group.Groups, group.CombineWith, entityType)
		if nestedCondition != "" {
			conditions = append(conditions, nestedCondition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	operator := " AND "
	if strings.ToUpper(group.CombineWith) == "OR" {
		operator = " OR "
	}

	return "(" + strings.Join(conditions, operator) + ")"
}

// optimizeSubscriptionFilters optimizes subscription filters to avoid overly restrictive conditions
func (c *Converter) optimizeSubscriptionFilters(conditions []string, group models.Group) []string {
	// Check for Premium + trial combination and suggest optimization
	var hasPackagePremium, hasStatusTrial bool
	var trialIndex int

	for i, condition := range conditions {
		if strings.Contains(condition, `eq(chorki_subscriptions.package, "Premium")`) {
			hasPackagePremium = true
		}
		if strings.Contains(condition, `eq(chorki_subscriptions.status, "trial")`) {
			hasStatusTrial = true
			trialIndex = i
		}
	}

	// If we have Premium package AND trial status with AND operator, suggest optimization
	if hasPackagePremium && hasStatusTrial && strings.ToUpper(group.CombineWith) == "AND" {
		// Replace the trial condition with a more inclusive status condition
		optimizedConditions := make([]string, len(conditions))
		copy(optimizedConditions, conditions)

		// Replace trial-only status with active OR trial
		optimizedConditions[trialIndex] = `(eq(chorki_subscriptions.status, "trial") OR eq(chorki_subscriptions.status, "active"))`

		// Log this optimization (in production, you might want to use a logger)
		// fmt.Printf("OPTIMIZATION: Subscription filter expanded from trial-only to trial OR active for better results\n")

		return optimizedConditions
	}

	return conditions
}

// buildFilterCondition builds a single filter condition
func (c *Converter) buildFilterCondition(filter models.Filter, entityType string) string {
	// Find field mapping for this entity type
	mappings, exists := c.schema.FieldMappings[filter.Field]
	if !exists {
		return ""
	}

	var relevantMapping *models.FieldMapping
	for _, mapping := range mappings {
		if mapping.EntityType == entityType {
			relevantMapping = &mapping
			break
		}
	}

	if relevantMapping == nil {
		return ""
	}

	return c.buildDQLCondition(relevantMapping, filter)
}

// buildDQLCondition builds the actual DQL condition string
func (c *Converter) buildDQLCondition(mapping *models.FieldMapping, filter models.Filter) string {
	dqlFunction := c.operators[filter.Op]
	if dqlFunction == "" {
		return ""
	}

	switch filter.Op {
	case "IN":
		return c.buildInCondition(mapping, filter)
	case "NOT_IN":
		return c.buildNotInCondition(mapping, filter)
	case "=", ">=", "<=", ">", "<", "!=":
		return c.buildComparisonCondition(mapping, filter, dqlFunction)
	case "LIKE", "ILIKE", "CONTAINS":
		return c.buildTextSearchCondition(mapping, filter, filter.Op)
	case "REGEX":
		return c.buildRegexCondition(mapping, filter)
	case "BETWEEN":
		return c.buildBetweenCondition(mapping, filter)
	case "IS_NULL":
		return c.buildNullCondition(mapping, filter, true)
	case "IS_NOT_NULL":
		return c.buildNullCondition(mapping, filter, false)
	case "STARTS_WITH":
		return c.buildStringPatternCondition(mapping, filter, "starts_with")
	case "ENDS_WITH":
		return c.buildStringPatternCondition(mapping, filter, "ends_with")
	default:
		return ""
	}
}

// buildInCondition builds IN condition (handles arrays and complex objects)
func (c *Converter) buildInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		// Handle simple array values
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}
		if len(conditions) > 1 {
			return "(" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return conditions[0]
		}

	case map[string]interface{}:
		// Handle complex objects like watched_content
		return c.buildComplexObjectCondition(mapping, v)

	default:
		// Handle single value as if it's an array with one element
		value := c.formatValue(v, mapping.DataType)
		if value != "" {
			return fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value)
		}
	}

	return ""
}

// buildComplexObjectCondition handles complex object conditions like watched_content
func (c *Converter) buildComplexObjectCondition(mapping *models.FieldMapping, obj map[string]interface{}) string {
	if mapping.JSONField == "watched_content" {
		// Extract content_type and ids from the complex object
		if contentType, exists := obj["content_type"]; exists {
			if ids, idsExist := obj["ids"]; idsExist {
				if idArray, ok := ids.([]interface{}); ok {
					var conditions []string

					// Add content type condition if mapping exists
					if ctMappings, ctExists := c.schema.FieldMappings["content_type"]; ctExists {
						for _, ctMapping := range ctMappings {
							if ctMapping.EntityType == mapping.EntityType {
								typeValue := c.formatValue(contentType, "string")
								conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", ctMapping.DgraphField, typeValue))
								break
							}
						}
					}

					// Add ID conditions
					var idConditions []string
					for _, id := range idArray {
						idValue := c.formatValue(id, "int")
						if idValue != "" {
							idConditions = append(idConditions, idValue)
						}
					}

					if len(idConditions) > 0 {
						conditions = append(conditions, fmt.Sprintf("uid_in(%s, %s)", mapping.DgraphField, strings.Join(idConditions, ", ")))
					}

					if len(conditions) > 0 {
						return "(" + strings.Join(conditions, " AND ") + ")"
					}
				}
			}
		}
	}

	return ""
}

// buildComparisonCondition builds comparison conditions (=, >, <, etc.)
func (c *Converter) buildComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Check if this is a version field that needs special handling
	if mode, isVersionField := c.versionFields[filter.Field]; isVersionField && mode == "numeric" {
		return c.buildVersionComparisonCondition(mapping, filter, dqlFunction)
	}

	value := c.formatValue(filter.Value, mapping.DataType)
	if value == "" {
		return ""
	}

	return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
}

// buildVersionComparisonCondition handles version field comparisons using numeric conversion
func (c *Converter) buildVersionComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Convert version string to numeric value
	versionStr, ok := filter.Value.(string)
	if !ok {
		// Fallback to regular comparison if not a string
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	numericVersion, err := utils.ConvertVersionToNumeric(versionStr)
	if err != nil {
		// If conversion fails, fallback to string comparison with warning
		// In production, you might want to log this warning
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	// Use numeric comparison - note: we need a numeric version of the field in schema
	// For now, assume we have a parallel numeric field like app_version_numeric
	numericField := mapping.DgraphField + "_numeric"
	return fmt.Sprintf("%s(%s, %d)", dqlFunction, numericField, numericVersion)
}

// formatValue formats a value according to its data type for DQL
func (c *Converter) formatValue(value interface{}, dataType string) string {
	if value == nil {
		return ""
	}

	switch dataType {
	case "string":
		if str, ok := value.(string); ok {
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(str, `"`, `\"`))
		}
		return fmt.Sprintf(`"%v"`, value)

	case "int":
		switch v := value.(type) {
		case int:
			return strconv.Itoa(v)
		case float64:
			return strconv.Itoa(int(v))
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return strconv.Itoa(i)
			}
		}
		return fmt.Sprintf("%v", value)

	case "float":
		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.FormatFloat(float64(v), 'f', -1, 64)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return strconv.FormatFloat(f, 'f', -1, 64)
			}
		}
		return fmt.Sprintf("%v", value)

	case "bool":
		if b, ok := value.(bool); ok {
			return strconv.FormatBool(b)
		}
		return "false"

	default:
		return fmt.Sprintf(`"%v"`, value)
	}
}

// getQueryName generates a query name based on entity type
func (c *Converter) getQueryName(entityType string) string {
	switch entityType {
	case "chorki_customers":
		return "customers"
	case "chorki_subscriptions":
		return "subscriptions"
	case "chorki_watch_histories":
		return "watch_histories"
	case "chorki_contents":
		return "contents"
	case "chorki_devices":
		return "devices"
	default:
		return strings.ReplaceAll(entityType, "chorki_", "")
	}
}

// buildFieldsSelection builds the fields selection for an entity type
func (c *Converter) buildFieldsSelection(entityType string) string {
	fields := c.schema.DefaultFields[entityType]
	if len(fields) == 0 {
		return "uid\nexpand(_all_)"
	}

	var fieldLines []string
	for _, field := range fields {
		fieldLines = append(fieldLines, "    "+field)
	}

	// Add related entities if they exist
	if relationships, exists := c.schema.Relationships[entityType]; exists {
		for _, relatedEntity := range relationships {
			relatedFields := c.schema.DefaultFields[relatedEntity]
			if len(relatedFields) > 0 {
				relationName := c.getRelationshipName(entityType, relatedEntity)
				fieldLines = append(fieldLines, "")
				fieldLines = append(fieldLines, fmt.Sprintf("    %s {", relationName))
				for _, relField := range relatedFields {
					fieldLines = append(fieldLines, "      "+relField)
				}
				fieldLines = append(fieldLines, "    }")
			}
		}
	}

	return strings.Join(fieldLines, "\n")
}

// getRelationshipName returns the relationship predicate name between two entities
func (c *Converter) getRelationshipName(fromEntity, toEntity string) string {
	switch {
	// Forward relationships from customers
	case fromEntity == "chorki_customers" && toEntity == "chorki_subscriptions":
		return "chorki_customers.subscriptions"
	case fromEntity == "chorki_customers" && toEntity == "chorki_watch_histories":
		return "chorki_customers.watch_histories"
	case fromEntity == "chorki_customers" && toEntity == "chorki_devices":
		return "chorki_customers.devices"

	// Reverse relationships to customers
	case fromEntity == "chorki_subscriptions" && toEntity == "chorki_customers":
		return "~chorki_customers.subscriptions" // reverse edge
	case fromEntity == "chorki_devices" && toEntity == "chorki_customers":
		return "~chorki_customers.devices" // reverse edge
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_customers":
		return "~chorki_customers.watch_histories" // reverse edge

	// Content relationships
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_contents":
		return "chorki_watch_histories.content"
	case fromEntity == "chorki_contents" && toEntity == "chorki_watch_histories":
		return "~chorki_watch_histories.content" // reverse edge

	// Default fallback - use simple name for relationships
	default:
		// For reverse relationships, check if it should be a reverse predicate
		if toEntity == "chorki_customers" {
			return "customers" // Simple name for reverse edge
		}
		return strings.ReplaceAll(toEntity, "chorki_", "")
	}
}

// GenerateDQLString generates the final DQL query string
func (c *Converter) GenerateDQLString(dqlQuery *models.DQLQuery) string {
	if len(dqlQuery.Queries) == 0 {
		return "{}"
	}

	var queryBlocks []string

	for _, query := range dqlQuery.Queries {
		block := fmt.Sprintf("  %s(func: %s) %s {\n%s\n  }",
			query.Name,
			query.Function,
			query.Filter,
			query.Fields,
		)
		queryBlocks = append(queryBlocks, block)
	}

	return "{\n" + strings.Join(queryBlocks, "\n\n") + "\n}"
}

// AnalyzeComplexity returns the complexity analysis for a given query
func (c *Converter) AnalyzeComplexity(jsonQuery *models.JSONQuery) *analyzer.ComplexityScore {
	return c.complexityAnalyzer.AnalyzeComplexity(jsonQuery)
}

// GetComplexityLimits returns the current complexity limits
func (c *Converter) GetComplexityLimits() map[string]int {
	return c.complexityAnalyzer.GetComplexityLimits()
}

// buildNotInCondition builds NOT IN condition
func (c *Converter) buildNotInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}
		if len(conditions) > 1 {
			return "NOT (" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return "NOT " + conditions[0]
		}
	default:
		value := c.formatValue(v, mapping.DataType)
		if value != "" {
			return fmt.Sprintf("NOT eq(%s, %s)", mapping.DgraphField, value)
		}
	}
	return ""
}

// buildTextSearchCondition builds text search conditions
func (c *Converter) buildTextSearchCondition(mapping *models.FieldMapping, filter models.Filter, op string) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}

	switch op {
	case "LIKE", "CONTAINS":
		return fmt.Sprintf("alloftext(%s, %s)", mapping.DgraphField, value)
	case "ILIKE":
		return fmt.Sprintf("anyoftext(%s, %s)", mapping.DgraphField, value)
	default:
		return ""
	}
}

// buildRegexCondition builds regex conditions
func (c *Converter) buildRegexCondition(mapping *models.FieldMapping, filter models.Filter) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}
	return fmt.Sprintf("regexp(%s, %s)", mapping.DgraphField, value)
}

// buildBetweenCondition builds BETWEEN conditions
func (c *Converter) buildBetweenCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		if len(v) == 2 {
			min := c.formatValue(v[0], mapping.DataType)
			max := c.formatValue(v[1], mapping.DataType)
			if min != "" && max != "" {
				return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
					mapping.DgraphField, min, mapping.DgraphField, max)
			}
		}
	case map[string]interface{}:
		if minVal, hasMin := v["min"]; hasMin {
			if maxVal, hasMax := v["max"]; hasMax {
				min := c.formatValue(minVal, mapping.DataType)
				max := c.formatValue(maxVal, mapping.DataType)
				if min != "" && max != "" {
					return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
						mapping.DgraphField, min, mapping.DgraphField, max)
				}
			}
		}
	}
	return ""
}

// buildNullCondition builds NULL/NOT NULL conditions
func (c *Converter) buildNullCondition(mapping *models.FieldMapping, filter models.Filter, isNull bool) string {
	if isNull {
		return fmt.Sprintf("NOT has(%s)", mapping.DgraphField)
	} else {
		return fmt.Sprintf("has(%s)", mapping.DgraphField)
	}
}

// buildStringPatternCondition builds string pattern conditions
func (c *Converter) buildStringPatternCondition(mapping *models.FieldMapping, filter models.Filter, pattern string) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}

	// Remove quotes for pattern matching
	cleanValue := strings.Trim(value, `"`)

	switch pattern {
	case "starts_with":
		return fmt.Sprintf("regexp(%s, /^%s/)", mapping.DgraphField, cleanValue)
	case "ends_with":
		return fmt.Sprintf("regexp(%s, /%s$/)", mapping.DgraphField, cleanValue)
	default:
		return ""
	}
}
