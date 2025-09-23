Write-Host "🚀 Starting Go server with hot reload..." -ForegroundColor Green
Write-Host "📁 Working directory: $PWD" -ForegroundColor Cyan
Write-Host "🔥 Air will automatically restart the server when files change" -ForegroundColor Yellow
Write-Host "📝 Edit any .go file to see the magic happen!" -ForegroundColor Magenta
Write-Host ""

# Create tmp directory if it doesn't exist
if (!(Test-Path "tmp")) {
    New-Item -ItemType Directory -Path "tmp"
}

# Start Air
air