# macOS zsh Integration

Phase 4.4 keeps the explicit macOS zsh install/uninstall flow and moves it onto the shared cross-platform runtime. ProdTag never edits `~/.zshrc` without the user clicking an install or uninstall action.

## Persistent Installation

Open **Integrations** and click **Install persistently**. In development, ProdTag:

1. Builds the helper into `~/Library/Application Support/ProdTag/bin/prodtag-helper`.
2. Copies the integration script to `~/Library/Application Support/ProdTag/integrations/prodtag.zsh`.
3. Backs up an existing `~/.zshrc` once as `~/.zshrc.prodtag-backup` before its first ProdTag modification.
4. Adds one idempotent block:

```sh
# >>> ProdTag >>>
export PRODTAG_HELPER='.../ProdTag/bin/prodtag-helper'
export PRODTAG_MATCHER_CACHE='.../ProdTag/matcher-cache.json'
export PRODTAG_CONFIG='.../ProdTag/config.json'
source '.../ProdTag/integrations/prodtag.zsh'
# <<< ProdTag <<<
```

The setup applies automatically to newly opened terminals. A partial marker block is reported for manual repair instead of being overwritten.

Packaged releases still need to bundle platform helper binaries and the zsh script. The current installer builds/copies them from the project checkout for development.

## Current Terminal

The desktop app cannot change an already-running terminal's environment. Use the copyable command shown in **Integrations**, or run the equivalent installed-path command:

```sh
export PRODTAG_HELPER="$HOME/Library/Application Support/ProdTag/bin/prodtag-helper"
export PRODTAG_MATCHER_CACHE="$HOME/Library/Application Support/ProdTag/matcher-cache.json"
export PRODTAG_CONFIG="$HOME/Library/Application Support/ProdTag/config.json"
source "$HOME/Library/Application Support/ProdTag/integrations/prodtag.zsh"
```

Disable ProdTag for the current terminal:

```sh
export PRODTAG_ZSH_ENABLED=0
```

Re-enable it with `export PRODTAG_ZSH_ENABLED=1`.

## Silent Operation and Debugging

Normal operation is silent. The zsh hook detaches helper calls with zsh job-control suppression and redirects helper output, so command prompts do not show background job IDs or completion messages. It preserves the completed command's exit code.

Enable debug logging for the current terminal:

```sh
export PRODTAG_ZSH_DEBUG=1
```

Print the exact debug path and inspect it:

```sh
echo "${TMPDIR:-/tmp}/prodtag-zsh-debug.log"
tail -n 50 "${TMPDIR:-/tmp}/prodtag-zsh-debug.log"
```

Disable debug mode:

```sh
export PRODTAG_ZSH_DEBUG=0
```

The debug file is created before matcher-cache filtering and records the command, exit code, event type, cache path/state, launch decision, helper path, and helper result/error. It retains the newest 200 lines. Override the path with `PRODTAG_ZSH_DEBUG_LOG`.

## Matcher Cache

ProdTag regenerates `matcher-cache.json` whenever config is saved, including rule create, edit, enable/disable, and delete operations. The cache contains only enabled event types and lightweight rule match hints; it excludes sound metadata.

The zsh script skips the helper when the cache proves there is no enabled rule for the inferred event type. Missing, invalid, or older-than-config caches fall back to the Go helper. The Go matcher remains the source of truth.

## Doctor

Use **Run doctor** in Integrations or:

```sh
build/bin/prodtag-helper doctor
```

Doctor checks platform playback support and alternatives, zsh/Bash/PowerShell availability, installed/executable helper state, integration scripts, profile markers, config and matcher-cache paths, cache validity, rule count, and runtime controls.

## Uninstall

Click **Uninstall** in Integrations. ProdTag removes only the text between its marker lines and preserves unrelated `.zshrc` content. Stable helper/script files remain so an already-running terminal is not broken; they can be replaced by a later reinstall. Open a new terminal after uninstalling.

## Manual Development Build

From the project root:

```sh
go build -o build/bin/prodtag-helper ./cmd/prodtag-helper
```

Test directly:

```sh
build/bin/prodtag-helper emit \
  --event-type command_success \
  --command "ls -a" \
  --exit-code 0 \
  --cwd "$PWD" \
  --duration-ms 120
```

## Troubleshooting

- No sound: confirm an enabled rule matches the inferred event and check Listening, Event engine, Muted, and Playback in Settings.
- No events: run Doctor, then copy the current-session enable command or open a new terminal after persistent installation.
- Partial `.zshrc` state: repair/remove the incomplete ProdTag marker block manually, then install again.
- Debug an event: set `PRODTAG_ZSH_DEBUG=1`, run a command, and inspect the debug log.
- Cache concern: save config or mutate a rule to regenerate it; invalid/missing cache safely falls back to the helper.
