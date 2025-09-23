#!/bin/bash

echo "🚀 Loading large dataset into Dgraph..."

# Wait for Dgraph to be ready
echo "⏰ Waiting for Dgraph Alpha to be ready..."
sleep 30

# First, apply the schema
echo "📋 Setting up schema..."
curl -X POST localhost:8080/admin/schema \
  -H "Content-Type: text/plain" \
  --data-binary @dgraph/schema.graphql

if [ $? -eq 0 ]; then
    echo "✅ Schema applied successfully!"
else
    echo "❌ Failed to apply schema"
    exit 1
fi

# Then load the large dataset
echo "📊 Loading large dataset (77,524 records)..."
curl -X POST localhost:8080/mutate?commitNow=true \
  -H "Content-Type: application/rdf" \
  --data-binary @dgraph/large_sample_data.rdf

if [ $? -eq 0 ]; then
    echo "✅ Large dataset loaded successfully!"
    echo "📈 Dataset contains:"
    echo "   - 10,000 customers"
    echo "   - 8,000 subscriptions" 
    echo "   - 19,849 devices"
    echo "   - 500 content items"
    echo "   - 39,175 watch histories"
    echo "   - Total: 77,524 records"
else
    echo "❌ Failed to load dataset"
    exit 1
fi

echo "🎉 Dgraph setup complete with large dataset!"