# PowerShell Integration

Phase 4.4 targets PowerShell 7 and supports Windows PowerShell when it is discoverable. If neither executable exists, ProdTag reports a structured unavailable state.

## Persistent Installation

ProdTag asks PowerShell for `$PROFILE.CurrentUserAllHosts`, backs up an existing profile once, and inserts one marked block that sets the helper/cache/config paths and dot-sources `prodtag.ps1`. Installation and uninstall happen only after explicit user actions.

Uninstall removes only content between:

```powershell
# >>> ProdTag >>>
# <<< ProdTag <<<
```

## Current Session

Use the copyable command in Integrations. Session controls are:

```powershell
$env:PRODTAG_POWERSHELL_ENABLED = '0'
$env:PRODTAG_POWERSHELL_ENABLED = '1'
```

## Debugging

```powershell
$env:PRODTAG_POWERSHELL_DEBUG = '1'
Join-Path ([System.IO.Path]::GetTempPath()) 'prodtag-powershell-debug.log'
Get-Content (Join-Path ([System.IO.Path]::GetTempPath()) 'prodtag-powershell-debug.log') -Tail 50
$env:PRODTAG_POWERSHELL_DEBUG = '0'
```

The newest 200 lines are retained. Debug mode records command/event/cache/helper decisions and waits for the helper so its result can be logged; normal mode launches the helper asynchronously with a hidden window on Windows.

## Prompt Safety and Limitations

- The existing `prompt` function is saved and called by the ProdTag wrapper.
- The same history entry is emitted only once.
- Native exit codes and PowerShell success state are mapped to success/failure events.
- Duration is intentionally optional in this MVP.
- Missing/invalid/stale cache state falls back to the Go matcher.
- Profile/script behavior must still be runtime-tested on Windows PowerShell and PowerShell 7 machines.
