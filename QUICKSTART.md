# Quick Start Guide - JSON to DQL Converter

## Overview
This tool converts dynamic JSON queries into Dgraph DQL (Dgraph Query Language) format for the Chorki streaming platform. It supports complex nested filters with AND/OR combinations across multiple entity types.

## Running the Application

### 1. Start the API Server
```bash
cd c:\SHAHARIAZ\GOLANG\jsonTodql
go run main.go
```

The server will start on `http://localhost:8080`

### 2. Test the Converter Directly
```bash
cd c:\SHAHARIAZ\GOLANG\jsonTodql\demo
go run main.go
```

This will run the demo showing various conversion examples.

### 3. Run Unit Tests
```bash
cd c:\SHAHARIAZ\GOLANG\jsonTodql
go test ./converter/ -v
```

## API Endpoints

- **GET** `/` - API information and examples
- **GET** `/api/v1/health` - Health check
- **GET** `/api/v1/schema` - Schema information and available fields
- **POST** `/api/v1/convert` - Convert JSON query to DQL

## Example Conversions

### Simple Filter
**Input:**
```json
{
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
}
```

**Output:**
```dql
{
  customers(func: type(chorki_customers)) @filter((ge(chorki_customers.age, 21) OR (eq(chorki_customers.country, "USA") OR eq(chorki_customers.country, "UK") OR eq(chorki_customers.country, "Canada")))) {
    uid
    chorki_customers.id
    chorki_customers.name
    chorki_customers.email
    # ... more fields and relationships
  }
}
```

### Complex Nested Filter
**Input:**
```json
{
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
}
```

**Output:**
```dql
{
  customers(func: type(chorki_customers)) @filter(((eq(chorki_customers.device, "iOS") OR eq(chorki_customers.device, "Android")) AND ge(chorki_customers.app_version, "5.0.0") AND (lt(chorki_customers.age, 18) OR eq(chorki_customers.country, "Bangladesh")))) {
    # ... customer fields and relationships
  }
  
  devices(func: type(chorki_devices)) @filter(((eq(chorki_devices.device_type, "iOS") OR eq(chorki_devices.device_type, "Android")) AND ge(chorki_devices.app_version, "5.0.0"))) {
    # ... device fields and relationships
  }
}
```

### Complex Object Filter (Watched Content)
**Input:**
```json
{
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
}
```

**Output:**
```dql
{
  watch_histories(func: type(chorki_watch_histories)) @filter((eq(chorki_watch_histories.type, "Movie") AND uid_in(chorki_watch_histories.content_id, 111, 222, 333))) {
    # ... watch history fields and relationships
  }
}
```

## Supported Features

### Operators
- `=` (equals)
- `>=` (greater than or equal)
- `<=` (less than or equal)
- `>` (greater than)
- `<` (less than)
- `IN` (in array/list)
- `!=` (not equal)

### Combination Logic
- `AND` - All conditions must be true
- `OR` - At least one condition must be true
- Unlimited nesting of groups

### Entity Types
- `chorki_customers` - Customer information
- `chorki_subscriptions` - Subscription data
- `chorki_watch_histories` - Viewing history
- `chorki_contents` - Content metadata
- `chorki_devices` - Device information

### Field Types
- **String fields**: `name`, `email`, `country`, `device`, etc.
- **Integer fields**: `age`, `last_login_days`, `content_id`, etc.
- **Array fields**: `genre`, `favorite_genres`, device lists
- **Complex objects**: `watched_content` with nested properties

### Relationship Traversals
The converter automatically includes related entities:
- Customers → Subscriptions, Watch Histories, Devices
- Subscriptions → Customers
- Watch Histories → Customers, Contents
- Devices → Customers

## Project Structure
```
jsonTodql/
├── main.go                 # API server entry point
├── demo/
│   └── main.go            # Demo program
├── models/
│   └── types.go           # Data structures
├── config/
│   └── schema.go          # Schema configuration
├── converter/
│   ├── converter.go       # Core conversion logic
│   └── converter_test.go  # Unit tests
├── handlers/
│   └── query_handler.go   # HTTP handlers
├── schema.dgraph          # Dgraph schema file
└── README.md              # Full documentation
```

## Testing Your Queries

1. **Use the demo program** for quick testing without the web server
2. **Use the API endpoints** for integration testing
3. **Check the unit tests** for additional examples

## Next Steps

1. **Extend field mappings** in `config/schema.go` for new fields
2. **Add new entity types** by updating the schema configuration
3. **Customize field selections** by modifying default fields
4. **Add new operators** by extending the operator mappings