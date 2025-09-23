# 🚀 Dgraph Integration Setup Guide

## 🎯 Overview
This guide shows you how to set up the complete Dgraph integration for the JSON to DQL converter, including:
- ✅ Dgraph database with sample data
- ✅ Ratel UI for database management  
- ✅ `/execute` endpoint that runs queries against real data

## 📋 Prerequisites
- Docker and Docker Compose installed
- Go 1.21+ installed
- Ports 8000, 8080, 9080, 5080, 6080 available

## 🚀 Quick Start

### 1. Start Dgraph with Docker Compose
```bash
# From the project root directory
docker-compose up -d

# Check that all services are running
docker-compose ps
```

Expected output:
```
Name                Command             State           Ports
-------------------------------------------------------------------
dgraph-alpha       dgraph alpha ...    Up      0.0.0.0:8080->8080/tcp, 0.0.0.0:9080->9080/tcp
dgraph-zero        dgraph zero ...     Up      0.0.0.0:5080->5080/tcp, 0.0.0.0:6080->6080/tcp  
dgraph-ratel       dgraph-ratel ...    Up      0.0.0.0:8000->8000/tcp
dgraph-init        sh -c ...           Exit 0
```

### 2. Verify Schema and Data Loading
```bash
# Check if schema was applied successfully
curl -X POST localhost:8080/admin/schema

# Check if sample data was loaded
curl -X POST localhost:8080/query -d '{ 
  customers(func: type(chorki_customers)) { 
    uid 
    chorki_customers.name 
    chorki_customers.email 
  } 
}'
```

### 3. Start the API Server
```bash
# Start the Go API server
go run main.go
```

Expected output:
```
✅ Connected to Dgraph at localhost:9080
🚀 JSON to DQL Converter API starting on port :8080
📋 Available endpoints:
   POST /api/v1/execute     - Convert JSON to DQL and execute against Dgraph
   ...
```

## 🎮 Testing the Complete Flow

### Test 1: Basic Customer Query
```bash
curl -X POST http://localhost:8080/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "combine_with": "OR",
    "groups": [
      {
        "combine_with": "AND", 
        "filters": [
          {"field": "age", "op": "<", "value": 18},
          {"field": "country", "op": "=", "value": "Bangladesh"}
        ]
      }
    ]
  }'
```

Expected response:
```json
{
  "success": true,
  "data": {
    "customers": [
      {
        "uid": "0x1",
        "chorki_customers.name": "Jane Smith",
        "chorki_customers.email": "jane.smith@example.com",
        "chorki_customers.age": 17,
        "chorki_customers.country": "Bangladesh"
      }
    ]
  },
  "query_info": {
    "dql": "{ customers(func: type(chorki_customers)) @filter(...) { ... } }",
    "query_time": "2.5ms",
    "stats": {
      "result_count": 1,
      "total_queries": 1
    }
  }
}
```

### Test 2: Premium Subscription Query
```bash
curl -X POST http://localhost:8080/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "combine_with": "AND",
    "groups": [
      {
        "combine_with": "AND",
        "filters": [
          {"field": "subscribed_package", "op": "=", "value": "Premium"},
          {"field": "subscription_status", "op": "=", "value": "trial"}
        ]
      }
    ]
  }'
```

### Test 3: Version Comparison Query  
```bash
curl -X POST http://localhost:8080/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "combine_with": "AND",
    "groups": [
      {
        "combine_with": "AND",
        "filters": [
          {"field": "device", "op": "IN", "value": ["iOS", "Android"]},
          {"field": "app_version", "op": ">=", "value": "5.0.0"}
        ]
      }
    ]
  }'
```

## 🎯 Ratel UI Access

### Open Ratel Web Interface
- URL: http://localhost:8000
- Server: `localhost:8080` (Dgraph Alpha endpoint)

### Sample Queries to Try in Ratel:

#### 1. All Customers with Subscriptions
```dql
{
  customers(func: type(chorki_customers)) {
    uid
    chorki_customers.name
    chorki_customers.email
    chorki_customers.subscriptions {
      chorki_subscriptions.package
      chorki_subscriptions.status
    }
  }
}
```

#### 2. Premium Subscriptions
```dql
{
  subscriptions(func: type(chorki_subscriptions)) @filter(eq(chorki_subscriptions.package, "Premium")) {
    uid
    chorki_subscriptions.package
    chorki_subscriptions.status
    ~chorki_customers.subscriptions {
      chorki_customers.name
      chorki_customers.email
    }
  }
}
```

## 🔧 Schema Management

### View Current Schema
```bash
curl -X POST localhost:8080/admin/schema
```

### Update Schema (if needed)
```bash
curl -X POST localhost:8080/admin/schema --data-binary @dgraph/schema.graphql
```

### Add More Sample Data
```bash
curl -X POST localhost:8080/mutate?commitNow=true \
  -H "Content-Type: application/rdf" \
  --data-binary @dgraph/sample_data.rdf
```

## 📊 Available Sample Data

The sample dataset includes:

### Customers (3 records)
- John Doe (25, Bangladesh, iOS, Premium trial)
- Jane Smith (17, Bangladesh, Android, Premium active)  
- Ahmed Rahman (30, India, iOS, Basic active)

### Subscriptions (3 records)
- Premium trial subscription
- Premium active subscription
- Basic active subscription

### Devices (2 records)
- iPhone 13 (iOS 16.5.0, app v5.2.1)
- Samsung Galaxy S21 (Android 13.0, app v11.0.0)

### Content & Watch History
- Bengali Drama movie
- Tech Talks series
- Associated watch history records

## 🐛 Troubleshooting

### Dgraph Connection Issues
```bash
# Check Dgraph logs
docker-compose logs dgraph-alpha

# Restart services
docker-compose down && docker-compose up -d
```

### Schema/Data Issues
```bash
# Reset and reload data
docker-compose down -v  # Remove volumes
docker-compose up -d    # Restart with fresh data
```

### API Server Issues
```bash
# Check if Dgraph is running
curl -f http://localhost:8080/health || echo "Dgraph not running"

# Test API server without Dgraph
curl http://localhost:8080/api/v1/convert  # Should work without Dgraph
```

## 🎉 Success Indicators

✅ **Dgraph Running**: `docker-compose ps` shows all services "Up"  
✅ **Schema Loaded**: Ratel UI shows schema with chorki_* types  
✅ **Data Loaded**: Sample queries return customer data  
✅ **API Connected**: Server logs show "Connected to Dgraph"  
✅ **Execute Endpoint**: `/execute` returns real data from Dgraph

---

## 🚀 Next Steps

Once everything is running:
1. Try the sample queries above
2. Explore the data in Ratel UI
3. Create your own queries using the `/execute` endpoint
4. Add more sample data if needed

**You now have a complete JSON → DQL → Dgraph → Results pipeline!** 🎯