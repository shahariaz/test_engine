# JSON to DQL Converter - Complete Architecture Guide

## Table of Contents
1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Data Flow](#data-flow)
4. [Core Components](#core-components)
5. [Conversion Process](#conversion-process)
6. [Code Reading Guide](#code-reading-guide)
7. [Implementation Details](#implementation-details)
8. [Building Your Own](#building-your-own)

## Overview

This system converts JSON-based query structures into Dgraph Query Language (DQL) queries. It provides a RESTful API that accepts JSON queries and returns executable DQL queries along with complexity analysis.

### Key Features
- **Multi-entity queries** with relationship traversal
- **Complex nested filtering** with AND/OR logic
- **Query complexity analysis** for performance protection
- **Schema validation** and field mapping
- **Production-ready** safety mechanisms

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   JSON Query    │───▶│   Converter     │───▶│   DQL Query     │
│   (Input)       │    │   (Core Logic)  │    │   (Output)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼
                    ┌─────────────────┐
                    │   Components    │
                    │                 │
                    │ • Analyzer      │
                    │ • Validator     │
                    │ • Schema Config │
                    │ • Field Mapping │
                    └─────────────────┘
```

### Directory Structure
```
test_engine/
├── main.go                 # API server entry point
├── models/
│   └── types.go           # Data structures and models
├── converter/
│   └── converter.go       # Core conversion logic
├── analyzer/
│   └── complexity.go      # Query complexity analysis
├── validation/
│   └── validator.go       # Query validation
├── config/
│   └── schema.go          # Schema and mapping configuration
├── handlers/
│   └── query_handler.go   # HTTP request handlers
└── dgraph/
    └── schema.graphql     # Dgraph schema definition
```

## Data Flow

### 1. Request Flow
```
HTTP Request → Handler → Validator → Converter → DQL Query → Response
```

### 2. Conversion Pipeline
```
JSON Query
    ↓
Complexity Analysis
    ↓
Entity Discovery
    ↓
Filter Building
    ↓
Field Selection
    ↓
DQL Generation
    ↓
Final DQL Query
```

## Core Components

### 1. Models (`models/types.go`)

#### JSONQuery Structure
```go
type JSONQuery struct {
    CombineWith string  `json:"combine_with"` // "AND" or "OR"
    Groups      []Group `json:"groups"`
}

type Group struct {
    CombineWith string   `json:"combine_with"` // "AND" or "OR" 
    Filters     []Filter `json:"filters"`
    Groups      []Group  `json:"groups"`       // Nested groups
}

type Filter struct {
    Field string      `json:"field"`  // Field name to filter on
    Op    string      `json:"op"`     // Operator: "=", ">=", "<=", ">", "<", "IN"
    Value interface{} `json:"value"`  // Filter value
}
```

**Purpose**: Defines the structure for incoming JSON queries. Supports nested groups for complex logical combinations.

#### DQLQuery Structure
```go
type DQLQuery struct {
    Queries []EntityQuery `json:"queries"`
}

type EntityQuery struct {
    Name     string `json:"name"`     // Query identifier
    Type     string `json:"type"`     // Entity type (e.g., "chorki_customers")
    Function string `json:"function"` // DQL function (e.g., "type(chorki_customers)")
    Filter   string `json:"filter"`   // DQL filter expression
    Fields   string `json:"fields"`   // Field selection
}
```

**Purpose**: Represents the output DQL query structure with separate queries for each entity type.

### 2. Schema Configuration (`config/schema.go`)

#### Field Mapping
```go
type FieldMapping struct {
    JSONField      string `json:"json_field"`      // Input field name
    DgraphField    string `json:"dgraph_field"`    // Dgraph predicate name
    EntityType     string `json:"entity_type"`     // Target entity type
    DataType       string `json:"data_type"`       // Data type for validation
    IsRelationship bool   `json:"is_relationship"` // Whether it's a relationship
}
```

**Purpose**: Maps JSON field names to Dgraph predicates and defines their properties.

### 3. Converter (`converter/converter.go`)

The converter is the heart of the system. It orchestrates the entire conversion process.

#### Main Components:
- **Schema mappings**: Field and operator mappings
- **Complexity analyzer**: Performance protection
- **Version fields**: Special handling for version comparisons
- **Reverse predicates**: Relationship traversal support

### 4. Analyzer (`analyzer/complexity.go`)

#### ComplexityScore
```go
type ComplexityScore struct {
    Score        int    `json:"score"`         // Numerical complexity score
    MaxScore     int    `json:"max_score"`     // Maximum allowed score
    IsAcceptable bool   `json:"is_acceptable"` // Whether query is acceptable
    Warning      string `json:"warning"`       // Warning message if complex
    Breakdown    map[string]int `json:"breakdown"` // Score breakdown by category
}
```

**Purpose**: Analyzes and scores query complexity to prevent performance issues.

## Conversion Process

### Step 1: Input Validation
```go
func (v *Validator) ValidateJSONQuery(query *models.JSONQuery) error {
    // Validates structure, required fields, and data types
    return v.validateGroups(query.Groups)
}
```

### Step 2: Complexity Analysis
```go
func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.DQLQuery, error) {
    // Analyze complexity first
    complexityScore := c.complexityAnalyzer.AnalyzeComplexity(jsonQuery)
    if !complexityScore.IsAcceptable {
        return nil, fmt.Errorf("query too complex: %s", complexityScore.Warning)
    }
    // ... continue with conversion
}
```

### Step 3: Entity Discovery
```go
func (c *Converter) getInvolvedEntityTypes(jsonQuery *models.JSONQuery) []string {
    entityTypeMap := make(map[string]bool)
    // Traverse all groups and filters to find entity types
    c.collectEntityTypesFromGroups(jsonQuery.Groups, entityTypeMap)
    // Return unique entity types
}
```

### Step 4: Filter Building
```go
func (c *Converter) buildFilterForEntity(jsonQuery *models.JSONQuery, entityType string) (string, error) {
    // Build DQL filter expressions for the specific entity type
    return c.buildGroupFilter(jsonQuery.Groups, entityType, jsonQuery.CombineWith)
}
```

### Step 5: DQL Generation
```go
// For each entity type, create an EntityQuery
query := models.EntityQuery{
    Name:     c.getQueryName(entityType),
    Type:     entityType,
    Function: fmt.Sprintf("type(%s)", entityType),
    Filter:   filter,
    Fields:   c.buildFieldsSelection(entityType),
}
```

## Code Reading Guide

### Start Here: Understanding the Flow

1. **Begin with `main.go`**
   - See how the server is set up
   - Understand the API endpoints
   - Follow the handler registration

2. **Read `models/types.go`**
   - Understand the data structures
   - See how JSON input is modeled
   - Understand the DQL output structure

3. **Examine `handlers/query_handler.go`**
   - See how HTTP requests are processed
   - Understand the validation flow
   - Follow the conversion pipeline

4. **Deep dive into `converter/converter.go`**
   - Start with `NewConverter()` to see initialization
   - Follow `ConvertToDQL()` method step by step
   - Understand entity discovery process
   - See how filters are built

5. **Study `analyzer/complexity.go`**
   - Understand complexity scoring
   - See performance protection mechanisms

### Key Methods to Understand

#### 1. ConvertToDQL (Main Entry Point)
```go
func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.DQLQuery, error)
```
**What it does**: Orchestrates the entire conversion process
**Key steps**: 
- Complexity analysis
- Entity discovery  
- Filter building
- Query construction

#### 2. getInvolvedEntityTypes (Entity Discovery)
```go
func (c *Converter) getInvolvedEntityTypes(jsonQuery *models.JSONQuery) []string
```
**What it does**: Determines which Dgraph entities are involved in the query
**How**: Traverses all filters and maps fields to entity types

#### 3. buildFilterForEntity (Filter Construction)
```go
func (c *Converter) buildFilterForEntity(jsonQuery *models.JSONQuery, entityType string) (string, error)
```
**What it does**: Builds DQL filter expressions for a specific entity type
**How**: Recursively processes groups and filters, applying AND/OR logic

#### 4. processFilter (Individual Filter Processing)
```go
func (c *Converter) processFilter(filter models.Filter, entityType string) (string, error)
```
**What it does**: Converts a single JSON filter to DQL
**How**: Maps operators, handles special cases, formats values

### Understanding the Recursive Logic

The system uses recursive functions to handle nested groups:

```go
func (c *Converter) buildGroupFilter(groups []models.Group, entityType string, combineWith string) (string, error) {
    var conditions []string
    
    for _, group := range groups {
        // Process filters in this group
        for _, filter := range group.Filters {
            if condition := c.processFilter(filter, entityType); condition != "" {
                conditions = append(conditions, condition)
            }
        }
        
        // Recursively process nested groups
        if len(group.Groups) > 0 {
            if nestedCondition := c.buildGroupFilter(group.Groups, entityType, group.CombineWith); nestedCondition != "" {
                conditions = append(conditions, fmt.Sprintf("(%s)", nestedCondition))
            }
        }
    }
    
    // Combine conditions with AND/OR
    return strings.Join(conditions, fmt.Sprintf(" %s ", combineWith)), nil
}
```

## Implementation Details

### Field Mapping Process

1. **JSON Field → Dgraph Predicate**
   ```go
   // Input: "customer_name"
   // Mapping: customer_name → name (for chorki_customers entity)
   // Output: DQL uses "name" predicate
   ```

2. **Entity Type Resolution**
   ```go
   // Field "customer_name" maps to:
   // - EntityType: "chorki_customers"  
   // - DgraphField: "name"
   // - DataType: "string"
   ```

### Operator Mapping

```go
var operatorMappings = map[string]string{
    "=":  "eq",     // Equal
    ">=": "ge",     // Greater than or equal
    "<=": "le",     // Less than or equal  
    ">":  "gt",     // Greater than
    "<":  "lt",     // Less than
    "IN": "anyofterms", // Contains any of terms
}
```

### Complex Filter Example

**Input JSON:**
```json
{
    "combine_with": "AND",
    "groups": [
        {
            "combine_with": "OR",
            "filters": [
                {"field": "customer_name", "op": "=", "value": "John"},
                {"field": "customer_email", "op": "=", "value": "john@example.com"}
            ]
        },
        {
            "combine_with": "AND", 
            "filters": [
                {"field": "age", "op": ">=", "value": 18},
                {"field": "status", "op": "=", "value": "active"}
            ]
        }
    ]
}
```

**Generated DQL:**
```dql
{
  customers(func: type(chorki_customers)) @filter(
    (eq(name, "John") OR eq(email, "john@example.com")) AND 
    (ge(age, 18) AND eq(status, "active"))
  ) {
    uid
    name
    email
    age
    status
  }
}
```

## Building Your Own

### 1. Core Requirements

#### Dependencies
```go
import (
    "github.com/gin-gonic/gin"     // Web framework
    "github.com/go-playground/validator/v10" // Validation
)
```

#### Database Schema
- Design your Dgraph schema first
- Define entity types and predicates
- Plan relationships between entities

### 2. Implementation Steps

#### Step 1: Define Data Models
```go
// Define your JSON input structure
type YourJSONQuery struct {
    // Your query structure
}

// Define your DQL output structure  
type YourDQLQuery struct {
    // Your DQL structure
}
```

#### Step 2: Create Field Mappings
```go
// Map JSON fields to Dgraph predicates
var fieldMappings = map[string][]FieldMapping{
    "your_field": {
        {
            JSONField:   "your_field",
            DgraphField: "your_predicate", 
            EntityType:  "your_entity_type",
            DataType:    "string",
        },
    },
}
```

#### Step 3: Implement Converter Logic
```go
type YourConverter struct {
    schema    *SchemaInfo
    operators map[string]string
}

func (c *YourConverter) Convert(query *YourJSONQuery) (*YourDQLQuery, error) {
    // 1. Validate input
    // 2. Analyze complexity
    // 3. Discover entities
    // 4. Build filters
    // 5. Generate DQL
}
```

#### Step 4: Add Complexity Analysis
```go
type ComplexityAnalyzer struct {
    maxDepth   int
    maxFields  int
    maxFilters int
}

func (ca *ComplexityAnalyzer) Analyze(query *YourJSONQuery) *ComplexityScore {
    // Calculate complexity score
    // Check against limits
    // Generate warnings
}
```

#### Step 5: Create HTTP Handlers
```go
func (h *Handler) ConvertQuery(c *gin.Context) {
    var query YourJSONQuery
    if err := c.ShouldBindJSON(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    dqlQuery, err := h.converter.Convert(&query)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, dqlQuery)
}
```

### 3. Testing Strategy

#### Unit Tests
```go
func TestConverter_ConvertToDQL(t *testing.T) {
    converter := NewConverter()
    
    tests := []struct {
        name     string
        input    *JSONQuery
        expected *DQLQuery
        wantErr  bool
    }{
        // Define test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := converter.ConvertToDQL(tt.input)
            // Assert results
        })
    }
}
```

#### Integration Tests
```go
func TestAPI_ConvertQuery(t *testing.T) {
    // Test the full HTTP API
    router := setupRouter()
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/convert", bytes.NewBuffer(jsonData))
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

### 4. Production Considerations

#### Error Handling
```go
// Always provide meaningful error messages
if err != nil {
    return nil, fmt.Errorf("failed to process filter for field %s: %w", filter.Field, err)
}
```

#### Logging
```go
import "log/slog"

slog.Info("Converting query", "entities", len(entities), "complexity", score.Score)
```

#### Configuration
```go
type Config struct {
    MaxComplexity int    `env:"MAX_COMPLEXITY" envDefault:"100"`
    DgraphURL     string `env:"DGRAPH_URL" envDefault:"localhost:9080"`
}
```

#### Performance Monitoring
```go
start := time.Now()
defer func() {
    duration := time.Since(start)
    slog.Debug("Query conversion completed", "duration", duration)
}()
```

## Advanced Topics

### 1. Relationship Traversal
```go
// Handle relationships between entities
func (c *Converter) handleRelationship(field string, entityType string) string {
    if reverseEdge, exists := c.reversePredicates[field]; exists {
        return fmt.Sprintf("~%s", reverseEdge)
    }
    return field
}
```

### 2. Custom Operators
```go
// Add support for custom operators
func (c *Converter) handleCustomOperator(op string, value interface{}) string {
    switch op {
    case "CONTAINS":
        return fmt.Sprintf("alloftext(%s, %q)", field, value)
    case "REGEX":
        return fmt.Sprintf("regexp(%s, /%v/)", field, value)
    default:
        return c.standardOperator(op, value)
    }
}
```

### 3. Query Optimization
```go
// Optimize query structure
func (c *Converter) optimizeQuery(query *DQLQuery) *DQLQuery {
    // Remove redundant conditions
    // Reorder for better performance
    // Combine similar filters
    return query
}
```

This guide provides everything you need to understand and rebuild this system from scratch. Start by reading the code in the order suggested, and use this documentation as a reference for understanding the architecture and implementation details.