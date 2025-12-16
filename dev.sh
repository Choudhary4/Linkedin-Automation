#!/bin/bash

# SubSpace Development Mode
# Like "npm run dev" for Node.js projects

echo "🚀 SubSpace - Development Mode"
echo "================================"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  No .env file found!"
    echo "   Creating from .env.example..."
    cp .env.example .env
    echo "   ✅ Please edit .env with your credentials before running"
    echo ""
    exit 1
fi

# Run with development-friendly settings
echo "Starting with:"
echo "  - Debug logging enabled"
echo "  - Visible browser (non-headless)"
echo "  - Slow motion: 500ms"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Run with debug settings
LOG_LEVEL=debug \
HEADLESS=false \
SLOW_MOTION=500 \
go run main.go "$@"
