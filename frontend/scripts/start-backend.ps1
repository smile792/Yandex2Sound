param(
  [int]$Port = 8080
)

$ErrorActionPreference = "Stop"
$portFile = Join-Path $PSScriptRoot "..\.backend-port"

if (Test-Path $portFile) {
  Remove-Item -LiteralPath $portFile -Force -ErrorAction SilentlyContinue
}

function Test-PortFree([int]$TargetPort) {
  $listener = $null
  try {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Any, $TargetPort)
    $listener.Start()
    return $true
  } catch {
    return $false
  } finally {
    if ($null -ne $listener) {
      $listener.Stop()
    }
  }
}

if (-not (Test-PortFree -TargetPort $Port)) {
  throw "Backend port $Port is busy. Stop the process using it or change Port in start-backend.ps1."
}
$port = $Port
$tmp = "$portFile.tmp"
Set-Content -LiteralPath $tmp -Value $port -NoNewline -Encoding utf8
Move-Item -LiteralPath $tmp -Destination $portFile -Force

$env:PORT = [string]$port
Set-Location (Join-Path $PSScriptRoot "..\..\backend")
go run ./cmd/main.go
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}

