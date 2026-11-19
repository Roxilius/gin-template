<#
.SYNOPSIS
  Padanan Makefile untuk Windows PowerShell.

.EXAMPLE
  ./scripts/dev.ps1 help
  ./scripts/dev.ps1 run
  ./scripts/dev.ps1 migrate-create create_products_table
  ./scripts/dev.ps1 migrate-down 2
#>
param(
    [Parameter(Position = 0)][string]$Task = 'help',
    [Parameter(Position = 1)][string]$Arg = ''
)

$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

$SwagVersion = 'v1.16.6'

function Invoke-Step([string]$Command) {
    Write-Host "> $Command" -ForegroundColor DarkGray
    Invoke-Expression $Command
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

$tasks = [ordered]@{
    'run'             = @{ Desc = 'Jalankan API';                                 Run = { Invoke-Step 'go run ./cmd/api' } }
    'dev'             = @{ Desc = 'Jalankan API dengan hot reload (air)';         Run = {
            if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
                Write-Host "air belum terpasang. Jalankan: ./scripts/dev.ps1 tools`n(pastikan folder '$(go env GOPATH)\bin' ada di PATH)" -ForegroundColor Red
                exit 1
            }
            Invoke-Step 'air' } }
    'build'           = @{ Desc = 'Build binary ke ./bin';                        Run = {
            Invoke-Step 'go build -trimpath -ldflags="-s -w" -o bin/api.exe ./cmd/api'
            Invoke-Step 'go build -trimpath -ldflags="-s -w" -o bin/migrate.exe ./cmd/migrate' } }
    'tidy'            = @{ Desc = 'Rapikan go.mod/go.sum';                        Run = { Invoke-Step 'go mod tidy' } }
    'fmt'             = @{ Desc = 'Format kode';                                  Run = { Invoke-Step 'gofmt -w .' } }
    'test'            = @{ Desc = 'Jalankan semua test';                          Run = { Invoke-Step 'go test ./... -count=1' } }
    'cover'           = @{ Desc = 'Test + laporan coverage (coverage.html)';      Run = {
            Invoke-Step 'go test ./... -count=1 -coverprofile=coverage.out'
            Invoke-Step 'go tool cover -html=coverage.out -o coverage.html' } }
    'vet'             = @{ Desc = 'go vet';                                       Run = { Invoke-Step 'go vet ./...' } }
    'lint'            = @{ Desc = 'golangci-lint';                                Run = { Invoke-Step 'golangci-lint run ./...' } }
    'swag'            = @{ Desc = 'Generate dokumentasi Swagger ke ./docs';       Run = {
            Invoke-Step 'swag init -g cmd/api/main.go -o docs --parseInternal'
            Invoke-Step 'swag fmt' } }
    'migrate-up'      = @{ Desc = 'Jalankan semua migration';                     Run = { Invoke-Step 'go run ./cmd/migrate up' } }
    'migrate-down'    = @{ Desc = 'Rollback migration (arg: N, default 1)';       Run = {
            $n = if ($Arg) { $Arg } else { '1' }
            Invoke-Step "go run ./cmd/migrate down $n" } }
    'migrate-version' = @{ Desc = 'Versi migration saat ini';                     Run = { Invoke-Step 'go run ./cmd/migrate version' } }
    'migrate-create'  = @{ Desc = 'Buat file migration baru (arg: nama)';         Run = {
            if (-not $Arg) { throw 'Contoh: ./scripts/dev.ps1 migrate-create create_products_table' }
            Invoke-Step "go run ./cmd/migrate create $Arg" } }
    'db-up'           = @{ Desc = 'Jalankan PostgreSQL saja (development lokal)'; Run = { Invoke-Step 'docker compose up -d db' } }
    'docker-up'       = @{ Desc = 'Jalankan seluruh stack (db + migrate + api)';  Run = { Invoke-Step 'docker compose up -d --build' } }
    'docker-down'     = @{ Desc = 'Hentikan stack';                               Run = { Invoke-Step 'docker compose down' } }
    'docker-logs'     = @{ Desc = 'Lihat log api';                                Run = { Invoke-Step 'docker compose logs -f api' } }
    'tools'           = @{ Desc = 'Install air, swag, golangci-lint';             Run = {
            Invoke-Step 'go install github.com/air-verse/air@latest'
            Invoke-Step "go install github.com/swaggo/swag/cmd/swag@$SwagVersion"
            Invoke-Step 'go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest' } }
}

if ($Task -eq 'help' -or -not $tasks.Contains($Task)) {
    if ($Task -ne 'help') { Write-Host "Task tidak dikenal: $Task`n" -ForegroundColor Red }
    Write-Host 'Pemakaian: ./scripts/dev.ps1 <task> [arg]' -ForegroundColor Cyan
    foreach ($k in $tasks.Keys) { '  {0,-16} {1}' -f $k, $tasks[$k].Desc }
    if ($Task -ne 'help') { exit 1 }
    exit 0
}

& $tasks[$Task].Run
