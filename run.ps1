[CmdletBinding()]
param(
    [ValidateSet('start', 'stop', 'restart', 'logs', 'status')]
    [string]$Action = 'start'
)

$ErrorActionPreference = 'Stop'
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $ProjectRoot

try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw 'Docker was not found. Install Docker Desktop, start it, then run .\run.bat again.'
    }

    & docker info *> $null
    if ($LASTEXITCODE -ne 0) {
        throw 'Docker Engine is not running. Start Docker Desktop and retry.'
    }

    if ($Action -eq 'stop') {
        & docker compose down
        if ($LASTEXITCODE -ne 0) { throw 'Could not stop the Kingdom Arena containers.' }
        Write-Host 'Kingdom Arena stopped. Database volumes were preserved.' -ForegroundColor Green
        return
    }

    if ($Action -eq 'logs') {
        & docker compose logs --follow --tail=100
        return
    }

    if ($Action -eq 'status') {
        & docker compose ps
        return
    }

    $requestedPort = $env:WEB_PORT
    if ([string]::IsNullOrWhiteSpace($requestedPort)) {
        $webPort = 8088
        $portFound = $false
        for ($candidate = 8088; $candidate -le 8098; $candidate++) {
            $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $candidate)
            $available = $false
            try {
                $listener.Start()
                $available = $true
            } catch {
                $available = $false
            } finally {
                $listener.Stop()
            }
            if ($available) {
                $webPort = $candidate
                $portFound = $true
                break
            }
        }
        if (-not $portFound) { throw 'Ports 8088-8098 are busy. Set WEB_PORT to an available port and retry.' }
    } else {
        $parsedPort = 0
        if (-not [int]::TryParse($requestedPort, [ref]$parsedPort) -or $parsedPort -lt 1024 -or $parsedPort -gt 65535) {
            throw 'WEB_PORT must be a number between 1024 and 65535.'
        }
        $webPort = $parsedPort
        $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $webPort)
        try {
            $listener.Start()
        } catch {
            throw "Port $webPort is already in use. Choose another WEB_PORT."
        } finally {
            $listener.Stop()
        }
    }

    $env:WEB_PORT = [string]$webPort

    if ($Action -eq 'restart') {
        & docker compose down
        if ($LASTEXITCODE -ne 0) { throw 'Could not restart the Kingdom Arena containers.' }
    }

    Write-Host 'Building and starting Kingdom Arena...' -ForegroundColor Cyan
    & docker compose up --build --detach --remove-orphans
    if ($LASTEXITCODE -ne 0) { throw 'Docker Compose build/start failed. Run .\run.bat logs for diagnostics.' }

    $url = "http://localhost:$webPort"
    $deadline = (Get-Date).AddMinutes(3)
    $webReady = $false
    $apiReady = $false

    while ((Get-Date) -lt $deadline -and (-not $webReady -or -not $apiReady)) {
        try {
            $response = Invoke-WebRequest -Uri "$url/" -UseBasicParsing -TimeoutSec 4
            $webReady = ($response.StatusCode -eq 200)
        } catch {
            $webReady = $false
        }
        try {
            $response = Invoke-WebRequest -Uri "$url/health" -UseBasicParsing -TimeoutSec 4
            $apiReady = ($response.StatusCode -eq 200)
        } catch {
            $apiReady = $false
        }
        if (-not $webReady -or -not $apiReady) { Start-Sleep -Seconds 2 }
    }

    if (-not $webReady -or -not $apiReady) {
        & docker compose ps
        & docker compose logs --tail=80
        throw 'The app did not become healthy within 3 minutes. Check the logs above.'
    }

    Write-Host "Kingdom Arena is ready: $url" -ForegroundColor Green
    Write-Host 'The browser will open now. Press Ctrl+C in the Docker logs window only if you want to stop viewing logs; use .\run.bat stop to stop the game.' -ForegroundColor DarkGray
    Start-Process $url
} catch {
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}
