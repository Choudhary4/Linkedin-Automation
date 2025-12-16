#!/bin/bash

# SubSpace - LinkedIn Automation Tool
# Simple launcher script

echo "🚀 Starting SubSpace..."
echo ""

# Run the application with HEADLESS=false for visible browser
env HEADLESS=false go run main.go
