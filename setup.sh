#!/bin/bash

# SubSpace Setup Script
# Automates the initial setup process

set -e

echo "🚀 SubSpace Setup Script"
echo "========================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Found Go version: $GO_VERSION"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file with your LinkedIn credentials"
    echo ""
else
    echo "✅ .env file already exists"
    echo ""
fi

# Create data directory
echo "📁 Creating data directory..."
mkdir -p data/sessions
echo "✅ Data directory created"
echo ""

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"
echo ""

# Build the application
echo "🔨 Building SubSpace..."
go build -o subspace main.go
echo "✅ Build successful"
echo ""

# Display next steps
echo "✨ Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your LinkedIn credentials:"
echo "   nano .env"
echo ""
echo "2. Run SubSpace with:"
echo "   ./subspace -action=search -keywords=\"Software Engineer\" -max-results=5"
echo ""
echo "3. View all options:"
echo "   ./subspace -help"
echo ""
echo "⚠️  REMINDER: This tool is for EDUCATIONAL PURPOSES ONLY"
echo "    Using it may violate LinkedIn's Terms of Service"
echo ""
