# ProdTag PowerShell integration. Dot-source from the PowerShell profile or current session.

if (-not $env:PRODTAG_POWERSHELL_ENABLED) { $env:PRODTAG_POWERSHELL_ENABLED = '1' }
if (-not $env:PRODTAG_POWERSHELL_DEBUG) { $env:PRODTAG_POWERSHELL_DEBUG = '0' }
$script:ProdTagScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $env:PRODTAG_HELPER) { $env:PRODTAG_HELPER = Join-Path (Split-Path -Parent $script:ProdTagScriptDir) 'bin/prodtag-helper.exe' }
if (-not $env:PRODTAG_MATCHER_CACHE) { $env:PRODTAG_MATCHER_CACHE = Join-Path (Split-Path -Parent $script:ProdTagScriptDir) 'matcher-cache.json' }
if (-not $env:PRODTAG_CONFIG) { $env:PRODTAG_CONFIG = Join-Path (Split-Path -Parent $script:ProdTagScriptDir) 'config.json' }
if (-not $env:PRODTAG_POWERSHELL_DEBUG_LOG) { $env:PRODTAG_POWERSHELL_DEBUG_LOG = Join-Path ([System.IO.Path]::GetTempPath()) 'prodtag-powershell-debug.log' }

function Write-ProdTagDebug([string]$Message) {
    if ($env:PRODTAG_POWERSHELL_DEBUG -ne '1') { return }
    Add-Content -LiteralPath $env:PRODTAG_POWERSHELL_DEBUG_LOG -Value "$(Get-Date -AsUTC -Format o) $Message"
    $lines = Get-Content -LiteralPath $env:PRODTAG_POWERSHELL_DEBUG_LOG -ErrorAction SilentlyContinue
    if ($lines.Count -gt 200) { $lines | Select-Object -Last 200 | Set-Content -LiteralPath $env:PRODTAG_POWERSHELL_DEBUG_LOG }
}

function Get-ProdTagEventType([string]$Command, [int]$ExitCode) {
    $suffix = if ($ExitCode -eq 0) { 'success' } else { 'failure' }
    $value = $Command.Trim().ToLowerInvariant()
    if ($value.StartsWith('git commit')) { return "git_commit_$suffix" }
    if ($value.StartsWith('git push')) { return "git_push_$suffix" }
    if ($value -match '^(npm (run )?test|pnpm (run )?test|yarn test|pytest|go test|cargo test|flutter test)') { return "test_$suffix" }
    if ($value -match '^(npm run build|pnpm (run )?build|yarn build|go build|cargo build|flutter build)') { return "build_$suffix" }
    return "command_$suffix"
}

function Test-ProdTagCache([string]$EventType) {
    $script:ProdTagCacheState = 'missing'; $script:ProdTagCacheReason = 'cache missing; helper fallback'
    if (-not (Test-Path -LiteralPath $env:PRODTAG_MATCHER_CACHE)) { return $true }
    if ((Test-Path -LiteralPath $env:PRODTAG_CONFIG) -and ((Get-Item $env:PRODTAG_MATCHER_CACHE).LastWriteTimeUtc -lt (Get-Item $env:PRODTAG_CONFIG).LastWriteTimeUtc)) { $script:ProdTagCacheState='stale';$script:ProdTagCacheReason='cache older than config; helper fallback';return $true }
    try { $cache = Get-Content -Raw -LiteralPath $env:PRODTAG_MATCHER_CACHE | ConvertFrom-Json -ErrorAction Stop } catch { $script:ProdTagCacheState='invalid';$script:ProdTagCacheReason='cache validation failed; helper fallback';return $true }
    if ($cache.version -ne 2 -or -not $cache.complete) { $script:ProdTagCacheState='invalid';$script:ProdTagCacheReason='cache version incomplete; helper fallback';return $true }
    $script:ProdTagCacheState='valid'
    if ($cache.enabledEventTypes -contains $EventType) { $script:ProdTagCacheReason='enabled event type present; helper allowed';return $true }
    $script:ProdTagCacheReason='no enabled rule for event type; helper skipped';return $false
}

function Invoke-ProdTagEvent([int]$ExitCode) {
    if ($env:PRODTAG_POWERSHELL_ENABLED -eq '0') { return }
    $history = Get-History -Count 1 -ErrorAction SilentlyContinue
    if (-not $history) { return }
    if ($script:ProdTagLastHistoryId -eq $history.Id) { return }
    $script:ProdTagLastHistoryId = $history.Id
    $command = $history.CommandLine
    if ([string]::IsNullOrWhiteSpace($command) -or $command -match 'prodtag-helper|prodtag\.ps1|PRODTAG_POWERSHELL_') { return }
    $eventType = Get-ProdTagEventType $command $ExitCode
    Write-ProdTagDebug "captured command=$command exit_code=$ExitCode event_type=$eventType helper=$env:PRODTAG_HELPER cache=$env:PRODTAG_MATCHER_CACHE"
    if (-not (Test-ProdTagCache $eventType)) { Write-ProdTagDebug "cache_state=$script:ProdTagCacheState launch=skipped reason=$script:ProdTagCacheReason"; return }
    Write-ProdTagDebug "cache_state=$script:ProdTagCacheState launch=allowed reason=$script:ProdTagCacheReason"
    if (-not (Test-Path -LiteralPath $env:PRODTAG_HELPER)) { Write-ProdTagDebug "helper_result=error reason=helper missing path=$env:PRODTAG_HELPER"; return }
    $arguments = @('emit','--event-type',$eventType,'--command',$command,'--exit-code',"$ExitCode",'--cwd',(Get-Location).Path)
    if ($env:PRODTAG_POWERSHELL_DEBUG -eq '1') { $result = & $env:PRODTAG_HELPER @arguments 2>&1 | Out-String; Write-ProdTagDebug "helper_result_code=$LASTEXITCODE helper_result=$($result.Trim())" }
    else { try { Start-Process -FilePath $env:PRODTAG_HELPER -ArgumentList $arguments -WindowStyle Hidden | Out-Null } catch { Start-Process -FilePath $env:PRODTAG_HELPER -ArgumentList $arguments | Out-Null } }
}

if (-not $global:ProdTagOriginalPrompt) { $global:ProdTagOriginalPrompt = ${function:prompt} }
function global:prompt {
    $prodTagSuccess = $?
    $prodTagNativeExit = $global:LASTEXITCODE
    $prodTagExitCode = if ($prodTagSuccess) { 0 } elseif ($prodTagNativeExit -is [int] -and $prodTagNativeExit -ne 0) { $prodTagNativeExit } else { 1 }
    Invoke-ProdTagEvent -ExitCode $prodTagExitCode
    if ($global:ProdTagOriginalPrompt) { return & $global:ProdTagOriginalPrompt }
    return "PS $($executionContext.SessionState.Path.CurrentLocation)> "
}
