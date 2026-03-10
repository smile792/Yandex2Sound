$ErrorActionPreference = "Stop"
$portFile = Join-Path $PSScriptRoot "..\.backend-port"
$maxWaitMs = 30000
$stepMs = 200
$elapsed = 0
$port = ""

while ($elapsed -lt $maxWaitMs) {
  if (Test-Path $portFile) {
    try {
      $raw = Get-Content -Raw -LiteralPath $portFile -ErrorAction Stop
      $candidate = $raw.Trim()
      if ($candidate -match '^[0-9]{2,5}$') {
        $port = $candidate
        break
      }
    } catch {
      # file may be in transient state; retry
    }
  }

  Start-Sleep -Milliseconds $stepMs
  $elapsed += $stepMs
}

if (-not $port) {
  throw "Backend port file was not ready within $maxWaitMs ms"
}

$env:VITE_API_URL = "http://localhost:$port"
vite
