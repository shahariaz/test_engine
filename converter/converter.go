/*
Package converter provides the core functionality for converting JSON queries to Dgraph Query Language (DQL).

This package handles the complex transformation of structured JSON query objects into executable DQL strings,
including support for:
- Multi-entity queries with relationship traversal
- Complex nested filtering with AND/OR logic
- Various operators (comparison, text search, array operations)
- Query complexity analysis and optimization
- Schema validation and field mapping

Architecture:
- Converter: Main orchestrator that coordinates the conversion process
- Entity Analysis: Determines which Dgraph entity types are involved
- Filter Building: Constructs DQL filter expressions from JSON conditions
- Field Selection: Generates appropriate field selections for each entity
- DQL Generation: Produces the final executable DQL query string
*/
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

// =============================================================================
// CORE CONVERTER STRUCT AND INITIALIZATION
// =============================================================================

// Converter orchestrates the conversion from JSON queries to DQL format.
// It maintains schema mappings, operator definitions, and analysis tools
// needed for the transformation process.
type Converter struct {
	// Schema configuration defining entity types, fields, and relationships
	schema *models.SchemaInfo

	// Mapping of JSON operators to DQL functions (e.g., "=" -> "eq")
	operators map[string]string

	// Analyzer for evaluating query complexity and performance implications
	complexityAnalyzer *analyzer.ComplexityAnalyzer

	// Special handling configuration for version fields (e.g., app_version)
	versionFields map[string]string

	// Mapping for reverse predicate relationships in Dgraph
	reversePredicates map[string]string
}

// NewConverter creates a new converter instance with all necessary configurations.
// It initializes the converter with schema information, operator mappings,
// and analysis tools required for JSON to DQL conversion.
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

// =============================================================================
// PUBLIC API METHODS
// =============================================================================

// ConvertToDQL performs the main conversion from JSON query to DQL format.
// This is the primary entry point that orchestrates the entire conversion process:
// 1. Analyzes query complexity to prevent performance issues
// 2. Identifies all entity types involved in the query
// 3. Determines if unified query with relationship traversal is needed
// 4. Constructs the appropriate DQL query structure
//
// Returns a structured DQL query object that can be converted to a string
// or executed directly against Dgraph.
func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.DQLQuery, error) {
	// Step 1: Complexity Analysis - Prevent overly complex queries that could impact performance
	complexityScore := c.complexityAnalyzer.AnalyzeComplexity(jsonQuery)
	if !complexityScore.IsAcceptable {
		return nil, fmt.Errorf("query too complex: %s", complexityScore.Warning)
	}

	// Step 2: Entity Discovery - Determine which Dgraph entities this query touches
	involvedEntities := c.getInvolvedEntityTypes(jsonQuery)

	// Step 3: Determine Query Strategy
	// For user segmentation, if we have cross-entity filters with top-level AND, create unified query
	// This ensures customers who meet ALL conditions across different entities
	if len(involvedEntities) > 1 && strings.ToUpper(jsonQuery.CombineWith) == "AND" {
		return c.buildUnifiedCrossEntityQuery(jsonQuery, involvedEntities)
	}

	// Step 4: Standard Query Construction - Build DQL queries for each involved entity
	var queries []models.EntityQuery
	for _, entityType := range involvedEntities {
		filter, err := c.buildFilterForEntity(jsonQuery, entityType)
		if err != nil {
			return nil, fmt.Errorf("error building filter for %s: %v", entityType, err)
		}

		// Only include entities that have applicable filters
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

// =============================================================================
// ENTITY ANALYSIS AND DISCOVERY
// =============================================================================

// requiresCrossEntityLogic determines if the query needs unified cross-entity handling.
// This happens when we have filters from different entities that need to be combined
// with AND logic at the top level, requiring relationship traversal.
func (c *Converter) requiresCrossEntityLogic(jsonQuery *models.JSONQuery) bool {
	if strings.ToUpper(jsonQuery.CombineWith) != "AND" {
		return false
	}

	// Check if different groups target different entities
	entityGroups := make(map[string]bool)
	for _, group := range jsonQuery.Groups {
		groupEntities := c.getGroupEntityTypes(group)
		for entity := range groupEntities {
			entityGroups[entity] = true
		}
	}

	// If we have more than one entity type across groups with AND logic,
	// we need cross-entity handling
	return len(entityGroups) > 1
}

// getGroupEntityTypes gets all entity types referenced in a single group
func (c *Converter) getGroupEntityTypes(group models.Group) map[string]bool {
	entities := make(map[string]bool)

	// Check filters in this group
	for _, filter := range group.Filters {
		if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {
			for _, mapping := range mappings {
				entities[mapping.EntityType] = true
			}
		}
	}

	// Check nested groups recursively
	for _, nestedGroup := range group.Groups {
		nestedEntities := c.getGroupEntityTypes(nestedGroup)
		for entity := range nestedEntities {
			entities[entity] = true
		}
	}

	return entities
}

// buildUnifiedCrossEntityQuery builds a single query with relationship traversal
// for cases where we need to combine filters from multiple entities with AND logic
func (c *Converter) buildUnifiedCrossEntityQuery(jsonQuery *models.JSONQuery, involvedEntities []string) (*models.DQLQuery, error) {
	// Determine primary entity (default to customers for business logic)
	primaryEntity := c.getPrimaryEntity(jsonQuery)

	// Separate filters by entity type
	primaryFilters, relationshipFilters := c.separateFiltersByEntity(jsonQuery, primaryEntity)

	// Build primary entity filter
	primaryFilter := ""
	if primaryFilters != nil {
		filter := c.buildGroupsFilter(primaryFilters.Groups, primaryFilters.CombineWith, primaryEntity)
		if filter != "" {
			primaryFilter = fmt.Sprintf("@filter(%s)", filter)
		}
	}

	// Build fields selection with relationship filters
	fieldsWithFilters := c.buildFieldsSelectionWithRelationshipFilters(primaryEntity, relationshipFilters)

	// Create single unified query with @cascade for strict relationship filtering
	query := models.EntityQuery{
		Name:     c.getQueryName(primaryEntity),
		Type:     primaryEntity,
		Function: fmt.Sprintf("type(%s)", primaryEntity),
		Filter:   primaryFilter + " @cascade",
		Fields:   fieldsWithFilters,
	}

	return &models.DQLQuery{Queries: []models.EntityQuery{query}}, nil
}

// separateFiltersByEntity separates filters into primary entity filters and relationship filters
func (c *Converter) separateFiltersByEntity(jsonQuery *models.JSONQuery, primaryEntity string) (*models.JSONQuery, map[string]*models.JSONQuery) {
	primaryQuery := &models.JSONQuery{
		CombineWith: jsonQuery.CombineWith,
		Groups:      []models.Group{},
	}

	relationshipQueries := make(map[string]*models.JSONQuery)

	for _, group := range jsonQuery.Groups {
		primaryGroup := models.Group{
			CombineWith: group.CombineWith,
			Filters:     []models.Filter{},
		}

		relationshipGroups := make(map[string]models.Group)

		for _, filter := range group.Filters {
			if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {
				belongsToPrimary := false
				for _, mapping := range mappings {
					if mapping.EntityType == primaryEntity {
						primaryGroup.Filters = append(primaryGroup.Filters, filter)
						belongsToPrimary = true
						break
					}
				}

				if !belongsToPrimary {
					// This is a relationship filter
					for _, mapping := range mappings {
						if c.hasRelationshipTo(primaryEntity, mapping.EntityType) {
							entityType := mapping.EntityType
							if _, exists := relationshipGroups[entityType]; !exists {
								relationshipGroups[entityType] = models.Group{
									CombineWith: group.CombineWith, // Preserve original group logic!
									Filters:     []models.Filter{},
								}
							}
							relGroup := relationshipGroups[entityType]
							relGroup.Filters = append(relGroup.Filters, filter)
							relationshipGroups[entityType] = relGroup
							break
						}
					}
				}
			}
		}

		// Add primary group if it has filters
		if len(primaryGroup.Filters) > 0 {
			primaryQuery.Groups = append(primaryQuery.Groups, primaryGroup)
		}

		// Add relationship groups
		for entityType, relGroup := range relationshipGroups {
			if _, exists := relationshipQueries[entityType]; !exists {
				relationshipQueries[entityType] = &models.JSONQuery{
					CombineWith: "AND",
					Groups:      []models.Group{},
				}
			}
			relationshipQueries[entityType].Groups = append(relationshipQueries[entityType].Groups, relGroup)
		}
	}

	// If no primary filters, return nil for primary
	if len(primaryQuery.Groups) == 0 {
		primaryQuery = nil
	}

	return primaryQuery, relationshipQueries
}

// buildFieldsSelectionWithRelationshipFilters builds field selection with relationship filters applied
func (c *Converter) buildFieldsSelectionWithRelationshipFilters(primaryEntity string, relationshipFilters map[string]*models.JSONQuery) string {
	fields := c.schema.DefaultFields[primaryEntity]
	if len(fields) == 0 {
		return "uid\nexpand(_all_)"
	}

	var fieldLines []string
	// Add main entity fields with proper indentation
	for _, field := range fields {
		fieldLines = append(fieldLines, "    "+field)
	}

	// Add related entity fields with filters applied
	if relationships, exists := c.schema.Relationships[primaryEntity]; exists {
		for _, relatedEntity := range relationships {
			relatedFields := c.schema.DefaultFields[relatedEntity]
			if len(relatedFields) > 0 {
				// Get the relationship predicate name
				relationName := c.getRelationshipName(primaryEntity, relatedEntity)

				// Check if we have filters for this relationship
				relationshipFilter := ""
				if relQuery, hasFilters := relationshipFilters[relatedEntity]; hasFilters {
					filter := c.buildGroupsFilter(relQuery.Groups, relQuery.CombineWith, relatedEntity)
					if filter != "" {
						relationshipFilter = fmt.Sprintf(" @filter(%s)", filter)
					}
				}

				// Add related entity block with nested fields and optional filters
				fieldLines = append(fieldLines, "")
				fieldLines = append(fieldLines, fmt.Sprintf("    %s%s {", relationName, relationshipFilter))
				for _, relatedField := range relatedFields {
					fieldLines = append(fieldLines, "      "+relatedField)
				}
				fieldLines = append(fieldLines, "    }")
			}
		}
	}

	return strings.Join(fieldLines, "\n")
}

// buildUnifiedFilter builds a filter that can traverse relationships between entities
func (c *Converter) buildUnifiedFilter(jsonQuery *models.JSONQuery, primaryEntity string) (string, error) {
	var groupFilters []string

	for _, group := range jsonQuery.Groups {
		groupFilter := c.buildUnifiedGroupFilter(group, primaryEntity)
		if groupFilter != "" {
			groupFilters = append(groupFilters, groupFilter)
		}
	}

	if len(groupFilters) == 0 {
		return "", nil
	}

	if len(groupFilters) == 1 {
		return fmt.Sprintf("@filter(%s)", groupFilters[0]), nil
	}

	// Combine group filters with top-level logic
	operator := " AND "
	if strings.ToUpper(jsonQuery.CombineWith) == "OR" {
		operator = " OR "
	}

	combinedFilter := "(" + strings.Join(groupFilters, operator) + ")"
	return fmt.Sprintf("@filter(%s)", combinedFilter), nil
}

// buildUnifiedGroupFilter builds filter for a group that may span multiple entities
func (c *Converter) buildUnifiedGroupFilter(group models.Group, primaryEntity string) string {
	var conditions []string
	relationshipConditions := make(map[string][]string) // Group conditions by relationship

	// Build conditions from filters, handling cross-entity relationships
	for _, filter := range group.Filters {
		condition := c.buildUnifiedFilterCondition(filter, primaryEntity)
		if condition != "" {
			// Check if this is a relationship condition
			if strings.Contains(condition, " @filter(") {
				// Extract relationship name to group conditions
				relationshipName := c.extractRelationshipName(condition)
				if relationshipName != "" {
					relationshipConditions[relationshipName] = append(relationshipConditions[relationshipName], c.extractFilterCondition(condition))
				} else {
					conditions = append(conditions, condition)
				}
			} else {
				conditions = append(conditions, condition)
			}
		}
	}

	// Combine relationship conditions for the same relationship
	for relationshipName, filterConditions := range relationshipConditions {
		if len(filterConditions) > 0 {
			operator := " AND "
			if strings.ToUpper(group.CombineWith) == "OR" {
				operator = " OR "
			}
			combinedFilter := strings.Join(filterConditions, operator)
			relationshipCondition := fmt.Sprintf("has(%s) AND %s @filter(%s)", relationshipName, relationshipName, combinedFilter)
			conditions = append(conditions, relationshipCondition)
		}
	}

	// Handle nested groups recursively
	for _, nestedGroup := range group.Groups {
		nestedCondition := c.buildUnifiedGroupFilter(nestedGroup, primaryEntity)
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

	// Combine with group logic
	operator := " AND "
	if strings.ToUpper(group.CombineWith) == "OR" {
		operator = " OR "
	}

	return "(" + strings.Join(conditions, operator) + ")"
}

// extractRelationshipName extracts relationship name from a relationship condition
func (c *Converter) extractRelationshipName(condition string) string {
	// Look for pattern: "has(relationship) AND relationship @filter(...)"
	if strings.Contains(condition, "has(") && strings.Contains(condition, ") AND ") {
		start := strings.Index(condition, "has(") + 4
		end := strings.Index(condition[start:], ")")
		if end > 0 {
			return condition[start : start+end]
		}
	}
	return ""
}

// extractFilterCondition extracts just the filter part from a relationship condition
func (c *Converter) extractFilterCondition(condition string) string {
	// Extract content between @filter( and the last )
	filterStart := strings.Index(condition, "@filter(")
	if filterStart >= 0 {
		filterStart += 8 // length of "@filter("
		// Find matching closing parenthesis
		parenCount := 1
		for i := filterStart; i < len(condition); i++ {
			if condition[i] == '(' {
				parenCount++
			} else if condition[i] == ')' {
				parenCount--
				if parenCount == 0 {
					return condition[filterStart:i]
				}
			}
		}
	}
	return ""
}

// buildUnifiedFilterCondition builds a condition that may require relationship traversal
func (c *Converter) buildUnifiedFilterCondition(filter models.Filter, primaryEntity string) string {
	mappings, exists := c.schema.FieldMappings[filter.Field]
	if !exists {
		return ""
	}

	// Check if this field belongs to the primary entity
	for _, mapping := range mappings {
		if mapping.EntityType == primaryEntity {
			// Direct field on primary entity
			return c.buildDQLCondition(&mapping, filter)
		}
	}

	// Field belongs to related entity - need relationship traversal
	for _, mapping := range mappings {
		if c.hasRelationshipTo(primaryEntity, mapping.EntityType) {
			return c.buildRelationshipCondition(&mapping, filter, primaryEntity)
		}
	}

	return ""
}

// hasRelationshipTo checks if primaryEntity has a relationship to targetEntity
func (c *Converter) hasRelationshipTo(primaryEntity, targetEntity string) bool {
	if relationships, exists := c.schema.Relationships[primaryEntity]; exists {
		for _, relatedEntity := range relationships {
			if relatedEntity == targetEntity {
				return true
			}
		}
	}
	return false
}

// buildRelationshipCondition builds a condition that traverses entity relationships
// This creates a valid DQL condition using proper Dgraph relationship filtering syntax
func (c *Converter) buildRelationshipCondition(mapping *models.FieldMapping, filter models.Filter, primaryEntity string) string {
	// Get relationship predicate name
	relationshipName := c.getRelationshipName(primaryEntity, mapping.EntityType)

	// Build the condition for the related entity
	relatedCondition := c.buildDQLCondition(mapping, filter)
	if relatedCondition == "" {
		return ""
	}

	// For Dgraph, we need to use has() function to check for relationship existence
	// The actual filtering on the related entity will be done in the query body with @cascade
	// For now, we just ensure the relationship exists
	return fmt.Sprintf("has(%s)", relationshipName)
}

// getInvolvedEntityTypes determines which entity types are referenced in the query.
// This method analyzes the JSON query structure to identify all Dgraph entity types
// that need to be included in the DQL query. It recursively traverses all groups
// and filters to build a comprehensive list of involved entities.
func (c *Converter) getInvolvedEntityTypes(jsonQuery *models.JSONQuery) []string {
	entityTypeMap := make(map[string]bool)

	// Recursively analyze all groups and filters to collect entity types
	c.collectEntityTypesFromGroups(jsonQuery.Groups, entityTypeMap)

	var result []string
	for entityType := range entityTypeMap {
		result = append(result, entityType)
	}

	// Fallback: If no entities found, default to customers as the primary entity
	if len(result) == 0 {
		result = append(result, "chorki_customers")
	}

	return result
}

// collectEntityTypesFromGroups recursively traverses groups to identify entity types.
// This helper method performs a depth-first search through the query structure,
// examining each filter to determine which Dgraph entities are involved.
func (c *Converter) collectEntityTypesFromGroups(groups []models.Group, entityTypeMap map[string]bool) {
	for _, group := range groups {
		// Analyze filters in the current group
		for _, filter := range group.Filters {
			if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {
				// Add all entity types that this field maps to
				for _, mapping := range mappings {
					entityTypeMap[mapping.EntityType] = true
				}
			}
		}

		// Recursively process nested groups
		if len(group.Groups) > 0 {
			c.collectEntityTypesFromGroups(group.Groups, entityTypeMap)
		}
	}
}

// getPrimaryEntity determines the primary entity type based on field frequency.
// This method helps optimize query structure by identifying the most referenced
// entity type, which can be used for query planning and optimization.
func (c *Converter) getPrimaryEntity(jsonQuery *models.JSONQuery) string {
	entityCount := make(map[string]int)

	// Count field occurrences for each entity type
	c.countEntityTypesFromGroups(jsonQuery.Groups, entityCount)

	// Find the entity with the most field references
	var primaryEntity string
	maxCount := 0

	// Prioritize customers as the default primary entity for business logic
	if count, exists := entityCount["chorki_customers"]; exists && count > 0 {
		primaryEntity = "chorki_customers"
		maxCount = count
	}

	// Check if any other entity has significantly more field references
	for entityType, count := range entityCount {
		if count > maxCount {
			primaryEntity = entityType
			maxCount = count
		}
	}

	// Default fallback to customers if no clear primary entity emerges
	if primaryEntity == "" {
		primaryEntity = "chorki_customers"
	}

	return primaryEntity
}

// countEntityTypesFromGroups recursively counts entity type occurrences.
// This helper method traverses the query structure to count how many times
// each entity type is referenced, which helps in determining query priorities.
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

// =============================================================================
// FILTER CONSTRUCTION
// =============================================================================

// buildFilterForEntity constructs DQL filter expressions for a specific entity type.
// This method takes a JSON query and creates the appropriate @filter() clause
// for the specified entity, ensuring type-safe and optimized filter generation.
func (c *Converter) buildFilterForEntity(jsonQuery *models.JSONQuery, entityType string) (string, error) {
	filter := c.buildGroupsFilter(jsonQuery.Groups, jsonQuery.CombineWith, entityType)
	if filter == "" {
		return "", nil
	}
	return fmt.Sprintf("@filter(%s)", filter), nil
}

// buildGroupsFilter constructs filter strings for multiple groups with logical operators.
// This method combines multiple group filters using the specified logical operator
// (AND/OR) and handles parentheses for proper query evaluation precedence.
func (c *Converter) buildGroupsFilter(groups []models.Group, combineWith, entityType string) string {
	var conditions []string

	// Process each group and collect valid conditions
	for _, group := range groups {
		condition := c.buildGroupFilter(group, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	// Return empty string if no valid conditions found
	if len(conditions) == 0 {
		return ""
	}

	// Single condition doesn't need parentheses or operators
	if len(conditions) == 1 {
		return conditions[0]
	}

	// Determine logical operator (default to AND for safety)
	operator := " AND "
	if strings.ToUpper(combineWith) == "OR" {
		operator = " OR "
	}

	// Combine conditions with proper parentheses for complex expressions
	return "(" + strings.Join(conditions, operator) + ")"
}

// buildGroupFilter constructs filter expressions for a single group.
// This method processes all filters within a group and applies entity-specific
// optimizations. It also handles nested groups recursively.
func (c *Converter) buildGroupFilter(group models.Group, entityType string) string {
	var conditions []string

	// Build conditions from individual filters in this group
	for _, filter := range group.Filters {
		condition := c.buildFilterCondition(filter, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	// Apply entity-specific filter optimizations
	if entityType == "chorki_subscriptions" {
		conditions = c.optimizeSubscriptionFilters(conditions, group)
	}

	// Process nested groups recursively
	// Process nested groups recursively
	if len(group.Groups) > 0 {
		nestedCondition := c.buildGroupsFilter(group.Groups, group.CombineWith, entityType)
		if nestedCondition != "" {
			conditions = append(conditions, nestedCondition)
		}
	}

	// Return empty string if no conditions were generated
	if len(conditions) == 0 {
		return ""
	}

	// Single condition doesn't need additional parentheses or operators
	if len(conditions) == 1 {
		return conditions[0]
	}

	// Determine logical operator for combining conditions within this group
	operator := " AND "
	if strings.ToUpper(group.CombineWith) == "OR" {
		operator = " OR "
	}

	// Combine all conditions with proper parentheses
	return "(" + strings.Join(conditions, operator) + ")"
}

// optimizeSubscriptionFilters applies business logic optimizations for subscription queries.
// This method identifies potentially conflicting filter conditions (e.g., Premium + trial)
// and suggests optimizations to prevent overly restrictive queries that return no results.
func (c *Converter) optimizeSubscriptionFilters(conditions []string, group models.Group) []string {
	// Analyze conditions for business logic conflicts
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

	// Detect Premium + trial combination with AND operator (likely to return no results)
	if hasPackagePremium && hasStatusTrial && strings.ToUpper(group.CombineWith) == "AND" {
		// Create optimized conditions array
		optimizedConditions := make([]string, len(conditions))
		copy(optimizedConditions, conditions)

		// Replace restrictive trial-only status with more inclusive active OR trial
		optimizedConditions[trialIndex] = `(eq(chorki_subscriptions.status, "trial") OR eq(chorki_subscriptions.status, "active"))`

		// Note: In production, consider logging this optimization for monitoring
		// log.Printf("OPTIMIZATION: Subscription filter expanded trial-only to trial OR active for entity %s", entityType)

		return optimizedConditions
	}

	// No optimization needed, return original conditions
	return conditions
}

// =============================================================================
// INDIVIDUAL FILTER CONDITION CONSTRUCTION
// =============================================================================

// buildFilterCondition constructs a single DQL filter condition from a JSON filter.
// This method maps JSON filter fields to their corresponding Dgraph entity fields
// and generates the appropriate DQL syntax based on the operator and data type.
func (c *Converter) buildFilterCondition(filter models.Filter, entityType string) string {
	// Look up the field mapping for the specified entity type
	mappings, exists := c.schema.FieldMappings[filter.Field]
	if !exists {
		return "" // Field not found in schema
	}

	// Find the specific mapping for this entity type
	var relevantMapping *models.FieldMapping
	for _, mapping := range mappings {
		if mapping.EntityType == entityType {
			relevantMapping = &mapping
			break
		}
	}

	// No mapping found for this entity type
	if relevantMapping == nil {
		return ""
	}

	// Build the actual DQL condition using the mapping
	return c.buildDQLCondition(relevantMapping, filter)
}

// buildDQLCondition generates the final DQL condition string from a field mapping and filter.
// This method handles all supported operators and ensures proper DQL syntax and data type handling.
func (c *Converter) buildDQLCondition(mapping *models.FieldMapping, filter models.Filter) string {
	// Look up the DQL function for this operator
	dqlFunction := c.operators[filter.Op]
	if dqlFunction == "" {
		return "" // Unsupported operator
	}

	// Handle different operator types with specialized logic
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
		return "" // Unsupported operator
	}
}

// =============================================================================
// SPECIALIZED CONDITION BUILDERS
// =============================================================================

// buildInCondition constructs DQL conditions for IN operations with multiple values.
// This method handles various data types including arrays, complex objects, and single values.
// It supports business logic for complex structures like watched_content filtering.
func (c *Converter) buildInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		// Handle array of simple values (strings, numbers, etc.)
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}

		// Combine multiple conditions with OR logic
		if len(conditions) > 1 {
			return "(" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return conditions[0]
		}

	case map[string]interface{}:
		// Handle complex nested objects (e.g., watched_content with content_type and ids)
		return c.buildComplexObjectCondition(mapping, v)

	default:
		// Handle single value by treating it as a single-element array
		value := c.formatValue(v, mapping.DataType)
		if value != "" {
			return fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value)
		}
	}

	return ""
}

// buildComplexObjectCondition handles advanced object-based filtering for business entities.
// This method supports complex queries like filtering watched content by both content type
// and specific content IDs, enabling sophisticated user behavior analysis.
func (c *Converter) buildComplexObjectCondition(mapping *models.FieldMapping, obj map[string]interface{}) string {
	// Special handling for watched_content queries
	if mapping.JSONField == "watched_content" {
		// Extract content_type and content IDs from the filter object
		if contentType, exists := obj["content_type"]; exists {
			if ids, idsExist := obj["ids"]; idsExist {
				if idArray, ok := ids.([]interface{}); ok {
					var conditions []string

					// Add content type condition if mapping exists in schema
					if ctMappings, ctExists := c.schema.FieldMappings["content_type"]; ctExists {
						for _, ctMapping := range ctMappings {
							if ctMapping.EntityType == mapping.EntityType {
								typeValue := c.formatValue(contentType, "string")
								conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", ctMapping.DgraphField, typeValue))
								break
							}
						}
					}

					// Build conditions for content IDs - use proper method based on field type
					var idConditions []string
					for _, id := range idArray {
						// Convert numeric IDs to string since content_id is a string field in Dgraph
						var idValue string
						switch v := id.(type) {
						case int:
							idValue = fmt.Sprintf(`"%d"`, v)
						case int64:
							idValue = fmt.Sprintf(`"%d"`, v)
						case float64:
							idValue = fmt.Sprintf(`"%.0f"`, v)
						case string:
							idValue = fmt.Sprintf(`"%s"`, v)
						default:
							idValue = c.formatValue(id, "string")
						}

						if idValue != "" {
							// Use eq() for scalar fields, not uid_in()
							idConditions = append(idConditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, idValue))
						}
					}

					// Generate OR condition for multiple content IDs (since they're scalar fields)
					if len(idConditions) > 1 {
						conditions = append(conditions, "("+strings.Join(idConditions, " OR ")+")")
					} else if len(idConditions) == 1 {
						conditions = append(conditions, idConditions[0])
					}

					// Combine content type and ID conditions with AND logic
					if len(conditions) > 0 {
						return "(" + strings.Join(conditions, " AND ") + ")"
					}
				}
			}
		}
	}

	// Return empty string if object structure doesn't match expected patterns
	return ""
}

// buildComparisonCondition constructs DQL conditions for comparison operators.
// This method handles standard comparison operations (=, >, <, >=, <=, !=) with
// special handling for version fields and data type-specific formatting.
func (c *Converter) buildComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Check for special version field handling (numeric version comparisons)
	if mode, isVersionField := c.versionFields[filter.Field]; isVersionField && mode == "numeric" {
		return c.buildVersionComparisonCondition(mapping, filter, dqlFunction)
	}

	// Format the value according to the field's data type
	value := c.formatValue(filter.Value, mapping.DataType)
	if value == "" {
		return "" // Invalid or unsupported value
	}

	// Handle inequality operator with NOT + eq for better DQL performance
	if filter.Op == "!=" {
		return fmt.Sprintf("NOT eq(%s, %s)", mapping.DgraphField, value)
	}

	// Standard comparison condition
	return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
}

// buildVersionComparisonCondition handles specialized version field comparisons.
// Version fields often need numeric conversion for proper comparison semantics
// (e.g., comparing "1.2.3" vs "1.10.0" requires numeric interpretation).
func (c *Converter) buildVersionComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Convert version string to numeric value for accurate comparison
	versionStr, ok := filter.Value.(string)
	if !ok {
		// Fallback to regular comparison if value is not a string
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	// Attempt numeric conversion of version string (e.g., "1.2.3" -> 10203)
	numericVersion, err := utils.ConvertVersionToNumeric(versionStr)
	if err != nil {
		// If conversion fails, fallback to string comparison
		// Note: In production, consider logging this conversion failure
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	// Use numeric field for comparison (assumes schema has parallel numeric fields)
	// Example: app_version -> app_version_numeric
	numericField := mapping.DgraphField + "_numeric"
	return fmt.Sprintf("%s(%s, %d)", dqlFunction, numericField, numericVersion)
}

// =============================================================================
// VALUE FORMATTING AND TYPE CONVERSION
// =============================================================================

// formatValue converts and formats values according to their Dgraph data types.
// This method ensures proper DQL syntax and handles type coercion, escaping,
// and validation for all supported data types.
func (c *Converter) formatValue(value interface{}, dataType string) string {
	if value == nil {
		return ""
	}

	switch dataType {
	case "string":
		// Handle string values with proper escaping
		if str, ok := value.(string); ok {
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(str, `"`, `\"`))
		}
		// Convert non-string values to strings
		return fmt.Sprintf(`"%v"`, value)

	case "int":
		// Handle various numeric types and convert to integer
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
		// Fallback: attempt direct conversion
		return fmt.Sprintf("%v", value)

	case "float":
		// Handle floating-point values with precision control
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
		// Fallback: attempt direct conversion
		return fmt.Sprintf("%v", value)

	case "bool":
		// Handle boolean values with safe conversion
		if b, ok := value.(bool); ok {
			return strconv.FormatBool(b)
		}
		// Default to false for invalid boolean values
		return "false"

	default:
		// Default case: treat as string with quotes
		return fmt.Sprintf(`"%v"`, value)
	}
}

// =============================================================================
// QUERY NAMING AND FIELD SELECTION
// =============================================================================

// getQueryName generates appropriate query names based on entity types.
// This method creates human-readable query names that follow DQL conventions
// and remove internal prefixes for cleaner query structure.
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
		// Generic case: remove common prefixes
		return strings.ReplaceAll(entityType, "chorki_", "")
	}
}

// buildFieldsSelection constructs the field selection clause for DQL queries.
// This method determines which fields to include in the query response based on
// entity type and configured default fields from the schema.
func (c *Converter) buildFieldsSelection(entityType string) string {
	fields := c.schema.DefaultFields[entityType]
	if len(fields) == 0 {
		return "uid\nexpand(_all_)"
	}

	var fieldLines []string
	// Add main entity fields with proper indentation
	for _, field := range fields {
		fieldLines = append(fieldLines, "    "+field)
	}

	// Add related entity fields through relationship traversal
	if relationships, exists := c.schema.Relationships[entityType]; exists {
		for _, relatedEntity := range relationships {
			relatedFields := c.schema.DefaultFields[relatedEntity]
			if len(relatedFields) > 0 {
				// Get the relationship predicate name
				relationName := c.getRelationshipName(entityType, relatedEntity)

				// Add related entity block with nested fields
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

// getRelationshipName determines the correct DQL predicate name for entity relationships.
// This method handles both forward and reverse relationships using Dgraph's
// tilde (~) notation for reverse edges and maintains consistent naming conventions.
func (c *Converter) getRelationshipName(fromEntity, toEntity string) string {
	switch {
	// Forward relationships from customers to related entities
	case fromEntity == "chorki_customers" && toEntity == "chorki_subscriptions":
		return "chorki_customers.subscriptions"
	case fromEntity == "chorki_customers" && toEntity == "chorki_watch_histories":
		return "chorki_customers.watch_histories"
	case fromEntity == "chorki_customers" && toEntity == "chorki_devices":
		return "chorki_customers.devices"

	// Reverse relationships back to customers (using ~ notation)
	case fromEntity == "chorki_subscriptions" && toEntity == "chorki_customers":
		return "~chorki_customers.subscriptions"
	case fromEntity == "chorki_devices" && toEntity == "chorki_customers":
		return "~chorki_customers.devices"
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_customers":
		return "~chorki_customers.watch_histories"

	// Content relationships
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_contents":
		return "chorki_watch_histories.content"
	case fromEntity == "chorki_contents" && toEntity == "chorki_watch_histories":
		return "~chorki_watch_histories.content"

	// Default fallback - construct relationship name dynamically
	default:
		// Check if this should be a reverse relationship
		if toEntity == "chorki_customers" {
			return "customers" // Simple name for reverse edge
		}
		// Generic relationship name: remove prefix and use simple name
		return strings.ReplaceAll(toEntity, "chorki_", "")
	}
}

// =============================================================================
// FINAL DQL GENERATION AND UTILITY METHODS
// =============================================================================

// GenerateDQLString converts a structured DQL query object into executable DQL string.
// This method handles final formatting, ensures proper DQL syntax, and creates
// a well-formatted query ready for execution against Dgraph.
func (c *Converter) GenerateDQLString(dqlQuery *models.DQLQuery) string {
	if len(dqlQuery.Queries) == 0 {
		return "{}"
	}

	var queryBlocks []string

	// Format each query block with proper indentation and structure
	for _, query := range dqlQuery.Queries {
		block := fmt.Sprintf("  %s(func: %s) %s {\n%s\n  }",
			query.Name,
			query.Function,
			query.Filter,
			query.Fields,
		)
		queryBlocks = append(queryBlocks, block)
	}

	// Combine all query blocks into final DQL
	return "{\n" + strings.Join(queryBlocks, "\n\n") + "\n}"
}

// AnalyzeComplexity returns the complexity analysis for a given query.
// This method exposes the internal complexity analyzer for external use,
// allowing callers to understand query performance implications before execution.
func (c *Converter) AnalyzeComplexity(jsonQuery *models.JSONQuery) *analyzer.ComplexityScore {
	return c.complexityAnalyzer.AnalyzeComplexity(jsonQuery)
}

// GetComplexityLimits returns the current complexity limits configuration.
// Useful for clients to understand the boundaries of acceptable query complexity
// and implement appropriate validation or optimization strategies.
func (c *Converter) GetComplexityLimits() map[string]int {
	return c.complexityAnalyzer.GetComplexityLimits()
}

// =============================================================================
// SPECIALIZED CONDITION BUILDERS (CONTINUED)
// =============================================================================

// buildNotInCondition constructs DQL conditions for NOT_IN operations.
// This method creates negated conditions for excluding specific values,
// handling both arrays and single values with proper NOT logic.
func (c *Converter) buildNotInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		// Handle array of values to exclude
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}
		// Negate multiple conditions with proper grouping
		if len(conditions) > 1 {
			return "NOT (" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return "NOT " + conditions[0]
		}
	default:
		// Handle single value to exclude
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
		// Expect exactly two values: [min, max]
		if len(v) == 2 {
			min := c.formatValue(v[0], mapping.DataType)
			max := c.formatValue(v[1], mapping.DataType)
			if min != "" && max != "" {
				// Combine min and max conditions with AND logic
				return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
					mapping.DgraphField, min, mapping.DgraphField, max)
			}
		}
	case map[string]interface{}:
		// Handle object-style range specification: {"min": value, "max": value}
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
