#!/bin/bash

# Glimpse Z.AI Setup Script
# This script helps configure Glimpse to use Z.AI API

echo "🤖 Glimpse Z.AI Setup"
echo "======================="
echo

# Check if ZAI_API_KEY is set
if [ -z "$ZAI_API_KEY" ]; then
    echo "❌ ZAI_API_KEY environment variable not found"
    echo "Please get your API key from https://z.ai/manage-apikey/apikey-list"
    echo "Then run: export ZAI_API_KEY=\"your-api-key-here\""
    echo
    read -p "Would you like to enter your ZAI API key now? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter your ZAI API key: " -s api_key
        echo
        echo "export ZAI_API_KEY=\"$api_key\"" >> ~/.bashrc
        echo "✅ API key added to ~/.bashrc"
        echo "Please restart your terminal or run: source ~/.bashrc"
    else
        echo "❌ Setup cancelled"
        exit 1
    fi
else
    echo "✅ ZAI_API_KEY found"
fi

# Create .glimpse.yaml if it doesn't exist
if [ ! -f ".glimpse.yaml" ]; then
    echo
    echo "📝 Creating .glimpse.yaml with Z.AI configuration..."
    cp .glimpse.zai.yaml .glimpse.yaml
    echo "✅ Configuration file created"
else
    echo
    echo "⚠️  .glimpse.yaml already exists"
    echo "You can manually update it to use Z.AI by setting:"
    echo "  provider: zai"
    echo "  model: glm-4.6"
fi

# Test the setup
echo
echo "🧪 Testing configuration..."
if go build -o glimpse 2>/dev/null; then
    echo "✅ Glimpse builds successfully"
    echo
    echo "🎉 Setup complete! Run Glimpse with:"
    echo "  ./glimpse"
    echo
    echo "Available Z.AI models:"
    echo "  • glm-4.6 - High-performance model"
    echo "  • glm-4-air - Lightweight, efficient model"
    echo "  • glm-4-32b - Large context model"
else
    echo "❌ Build failed. Please check your Go installation"
    exit 1
fi