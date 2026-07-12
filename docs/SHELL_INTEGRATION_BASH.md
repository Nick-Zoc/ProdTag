# Bash Integration

Phase 4.4 adds Bash support on macOS and Linux. Installation is always an explicit action from **Integrations**.

## Persistent Installation

ProdTag uses `~/.bash_profile` on macOS and `~/.bashrc` on Linux. It builds/copies the helper and script to ProdTag app data, creates a one-time `.prodtag-backup`, and inserts one idempotent marker block:

```sh
# >>> ProdTag >>>
export PRODTAG_HELPER='.../prodtag-helper'
export PRODTAG_MATCHER_CACHE='.../matcher-cache.json'
export PRODTAG_CONFIG='.../config.json'
source '.../prodtag.bash'
# <<< ProdTag <<<
```

Uninstall removes only that block. Partial marker blocks are reported instead of overwritten.

## Current Session

Copy the exact installed-path command from Integrations. Session controls are:

```sh
export PRODTAG_BASH_ENABLED=0
export PRODTAG_BASH_ENABLED=1
```

## Debugging

```sh
export PRODTAG_BASH_DEBUG=1
echo "${TMPDIR:-/tmp}/prodtag-bash-debug.log"
tail -n 50 "${TMPDIR:-/tmp}/prodtag-bash-debug.log"
export PRODTAG_BASH_DEBUG=0
```

The newest 200 debug lines are retained. Normal mode redirects helper output and runs asynchronously.

## Behavior and Limitations

- ProdTag prepends its function to `PROMPT_COMMAND` and preserves the previous value.
- It does not replace existing traps.
- Command text comes from Bash history, so commands excluded from history cannot be classified precisely.
- Duration is omitted because reliable pre-execution timing would require intrusive DEBUG traps, especially on macOS Bash 3.2.
- Cache filtering is conservative; missing, stale, or invalid cache state invokes the Go helper.
- Open a new terminal after persistent install/uninstall, or use the current-session command.
