# Scaffold Go project with Professional Go Layout
$PROJECT_NAME = "insta-mock"

Write-Host "Creating project structure for $PROJECT_NAME..." -ForegroundColor Cyan

# Create directory structure
New-Item -ItemType Directory -Force -Path "cmd\imock" | Out-Null
New-Item -ItemType Directory -Force -Path "internal\server" | Out-Null
New-Item -ItemType Directory -Force -Path "internal\generator" | Out-Null
New-Item -ItemType Directory -Force -Path "internal\tui" | Out-Null
New-Item -ItemType Directory -Force -Path "pkg\logger" | Out-Null
New-Item -ItemType Directory -Force -Path "examples" | Out-Null

# Initialize Go module
Write-Host "Initializing Go module..." -ForegroundColor Cyan
go mod init github.com/MiguelVivar/insta-mock

# Create main.go file with basic package declaration
Write-Host "Creating main.go files..." -ForegroundColor Cyan
@"
package main

func main() {
	// TODO: Implement
}
"@ | Out-File -FilePath "cmd\imock\main.go" -Encoding utf8

Write-Host ""
Write-Host "✓ Project structure created successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Directory tree:"
Write-Host "├── cmd/"
Write-Host "│   └── imock/"
Write-Host "│       └── main.go"
Write-Host "├── internal/"
Write-Host "│   ├── server/"
Write-Host "│   ├── generator/"
Write-Host "│   └── tui/"
Write-Host "├── pkg/"
Write-Host "│   └── logger/"
Write-Host "├── examples/"
Write-Host "└── go.mod"
