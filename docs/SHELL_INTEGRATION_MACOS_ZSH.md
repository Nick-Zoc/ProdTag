# macOS zsh Integration MVP

Phase 4.2 adds a manual macOS zsh integration path. It does not edit `~/.zshrc`, install scripts globally, or run a background daemon.

## Build the Helper

From the ProdTag project root:

```sh
go build -tags prodtaghelper -o build/bin/prodtag-helper .
```

The helper binary is expected at:

```sh
build/bin/prodtag-helper
```

## Source the zsh Script

For the current terminal session:

```sh
export PRODTAG_HELPER="$PWD/build/bin/prodtag-helper"
source "$PWD/scripts/prodtag.zsh"
```

This registers zsh hooks for the current shell only. Automatic install/uninstall is intentionally deferred.

## Test the Helper Directly

```sh
build/bin/prodtag-helper emit \
  --event-type command_success \
  --command "ls -a" \
  --exit-code 0 \
  --cwd "$PWD" \
  --duration-ms 120
```

Expected result:

- If a matching enabled rule exists, the helper reports the rule and sound, then plays the matched processed sound when available.
- If no rule matches, the helper prints a no-match response and exits successfully.
- If playback is muted/disabled, the event is still handled and logged.

## Test Through zsh

After sourcing the script, run:

```sh
ls -a
```

The zsh hook captures the command, exit code, current directory, and duration, then calls `prodtag-helper emit` in the background.

## Disable for Current Session

```sh
export PRODTAG_ZSH_ENABLED=0
```

Re-enable:

```sh
export PRODTAG_ZSH_ENABLED=1
```

## Event Type Inference

The zsh script maps commands into MVP event types:

- Exit code `0`: `command_success`
- Non-zero exit code: `command_failure`
- `git commit ...`: `git_commit_success` or `git_commit_failure`
- `git push ...`: `git_push_success` or `git_push_failure`
- Common test commands: `test_success` or `test_failure`
- Common build commands: `build_success` or `build_failure`

Common test commands include `npm test`, `pnpm test`, `yarn test`, `pytest`, `go test`, `cargo test`, and `flutter test`.

Common build commands include `npm run build`, `pnpm build`, `yarn build`, `go build`, `cargo build`, and `flutter build`.

## Persistent Event Log

Handled events are written as JSONL under ProdTag app data logs:

```sh
~/Library/Application Support/ProdTag/logs/handled-events.jsonl
```

The Integrations page reads this log for recent handled events.

## Troubleshooting

- Run `build/bin/prodtag-helper doctor` to verify config paths, rule count, sound count, listening state, event engine state, muted state, and playback method.
- If no sound plays, check that ProdTag has at least one enabled rule matching the emitted event type.
- If the helper reports muted or playback disabled, check Settings inside ProdTag.
- If `prodtag-helper` is not found, rebuild it or set `PRODTAG_HELPER` to the absolute helper path.
- If the script seems inactive, confirm `PRODTAG_ZSH_ENABLED=1` and that `source "$PWD/scripts/prodtag.zsh"` ran in the current zsh session.
