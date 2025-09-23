# Chorki JSON to DQL Converter# JSON to DQL Converter



A high-performance Go API that converts complex JSON queries to Dgraph Query Language (DQL) and executes them against Dgraph.A Go-based tool that converts dynamic JSON queries to Dgraph DQL (Dgraph Query Language) format for the Chorki streaming platform.



## Features## Features



- **Complex Query Support**: Handles nested groups, multiple operators, and complex filtering logic- **Dynamic JSON Query Processing**: Accepts complex nested JSON queries with AND/OR combinations

- **Multi-Entity Queries**: Supports customers, subscriptions, watch histories, and content data- **Multi-Entity Support**: Generates DQL queries for multiple entity types (customers, subscriptions, watch histories, etc.)

- **Real-time Execution**: Direct integration with Dgraph for immediate query execution- **Complex Filter Support**: Handles various operators (=, >=, <=, >, <, IN) and complex object filters

- **Validation & Analysis**: Built-in query validation and complexity analysis- **RESTful API**: Simple HTTP API for integration with frontend applications

- **Hot Reload**: Development environment with automatic server restart on code changes- **Schema Validation**: Validates incoming queries against predefined schema

- **Extensible Design**: Easy to add new entity types and field mappings

## Quick Start

## Installation

### Prerequisites

- Go 1.21+1. Clone the repository:

- Docker & Docker Compose```bash

git clone <repository-url>

### Setupcd jsonTodql

1. **Clone and setup**:```

   ```bash

   git clone <repository>2. Install dependencies:

   cd test_engine```bash

   ```go mod tidy

```

2. **Start Dgraph**:

   ```bash3. Run the application:

   docker-compose up -d```bash

   ```go run main.go

```

3. **Load schema and data**:

   ```bashThe server will start on port 8080.

   # Load schema

   curl -X POST localhost:8080/alter --data-binary '@simple_schema.dgraph'## API Endpoints

   

   # Load sample data### GET /

   curl -X POST localhost:8080/mutate?commitNow=true -H "Content-Type: application/json" --data-binary '@proper_dataset.json'Returns API information and usage examples.

   ```

### GET /api/v1/health

4. **Start the API server**:Health check endpoint.

   ```bash

   # With hot reload (recommended for development)### GET /api/v1/schema

   airReturns available fields, operators, and example queries.

   

   # Or standard Go run### POST /api/v1/convert

   go run main.goConverts JSON query to DQL format.

   ```

## Usage Examples

5. **Test the API**:

   ```bash### Simple Query

   curl http://localhost:8090/api/v1/health```json

   ```{

  "combine_with": "AND",

## API Endpoints  "groups": [

    {

- `GET /api/v1/health` - Health check      "combine_with": "OR",

- `GET /api/v1/schema` - Get available fields and operators      "filters": [

- `POST /api/v1/execute` - Execute JSON query against Dgraph        {"field": "age", "op": ">=", "value": 21},

- `POST /api/v1/convert` - Convert JSON to DQL (without execution)        {"field": "country", "op": "IN", "value": ["USA", "UK", "Canada"]}

- `POST /api/v1/validate` - Validate JSON query structure      ]

- `POST /api/v1/analyze` - Analyze query complexity    }

  ]

## Example Usage}

```

### JSON Query Format

```json### Complex Query with Nested Groups

{```json

  "combine_with": "AND",{

  "groups": [  "combine_with": "OR",

    {  "groups": [

      "combine_with": "OR",    {

      "filters": [      "combine_with": "AND",

        {"field": "age", "op": ">=", "value": 25},      "filters": [

        {"field": "country", "op": "IN", "value": ["USA", "Canada"]}        {"field": "device", "op": "IN", "value": ["iOS", "Android"]},

      ]        {"field": "app_version", "op": ">=", "value": "5.0.0"}

    },      ],

    {      "groups": [

      "combine_with": "AND",        {

      "filters": [          "combine_with": "OR",

        {"field": "is_active", "op": "=", "value": true}          "filters": [

      ]            {"field": "age", "op": "<", "value": 18},

    }            {"field": "country", "op": "=", "value": "Bangladesh"}

  ]          ]

}        }

```      ]

    }

### Test with curl  ]

```bash}

curl -X POST http://localhost:8090/api/v1/execute \```

  -H "Content-Type: application/json" \

  -d '{### Complex Object Filter

    "combine_with": "AND",```json

    "groups": [{

      {  "combine_with": "AND",

        "combine_with": "AND",  "groups": [

        "filters": [    {

          {"field": "country", "op": "=", "value": "USA"}      "combine_with": "OR",

        ]      "filters": [

      }        {

    ]          "field": "watched_content",

  }'          "op": "IN",

```          "value": {

            "content_type": "Movie",

## Development            "ids": [111, 222, 333]

          }

### Hot Reload Setup        }

The project uses Air for hot reloading during development:      ]

    }

```bash  ]

# Install Air (one-time setup)}

go install github.com/air-verse/air@latest```



# Start with hot reload## Supported Fields

air

### Customer Fields

# Or use helper scripts- `age`: Customer age (int)

./dev.sh    # Linux/Mac- `country`: Customer country (string)

./dev.ps1   # Windows PowerShell- `device`: Device type (string)

```- `app_version`: Application version (string)

- `last_login_days`: Days since last login (int)

### Project Structure- `email`: Customer email (string)

```- `name`: Customer name (string)

├── main.go              # Application entry point

├── handlers/            # HTTP handlers### Subscription Fields

├── converter/           # JSON to DQL conversion logic- `subscription_status`: Subscription status (string)

├── config/              # Schema configuration- `subscribed_package`: Package name (string)

├── models/              # Data models- `package`: Package name (string)

├── validation/          # Query validation- `status`: Status (string)

├── analyzer/            # Complexity analysis

├── dgraph/              # Dgraph client### Content Fields

├── utils/               # Utility functions- `watched_content`: Complex object with content_type and ids

├── scripts/             # Development scripts- `favorite_genres`: Array of genre strings

└── tests/               # Test files- `content_type`: Type of content (string)

```- `genre`: Content genre (array)

- `title`: Content title (string)

### Available Operators

- Comparison: `=`, `!=`, `>`, `>=`, `<`, `<=`### Device Fields

- Array: `IN`, `NOT_IN`- `device_type`: Type of device (string)

- Text: `LIKE`, `ILIKE`, `CONTAINS`, `REGEX`- `os_version`: Operating system version (string)

- Pattern: `STARTS_WITH`, `ENDS_WITH`

- Range: `BETWEEN`## Supported Operators

- Null: `IS_NULL`, `IS_NOT_NULL`

- `=`: Equals

### Supported Fields- `>=`: Greater than or equal

- **Customer**: age, country, city, device, email, name, is_active, app_version, last_login_days- `<=`: Less than or equal

- **Subscription**: package, status, price, currency, payment_method, auto_renewal, trial_period- `>`: Greater than

- **Content**: title, type, genre, rating, duration, release_year- `<`: Less than

- **Watch History**: content_id, completion_percentage, device_used, quality- `IN`: In array/list

- `!=`: Not equal

## Testing

## Entity Types

Sample queries are available in `postman_queries.json` for manual testing with Postman or any REST client.

- `chorki_customers`: Customer information

## License- `chorki_subscriptions`: Subscription data

- `chorki_watch_histories`: Viewing history

MIT License- `chorki_contents`: Content metadata
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