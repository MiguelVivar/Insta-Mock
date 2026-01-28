#!/bin/bash

# Scaffold Go project with Professional Go Layout
PROJECT_NAME="insta-mock"

echo "Creating project structure for $PROJECT_NAME..."

# Create directory structure
mkdir -p cmd/imock
mkdir -p internal/server
mkdir -p internal/generator
mkdir -p internal/tui
mkdir -p pkg/logger
mkdir -p examples

# Initialize Go module
echo "Initializing Go module..."
go mod init github.com/MiguelVivar/insta-mock

# Create main.go files
echo "Creating main.go files..."
touch cmd/imock/main.go

# Add basic package declarations
cat > cmd/imock/main.go << 'EOF'
package main

func main() {
	// TODO: Implement
}
EOF

echo "✓ Project structure created successfully!"
echo ""
echo "Directory tree:"
echo "├── cmd/"
echo "│   └── imock/"
echo "│       └── main.go"
echo "├── internal/"
echo "│   ├── server/"
echo "│   ├── generator/"
echo "│   └── tui/"
echo "├── pkg/"
echo "│   └── logger/"
echo "├── examples/"
echo "└── go.mod"
