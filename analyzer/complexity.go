package analyzer

import (
	"fmt"
	"jsonTodql/models"
)

// ComplexityScore represents the complexity analysis result
type ComplexityScore struct {
	TotalScore     int                    `json:"total_score"`
	FilterCount    int                    `json:"filter_count"`
	MaxNesting     int                    `json:"max_nesting"`
	EntityCount    int                    `json:"entity_count"`
	ComplexObjects int                    `json:"complex_objects"`
	IsAcceptable   bool                   `json:"is_acceptable"`
	Warning        string                 `json:"warning,omitempty"`
	Breakdown      map[string]int         `json:"breakdown"`
}

// ComplexityAnalyzer analyzes query complexity
type ComplexityAnalyzer struct {
	MaxScore       int
	MaxFilters     int
	MaxNesting     int
	MaxEntities    int
	FieldMappings  map[string][]models.FieldMapping
}

// NewComplexityAnalyzer creates a new complexity analyzer with default limits
func NewComplexityAnalyzer(fieldMappings map[string][]models.FieldMapping) *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		MaxScore:      100,  // Maximum allowed complexity score
		MaxFilters:    50,   // Maximum number of filters
		MaxNesting:    5,    // Maximum nesting depth
		MaxEntities:   10,   // Maximum number of different entities
		FieldMappings: fieldMappings,
	}
}

// AnalyzeComplexity calculates the complexity score of a JSON query
func (ca *ComplexityAnalyzer) AnalyzeComplexity(query *models.JSONQuery) *ComplexityScore {
	score := &ComplexityScore{
		Breakdown: make(map[string]int),
	}

	// Count filters and analyze nesting
	entityMap := make(map[string]bool)
	score.FilterCount = ca.countFilters(query.Groups, entityMap, 0, &score.MaxNesting)
	score.EntityCount = len(entityMap)

	// Count complex objects (like watched_content)
	score.ComplexObjects = ca.countComplexObjects(query.Groups)

	// Calculate complexity breakdown
	score.Breakdown["filters"] = score.FilterCount * 2           // 2 points per filter
	score.Breakdown["nesting"] = score.MaxNesting * 5            // 5 points per nesting level
	score.Breakdown["entities"] = score.EntityCount * 3          // 3 points per entity type
	score.Breakdown["complex_objects"] = score.ComplexObjects * 8 // 8 points per complex object

	// Calculate total score
	score.TotalScore = score.Breakdown["filters"] + 
					  score.Breakdown["nesting"] + 
					  score.Breakdown["entities"] + 
					  score.Breakdown["complex_objects"]

	// Determine if query is acceptable
	score.IsAcceptable = ca.isAcceptable(score)
	if !score.IsAcceptable {
		score.Warning = ca.generateWarning(score)
	}

	return score
}

// countFilters recursively counts filters and tracks nesting depth
func (ca *ComplexityAnalyzer) countFilters(groups []models.Group, entityMap map[string]bool, currentDepth int, maxDepth *int) int {
	if currentDepth > *maxDepth {
		*maxDepth = currentDepth
	}

	filterCount := 0
	
	for _, group := range groups {
		// Count filters in this group
		for _, filter := range group.Filters {
			filterCount++
			
			// Track entity types
			if mappings, exists := ca.FieldMappings[filter.Field]; exists {
				for _, mapping := range mappings {
					entityMap[mapping.EntityType] = true
				}
			}
		}

		// Recursively count nested groups
		if len(group.Groups) > 0 {
			filterCount += ca.countFilters(group.Groups, entityMap, currentDepth+1, maxDepth)
		}
	}

	return filterCount
}

// countComplexObjects counts complex object filters
func (ca *ComplexityAnalyzer) countComplexObjects(groups []models.Group) int {
	count := 0
	
	for _, group := range groups {
		for _, filter := range group.Filters {
			// Check for complex objects like watched_content
			if filter.Field == "watched_content" {
				count++
			}
			
			// Check for other complex value types
			switch v := filter.Value.(type) {
			case map[string]interface{}:
				if len(v) > 1 { // More than one key indicates complexity
					count++
				}
			case []interface{}:
				if len(v) > 10 { // Large arrays add complexity
					count++
				}
			}
		}

		// Recursively check nested groups
		if len(group.Groups) > 0 {
			count += ca.countComplexObjects(group.Groups)
		}
	}

	return count
}

// isAcceptable determines if the query complexity is within acceptable limits
func (ca *ComplexityAnalyzer) isAcceptable(score *ComplexityScore) bool {
	if score.TotalScore > ca.MaxScore {
		return false
	}
	if score.FilterCount > ca.MaxFilters {
		return false
	}
	if score.MaxNesting > ca.MaxNesting {
		return false
	}
	if score.EntityCount > ca.MaxEntities {
		return false
	}
	return true
}

// generateWarning generates a warning message for complex queries
func (ca *ComplexityAnalyzer) generateWarning(score *ComplexityScore) string {
	if score.TotalScore > ca.MaxScore {
		return fmt.Sprintf("Query complexity score (%d) exceeds maximum allowed (%d). Consider simplifying the query.", 
			score.TotalScore, ca.MaxScore)
	}
	if score.FilterCount > ca.MaxFilters {
		return fmt.Sprintf("Number of filters (%d) exceeds maximum allowed (%d).", 
			score.FilterCount, ca.MaxFilters)
	}
	if score.MaxNesting > ca.MaxNesting {
		return fmt.Sprintf("Query nesting depth (%d) exceeds maximum allowed (%d).", 
			score.MaxNesting, ca.MaxNesting)
	}
	if score.EntityCount > ca.MaxEntities {
		return fmt.Sprintf("Number of entity types (%d) exceeds maximum allowed (%d).", 
			score.EntityCount, ca.MaxEntities)
	}
	return "Query complexity exceeds acceptable limits."
}

// GetComplexityLimits returns the current complexity limits
func (ca *ComplexityAnalyzer) GetComplexityLimits() map[string]int {
	return map[string]int{
		"max_score":    ca.MaxScore,
		"max_filters":  ca.MaxFilters,
		"max_nesting":  ca.MaxNesting,
		"max_entities": ca.MaxEntities,
	}
}

// SetComplexityLimits allows updating complexity limits
func (ca *ComplexityAnalyzer) SetComplexityLimits(maxScore, maxFilters, maxNesting, maxEntities int) {
	if maxScore > 0 {
		ca.MaxScore = maxScore
	}
	if maxFilters > 0 {
		ca.MaxFilters = maxFilters
	}
	if maxNesting > 0 {
		ca.MaxNesting = maxNesting
	}
	if maxEntities > 0 {
		ca.MaxEntities = maxEntities
	}
}