[CmdletBinding()]
param(
    [switch]$WithInfra
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$rootDir = Split-Path -Parent $PSScriptRoot
$webDir = Join-Path $rootDir 'web'
$serverDir = Join-Path $rootDir 'server'

function Resolve-RequiredCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Names,

        [Parameter(Mandatory = $true)]
        [string]$DisplayName
    )

    foreach ($name in $Names) {
        $command = Get-Command $name -ErrorAction SilentlyContinue |
            Where-Object { $_.CommandType -in @('Application', 'ExternalScript') } |
            Select-Object -First 1

        if ($null -ne $command) {
            return $command
        }
    }

    throw "未找到 $DisplayName，请先安装并将其加入 PATH。"
}

function Resolve-WebDevCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$WorkingDirectory
    )

    $command = Join-Path $WorkingDirectory 'node_modules\.bin\vpr.cmd'
    if (Test-Path -LiteralPath $command -PathType Leaf) {
        return $command
    }

    throw "未找到前端依赖：$command。请先在 web 目录执行 pnpm install。"
}

function Stop-ProcessTree {
    param(
        [Parameter(Mandatory = $true)]
        [System.Diagnostics.Process]$Process
    )

    $Process.Refresh()
    if ($Process.HasExited) {
        return
    }

    try {
        & taskkill.exe /PID $Process.Id /T /F *> $null
        if ($LASTEXITCODE -ne 0) {
            Stop-Process -Id $Process.Id -Force -ErrorAction SilentlyContinue
        }
    }
    catch {
        Stop-Process -Id $Process.Id -Force -ErrorAction SilentlyContinue
    }
}

function Start-ManagedProcess {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,

        [Parameter(Mandatory = $true)]
        [string]$Arguments,

        [Parameter(Mandatory = $true)]
        [string]$WorkingDirectory,

        [hashtable]$Environment
    )

    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $FilePath
    $startInfo.Arguments = $Arguments
    $startInfo.WorkingDirectory = $WorkingDirectory
    $startInfo.UseShellExecute = $false
    $startInfo.CreateNoWindow = $false

    if ($null -ne $Environment) {
        foreach ($entry in $Environment.GetEnumerator()) {
            $startInfo.EnvironmentVariables[$entry.Key] = [string]$entry.Value
        }
    }

    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo

    if (-not $process.Start()) {
        throw "无法启动进程：$FilePath"
    }

    return $process
}

foreach ($directory in @($webDir, $serverDir)) {
    if (-not (Test-Path -LiteralPath $directory -PathType Container)) {
        throw "项目目录不存在：$directory"
    }
}

$goCommand = Resolve-RequiredCommand -Names @('go.exe', 'go') -DisplayName 'Go'
$cmdCommand = Resolve-RequiredCommand -Names @('cmd.exe', 'cmd') -DisplayName 'cmd.exe'
$webDevCommand = Resolve-WebDevCommand -WorkingDirectory $webDir
$webEnvironment = @{
    Path = "$(Split-Path -Parent $webDevCommand);$env:Path"
}

if ($WithInfra) {
    $dockerCommand = Resolve-RequiredCommand -Names @('docker.exe', 'docker') -DisplayName 'Docker'

    Push-Location $serverDir
    try {
        Write-Host '正在启动 PostgreSQL 和 Redis...' -ForegroundColor Cyan
        & $dockerCommand.Source compose up -d postgres redis
        if ($LASTEXITCODE -ne 0) {
            throw "Docker Compose 启动基础设施失败，退出码：$LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

$processes = @()
$exitCode = 0

try {
    Write-Host '正在启动后端：go run cmd/main.go' -ForegroundColor Green
    $serverProcess = Start-ManagedProcess `
        -FilePath $cmdCommand.Source `
        -Arguments '/d /s /c "go run cmd/main.go"' `
        -WorkingDirectory $serverDir
    $processes += $serverProcess

    Write-Host '正在启动前端：vite-plus run dev' -ForegroundColor Green
    $webProcess = Start-ManagedProcess `
        -FilePath $cmdCommand.Source `
        -Arguments ('/d /s /c ""{0}" dev"' -f $webDevCommand) `
        -WorkingDirectory $webDir `
        -Environment $webEnvironment
    $processes += $webProcess

    Write-Host ''
    Write-Host '开发服务已启动：' -ForegroundColor Cyan
    Write-Host '  前端：http://localhost:9527'
    Write-Host '  后端：http://localhost:8080/api/v1'
    Write-Host '按 Ctrl+C 停止前端和后端进程。' -ForegroundColor Yellow

    while ($true) {
        $finished = @()
        foreach ($process in $processes) {
            $process.Refresh()
            if ($process.HasExited) {
                $finished += $process
            }
        }

        if ($finished.Count -gt 0) {
            foreach ($process in $finished) {
                $process.WaitForExit()
                $process.Refresh()
                $processExitCode = $process.ExitCode
                if ($processExitCode -ne 0) {
                    if ($exitCode -eq 0) {
                        $exitCode = [int]$processExitCode
                    }
                    Write-Warning "进程 $($process.Id) 已退出，退出码：$processExitCode"
                }
            }
            break
        }

        Start-Sleep -Seconds 1
    }
}
finally {
    foreach ($process in $processes) {
        if ($null -ne $process -and -not $process.HasExited) {
            Write-Host "正在停止进程 $($process.Id)..." -ForegroundColor DarkGray
            Stop-ProcessTree -Process $process
        }
    }
}

exit $exitCode
