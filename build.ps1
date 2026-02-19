
# Build script for Windows (PowerShell)

Write-Host "Building Kakoclaw for Windows..." -ForegroundColor Green

# 1. Build Frontend
Write-Host "1. Building Frontend (Vue.js)..." -ForegroundColor Cyan
Push-Location "pkg/web/frontend"
try {
    npm install
    if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
    
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "npm run build failed" }
}
catch {
    Write-Error "Frontend build failed: $_"
    Pop-Location
    exit 1
}
Pop-Location

# 2. Build Backend
Write-Host "2. Building Backend (Go)..." -ForegroundColor Cyan
try {
    go build -o kakoclaw.exe ./cmd/kakoclaw
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
}
catch {
    Write-Error "Backend build failed: $_"
    exit 1
}

Write-Host "Build Complete! Binary is at .\kakoclaw.exe" -ForegroundColor Green
