param(
    [ValidateSet('start', 'stop', 'status', 'reset')]
    [string]$Action = 'start',
    [int]$Port = 55432,
    [string]$DataDir = (Join-Path $env:LOCALAPPDATA 'CuraClinic\pgdata'),
    [switch]$Force
)

$ErrorActionPreference = 'Stop'

function Resolve-PgBin {
    if ($env:PGBIN -and (Test-Path (Join-Path $env:PGBIN 'pg_ctl.exe'))) {
        return $env:PGBIN
    }
    $fromPath = Get-Command pg_ctl.exe -ErrorAction SilentlyContinue
    if ($fromPath) {
        return (Split-Path $fromPath.Source -Parent)
    }
    $candidate = 'C:\Program Files\PostgreSQL\18\bin'
    if (Test-Path (Join-Path $candidate 'pg_ctl.exe')) {
        return $candidate
    }
    throw 'PostgreSQL binaries not found. Set PGBIN to the folder containing pg_ctl.exe.'
}

$pgBin = Resolve-PgBin
$pgCtl = Join-Path $pgBin 'pg_ctl.exe'
$initDb = Join-Path $pgBin 'initdb.exe'
$psql = Join-Path $pgBin 'psql.exe'
$createdb = Join-Path $pgBin 'createdb.exe'
$logFile = Join-Path $DataDir 'server.log'

function Test-ClusterRunning {
    if (-not (Test-Path (Join-Path $DataDir 'PG_VERSION'))) {
        return $false
    }
    & $pgCtl -D $DataDir status *> $null
    return $LASTEXITCODE -eq 0
}

function Start-Cluster {
    if (-not (Test-Path (Join-Path $DataDir 'PG_VERSION'))) {
        New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
        Write-Host "Initializing PostgreSQL cluster in $DataDir"
        & $initDb -D $DataDir -U postgres -A trust --encoding=UTF8 --locale=C | Out-Host
        if ($LASTEXITCODE -ne 0) { throw 'initdb failed' }
    }

    if (-not (Test-ClusterRunning)) {
        Write-Host "Starting PostgreSQL on 127.0.0.1:$Port"
        $startLog = Join-Path $DataDir 'pgctl-start.log'
        & $pgCtl -D $DataDir -o "-p $Port -c listen_addresses=127.0.0.1" -l $logFile start *> $startLog
        if ($LASTEXITCODE -ne 0) { throw "pg_ctl start failed; see $startLog" }
        Start-Sleep -Seconds 2
    }

    $exists = & $psql -h 127.0.0.1 -p $Port -U postgres -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = 'curaclinic'"
    if ($exists.Trim() -ne '1') {
        Write-Host 'Creating database curaclinic'
        & $createdb -h 127.0.0.1 -p $Port -U postgres curaclinic
        if ($LASTEXITCODE -ne 0) { throw 'createdb failed' }
    }

    Write-Host ''
    Write-Host 'PostgreSQL is ready.'
    Write-Host "DATABASE_URL=postgres://postgres@127.0.0.1:$Port/curaclinic?sslmode=disable"
    Write-Host "Data directory: $DataDir"
    Write-Host "Log: $logFile"
}

function Stop-Cluster {
    if (Test-ClusterRunning) {
        & $pgCtl -D $DataDir stop -m fast
    } else {
        Write-Host 'Cluster is not running.'
    }
}

switch ($Action) {
    'start' { Start-Cluster }
    'stop' { Stop-Cluster }
    'status' {
        if (Test-ClusterRunning) {
            & $pgCtl -D $DataDir status
            Write-Host "DATABASE_URL=postgres://postgres@127.0.0.1:$Port/curaclinic?sslmode=disable"
        } else {
            Write-Host 'Cluster is not running.'
        }
    }
    'reset' {
        if (-not $Force) { throw 'reset removes all local data. Re-run with -Force to confirm.' }
        Stop-Cluster
        if (Test-Path $DataDir) { Remove-Item -Recurse -Force $DataDir }
        Write-Host 'Local cluster removed. Run start to initialize a new one.'
    }
}
