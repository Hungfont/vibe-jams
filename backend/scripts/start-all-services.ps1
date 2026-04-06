param(
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

$services = @(
    @{ Name = "jam-service"; Path = "jam-service"; PortMode = "SERVER_PORT"; DefaultPort = 8080 },
    @{ Name = "playback-service"; Path = "playback-service"; PortMode = "SERVER_PORT"; DefaultPort = 8082 },
    @{ Name = "api-service"; Path = "api-service"; PortMode = "SERVER_PORT"; DefaultPort = 8084 },
    @{ Name = "auth-service"; Path = "auth-service"; PortMode = "SERVER_ADDR"; DefaultPort = 8081 },
    @{ Name = "catalog-service"; Path = "catalog-service"; PortMode = "SERVER_PORT"; DefaultPort = 8083 },
    @{ Name = "rt-gateway"; Path = "rt-gateway"; PortMode = "SERVER_PORT"; DefaultPort = 8090 }
)

function Test-PortAvailable {
    param([int]$Port)

    $listener = $null
    try {
        $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $Port)
        $listener.Start()
        return $true
    } catch {
        return $false
    } finally {
        if ($listener -ne $null) {
            $listener.Stop()
        }
    }
}

function Get-RandomAvailablePort {
    $listener = $null
    try {
        $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
        $listener.Start()
        return $listener.LocalEndpoint.Port
    } finally {
        if ($listener -ne $null) {
            $listener.Stop()
        }
    }
}

$wtCommand = Get-Command "wt.exe" -ErrorAction SilentlyContinue
if (-not $wtCommand) {
    throw "Windows Terminal (wt.exe) is required to open service tabs in one terminal window."
}

$resolvedServices = @()
foreach ($service in $services) {
    $serviceDir = Join-Path $repoRoot $service.Path

    if (-not (Test-Path -LiteralPath $serviceDir)) {
        Write-Warning "Skipping $($service.Name): directory not found ($serviceDir)"
        continue
    }

    $selectedPort = $service.DefaultPort
    $usedDefaultPort = $true
    if (-not (Test-PortAvailable -Port $selectedPort)) {
        $selectedPort = Get-RandomAvailablePort
        $usedDefaultPort = $false
    }

    $runCmd = ""
    if ($service.PortMode -eq "SERVER_ADDR") {
        $runCmd = "set ""SERVER_ADDR=:$selectedPort"" && go run ./cmd/server"
    } else {
        $runCmd = "set ""SERVER_PORT=$selectedPort"" && go run ./cmd/server"
    }

    $resolvedServices += @{
        Name           = $service.Name
        Dir            = $serviceDir
        Port           = $selectedPort
        UsedDefault    = $usedDefaultPort
        CommandLine    = $runCmd
        PortMode       = $service.PortMode
        DefaultPort    = $service.DefaultPort
    }
}

if ($resolvedServices.Count -eq 0) {
    throw "No valid backend service directories were found."
}

$wtArgs = @("-w", "0")
for ($i = 0; $i -lt $resolvedServices.Count; $i++) {
    $svc = $resolvedServices[$i]

    if ($i -gt 0) {
        $wtArgs += ";"
    }

    $wtArgs += @(
        "new-tab",
        "--title", $svc.Name,
        "-p", "Command Prompt",
        "-d", $svc.Dir,
        "cmd", "/k", $svc.CommandLine
    )
}

if ($DryRun) {
    for ($i = 0; $i -lt $resolvedServices.Count; $i++) {
        $svc = $resolvedServices[$i]
        $portSource = "default"
        if (-not $svc.UsedDefault) {
            $portSource = "random"
        }
        Write-Host "[DRY-RUN] $($svc.Name): port=$($svc.Port) ($portSource, expected=$($svc.DefaultPort), mode=$($svc.PortMode))"
        Write-Host "[DRY-RUN] $($svc.Name): cmd /k \"$($svc.CommandLine)\" (cwd=$($svc.Dir))"
    }
    Write-Host "[DRY-RUN] wt arguments built successfully."
    Write-Host "Done."
    return
}

& $wtCommand.Source @wtArgs

Write-Host "Opened $($resolvedServices.Count) backend services as tabs in Windows Terminal using Command Prompt."
for ($i = 0; $i -lt $resolvedServices.Count; $i++) {
    $svc = $resolvedServices[$i]
    $portSource = "default"
    if (-not $svc.UsedDefault) {
        $portSource = "random"
    }
    Write-Host "- $($svc.Name): $($svc.Port) ($portSource)"
}
Write-Host "Done."
