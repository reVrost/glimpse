#!/bin/bash

# Glimpse Setup Script
# This script helps set up Glimpse for first-time use

set -e

echo "🔍 Glimpse Setup"
echo "================"

# Check if .glimpse.yaml exists
if [ -f ".glimpse.yaml" ]; then
    echo "✅ .glimpse.yaml already exists"
    CONFIG_EXISTS=true
else
    echo "📝 Creating .glimpse.yaml from example..."
    if [ -f ".glimpse.yaml.example" ]; then
        cp .glimpse.yaml.example .glimpse.yaml
        echo "✅ Created .glimpse.yaml (customizable)"
    else
        echo "❌ .glimpse.yaml.example not found"
        exit 1
    fi
    CONFIG_EXISTS=false
fi

# Check API keys
echo ""
echo "🔑 API Key Setup"

if [ -n "$OPENAI_API_KEY" ]; then
    echo "✅ OPENAI_API_KEY is set"
elif grep -q "api_key:" .glimpse.yaml 2>/dev/null; then
    echo "✅ API key found in .glimpse.yaml"
else
    echo "⚠️  No API key found"
    echo "   Set environment variable:"
    echo "   export OPENAI_API_KEY=\"your-key-here\""
    echo "   Or add api_key to .glimpse.yaml"
fi

# Check for log directory
echo ""
echo "📋 Log Setup"

if [ -d "tmp" ]; then
    echo "✅ tmp/ directory exists"
else
    echo "📁 Creating tmp/ directory for logs..."
    mkdir -p tmp
    echo "✅ Created tmp/ directory"
fi

# Check if Glimpse binary exists
echo ""
echo "🏗️  Build Status"

if [ -f "./glimpse" ]; then
    echo "✅ Glimpse binary found"
    GLIMPSE_CMD="./glimpse"
elif command -v glimpse &> /dev/null; then
    echo "✅ Glimpse installed in PATH"
    GLIMPSE_CMD="glimpse"
else
    echo "🔨 Building Glimpse..."
    go build -o glimpse
    echo "✅ Glimpse built successfully"
    GLIMPSE_CMD="./glimpse"
fi

# Show next steps
echo ""
echo "🚀 Next Steps:"
echo ""

if [ "$CONFIG_EXISTS" = false ]; then
    echo "1. Customize .glimpse.yaml for your project"
    echo ""
fi

echo "2. Set your API key:"
echo "   export OPENAI_API_KEY=\"your-key-here\""
echo ""
echo "3. Start your application with logging:"
echo "   go run . | tee tmp/server.log"
echo ""
echo "4. In another terminal, start Glimpse:"
echo "   $GLIMPSE_CMD"
echo ""
echo "5. Make code changes and watch for reviews!"

# Clean up
echo ""
echo "✨ Setup complete!"