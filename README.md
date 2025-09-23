# JSON to DQL Converter

A Go-based tool that converts dynamic JSON queries to Dgraph DQL (Dgraph Query Language) format for the Chorki streaming platform.

## Features

- **Dynamic JSON Query Processing**: Accepts complex nested JSON queries with AND/OR combinations
- **Multi-Entity Support**: Generates DQL queries for multiple entity types (customers, subscriptions, watch histories, etc.)
- **Complex Filter Support**: Handles various operators (=, >=, <=, >, <, IN) and complex object filters
- **RESTful API**: Simple HTTP API for integration with frontend applications
- **Schema Validation**: Validates incoming queries against predefined schema
- **Extensible Design**: Easy to add new entity types and field mappings

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd jsonTodql
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### GET /
Returns API information and usage examples.

### GET /api/v1/health
Health check endpoint.

### GET /api/v1/schema
Returns available fields, operators, and example queries.

### POST /api/v1/convert
Converts JSON query to DQL format.

## Usage Examples

### Simple Query
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

### Complex Query with Nested Groups
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

### Complex Object Filter
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

## Supported Fields

### Customer Fields
- `age`: Customer age (int)
- `country`: Customer country (string)
- `device`: Device type (string)
- `app_version`: Application version (string)
- `last_login_days`: Days since last login (int)
- `email`: Customer email (string)
- `name`: Customer name (string)

### Subscription Fields
- `subscription_status`: Subscription status (string)
- `subscribed_package`: Package name (string)
- `package`: Package name (string)
- `status`: Status (string)

### Content Fields
- `watched_content`: Complex object with content_type and ids
- `favorite_genres`: Array of genre strings
- `content_type`: Type of content (string)
- `genre`: Content genre (array)
- `title`: Content title (string)

### Device Fields
- `device_type`: Type of device (string)
- `os_version`: Operating system version (string)

## Supported Operators

- `=`: Equals
- `>=`: Greater than or equal
- `<=`: Less than or equal
- `>`: Greater than
- `<`: Less than
- `IN`: In array/list
- `!=`: Not equal

## Entity Types

- `chorki_customers`: Customer information
- `chorki_subscriptions`: Subscription data
- `chorki_watch_histories`: Viewing history
- `chorki_contents`: Content metadata
- `chorki_devices`: Device information

## Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## Project Structure

```
jsonTodql/
├── main.go                 # Main application entry point
├── go.mod                  # Go module definition
├── models/
│   └── types.go           # Data structure definitions
├── config/
│   └── schema.go          # Schema configuration and mappings
├── converter/
│   ├── converter.go       # Core conversion logic
│   └── converter_test.go  # Unit tests
├── handlers/
│   └── query_handler.go   # HTTP request handlers
└── README.md
```

## Example DQL Output

For the input:
```json
{
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
}
```

The converter generates:
```dql
{
  customers(func: type(chorki_customers)) @filter((ge(chorki_customers.age, 21) OR eq(chorki_customers.country, "USA"))) {
    uid
    chorki_customers.id
    chorki_customers.name
    chorki_customers.email
    chorki_customers.age
    chorki_customers.country
    chorki_customers.device
    chorki_customers.app_version

    chorki_customers.subscriptions {
      uid
      chorki_subscriptions.id
      chorki_subscriptions.package
      chorki_subscriptions.status
      chorki_subscriptions.start_date
      chorki_subscriptions.end_date
    }

    chorki_customers.watch_histories {
      uid
      chorki_watch_histories.id
      chorki_watch_histories.content_id
      chorki_watch_histories.content_title
      chorki_watch_histories.type
      chorki_watch_histories.genre
      chorki_watch_histories.watch_date
    }

    chorki_customers.devices {
      uid
      chorki_devices.id
      chorki_devices.device_type
      chorki_devices.device_model
      chorki_devices.app_version
      chorki_devices.is_active
    }
  }
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License.