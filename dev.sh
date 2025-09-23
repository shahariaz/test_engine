#!/bin/bash
echo "🚀 Starting Go server with hot reload..."
echo "📁 Working directory: $(pwd)"
echo "🔥 Air will automatically restart the server when files change"
echo "📝 Edit any .go file to see the magic happen!"
echo ""

# Create tmp directory if it doesn't exist
mkdir -p tmp

# Start Air
air