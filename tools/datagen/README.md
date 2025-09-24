# Chorki Data Generator

A production-ready, scalable data generator for the Chorki streaming platform. This tool generates realistic test data for customers, subscriptions, watch histories, devices, and content with proper relationships and validation.

## Features

🚀 **Production Ready**
- Configurable batch processing
- Memory-efficient data generation
- Comprehensive validation
- Progress tracking and logging
- Error handling and recovery

📊 **Scalable Architecture**
- Modular design with separation of concerns
- Configurable data distribution
- Multiple export formats (Dgraph, JSON, CSV)
- Direct Dgraph upload support

🎯 **Realistic Data**
- Proper relationships between entities
- Configurable data distribution
- Diverse demographics and usage patterns
- Realistic device and content metadata

## Project Structure

```
tools/datagen/
├── cmd/
│   └── main.go              # CLI entry point
├── pkg/
│   ├── generator.go         # Core generation logic
│   ├── data_factory.go      # Data generation functions
│   ├── progress.go          # Progress tracking
│   ├── validator.go         # Data validation
│   └── exporter.go          # Export functionality
├── configs/
│   └── datagen.yaml         # Configuration file
├── examples/
│   └── sample_config.yaml   # Example configuration
├── go.mod                   # Go module definition
└── README.md               # This file
```

## Quick Start

### 1. Installation

```bash
cd tools/datagen
go mod tidy
```

### 2. Generate Sample Configuration

```bash
go run cmd/main.go -generate-config
```

This creates a sample configuration at `configs/datagen.yaml`.

### 3. Basic Usage

```bash
# Generate 1000 customers and 500 contents with default settings
go run cmd/main.go

# Use custom parameters
go run cmd/main.go -customers=5000 -contents=2000 -batch=200

# Generate from configuration file
go run cmd/main.go -config=configs/datagen.yaml

# Dry run (no actual export)
go run cmd/main.go -dry-run -verbose
```

## Configuration

### Command Line Options

```bash
Usage: go run cmd/main.go [options]

Options:
  -config string         Configuration file path (default: configs/datagen.yaml)
  -customers int         Number of customers to generate (default: 1000)
  -contents int          Number of contents to generate (default: 500)
  -batch int            Batch size for processing (default: 100)
  -format string        Output format: dgraph, json, csv (default: dgraph)
  -output string        Output directory (default: ./output)
  -dgraph string        Dgraph server URL (default: localhost:9080)
  -validate            Enable data validation (default: true)
  -verbose             Enable verbose logging (default: true)
  -dry-run             Dry run mode - no actual export (default: false)
  -generate-config     Generate sample configuration file
```

### Configuration File

```yaml
# Generation parameters
customers: 1000
contents: 500
batch_size: 100

# Output settings
format: "dgraph"          # Options: dgraph, json, csv
output_dir: "./output"
dgraph_url: "localhost:9080"

# Processing options
validate: true
verbose: true
dry_run: false

# Data distribution settings
distribution:
  min_subscriptions_per_customer: 1
  max_subscriptions_per_customer: 3
  min_watch_histories_per_customer: 2
  max_watch_histories_per_customer: 15
  min_devices_per_customer: 1
  max_devices_per_customer: 4
  active_customer_ratio: 0.85
  premium_content_ratio: 0.4
```

## Generated Data Schema

### Customers (`chorki_customers`)
- **Basic Info**: ID, name, email, age, country, device
- **Activity**: app_version, last_login_days, is_active
- **Timestamps**: created_at, updated_at

### Subscriptions (`chorki_subscriptions`)
- **Plan Details**: package, status, price, currency
- **Dates**: start_date, end_date
- **Payment**: payment_method, auto_renewal, trial_period
- **Timestamps**: created_at, updated_at

### Watch Histories (`chorki_watch_histories`)
- **Content**: content_id, content_title, type, genre
- **Viewing**: watch_duration, completion_percentage
- **Timestamps**: watch_date, created_at

### Devices (`chorki_devices`)
- **Hardware**: device_type, device_model, os_version
- **App**: app_version, is_active, last_seen
- **Timestamps**: created_at, updated_at

### Contents (`chorki_contents`)
- **Metadata**: title, type, genre, duration, rating
- **Details**: language, description, release_date
- **Resources**: thumbnail_url, video_url
- **Business**: is_premium
- **Timestamps**: created_at, updated_at

## Output Formats

### 1. Dgraph Format (default)
Direct upload to Dgraph database with proper RDF/JSON mutation format.

### 2. JSON Format
Clean JSON files suitable for import into various systems.

### 3. CSV Format
Comma-separated values for analysis and import into spreadsheets.

## Examples

### Generate Small Dataset

```bash
go run cmd/main.go -customers=100 -contents=50 -batch=20 -verbose
```

### Generate Large Production Dataset

```bash
go run cmd/main.go -customers=50000 -contents=10000 -batch=500 -format=json
```

### Validation Only (No Export)

```bash
go run cmd/main.go -dry-run -validate -verbose
```

### Upload to Dgraph

```bash
go run cmd/main.go -dgraph=your-dgraph-server:9080 -format=dgraph
```

## Performance Considerations

- **Memory Usage**: Data is processed in batches to handle large datasets
- **CPU Usage**: Configurable concurrency for optimal performance
- **Network**: Batch uploads to minimize Dgraph load
- **Storage**: Efficient file writing with proper buffering

## Validation

The generator includes comprehensive validation:

- **Data Integrity**: Ensures all required fields are present
- **Relationships**: Validates foreign key relationships
- **Formats**: Checks email, date, and URL formats
- **Business Rules**: Validates completion percentages, ratings, etc.

## Error Handling

- **Graceful Degradation**: Continues processing on non-critical errors
- **Detailed Logging**: Comprehensive error messages and context
- **Recovery**: Automatic retry for transient failures
- **Validation**: Optional strict mode for development

## Troubleshooting

### Common Issues

1. **Dgraph Connection Failed**
   ```bash
   # Check if Dgraph is running
   curl http://localhost:8080/health
   
   # Verify the correct port (default: 9080 for Alpha, 8080 for query)
   go run cmd/main.go -dgraph=localhost:9080
   ```

2. **Out of Memory**
   ```bash
   # Reduce batch size
   go run cmd/main.go -batch=50
   ```

3. **Validation Errors**
   ```bash
   # Run with verbose logging to see details
   go run cmd/main.go -verbose -validate
   ```

### Performance Tuning

```bash
# For large datasets
go run cmd/main.go -customers=100000 -batch=1000 -validate=false

# For development/testing
go run cmd/main.go -customers=100 -batch=10 -verbose -dry-run
```

## Contributing

1. Add new data providers in `pkg/data_factory.go`
2. Extend validation rules in `pkg/validator.go`
3. Add new export formats in `pkg/exporter.go`
4. Update configuration structure in `pkg/generator.go`

## License

This tool is part of the Chorki project and follows the same licensing terms.