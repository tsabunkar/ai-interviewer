#!/bin/bash

# Script to start the local development server

echo "Starting Arch2Lead Interviewer backend server..."

# Set environment variables if needed
export PORT=8080

# Run the server
go run cmd/server/main.go