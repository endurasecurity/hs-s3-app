#!/bin/bash

# After Action Report Management System - Start Script
# This script starts the AAR demo application

echo "═══════════════════════════════════════════════════════════════════════"
echo "  Starting AAR Management System"
echo "═══════════════════════════════════════════════════════════════════════"
echo ""

# Check if wkhtmltopdf is installed (required for vulnerability demo)
if ! command -v wkhtmltopdf &> /dev/null; then
    echo "⚠ WARNING: wkhtmltopdf is not installed"
    echo "  The PDF generation vulnerability demo will not work without it."
    echo ""
    echo "  To install:"
    echo "    Ubuntu/Debian: sudo apt-get install wkhtmltopdf"
    echo "    macOS: brew install wkhtmltopdf"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Load environment variables if .env file exists
if [ -f .env ]; then
    echo "✓ Loading environment variables from .env"
    export $(grep -v '^#' .env | xargs)
else
    echo "ℹ No .env file found - using defaults (S3 features disabled)"
    echo "  Copy .env.example to .env and configure S3 credentials to enable S3"
fi

echo ""
echo "Starting application..."
echo ""

# Run the application
if [ -f ./hs-s3-app ]; then
    # Use compiled binary if it exists
    ./hs-s3-app
else
    # Otherwise fail
    echo "[error] hs-s3-app not found. run 'go build' first"
    exit 1
fi
