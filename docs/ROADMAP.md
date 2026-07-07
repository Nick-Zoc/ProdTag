# ProdTag Phased Roadmap

This roadmap is intentionally short. The goal is to build a usable v1 quickly, not accidentally start a six-month startup.

## Phase 1 — Skeleton App

Status: Complete.

Goal: Get the app opening and the basic project structure ready.

Tasks:

- [x] Create Wails v2 app with React + Tailwind.
- [x] Add base layout and navigation.
- [x] Create pages: Dashboard, Sounds, Rules, Hotkeys, Integrations, Settings.
- [x] Add Go services for config read/write.
- [x] Define app data directory structure.
- [x] Add initial `config.json` creation.

Done when:

- [x] The desktop app opens.
- [x] Navigation works.
- [x] Config can be loaded and saved.

Notes:

- Phase 1.1 added a component/page split, basic design tokens, clearer Settings copy, loading state, and path row polish.
- Phase 1.2 modernized frontend tooling to Vite 8, Tailwind 4 through `@tailwindcss/vite`, React Vite plugin 5, and TypeScript 6.
- Phase 1.2 removed obsolete Tailwind 3/PostCSS config files after the Tailwind 4 build passed.
- Manual `wails build` passed after Phase 1.2 and produced `build/bin/ProdTag.app`; the earlier Codex-side macOS packaging warning is considered resolved for now.
- Dev servers should be stopped after Codex verification so the user can run them manually.

## Phase 2 — Sound Library

Status: Phase 2.6 complete except playlists/groups.

Goal: Let users import, preview, and manage sounds.

Tasks:

- [x] Add drag-and-drop sound import.
- [x] Store original files in app data.
- [x] Use FFmpeg to normalize and convert imported audio to internal WAV.
- [x] Show import progress steps in the UI.
- [x] Add sound list with name, duration, format, and preview button.
- [x] Add delete/rename sound.
- [ ] Add playlists/groups.

Done when:

- [ ] User can import a sound, see progress, preview it, and assign it to a group.

Notes:

- Phase 2 MVP intentionally keeps `processedPath` and `durationMs` empty until FFmpeg probing/normalization.
- Preview currently uses UI-side audio from a backend data URL, not the future background helper.
- Playlist/group assignment remains open.
- Phase 2.1 enabled Wails file drop, added active drop feedback, custom delete confirmation, single/bulk delete, and toast/status feedback.
- Phase 2.5 added FFmpeg/ffprobe detection, duration probing, manual per-sound normalization, Normalize all, processed WAV output, processed-file preview preference, and processed-file cleanup on delete.
- Phase 2.5 kept normalization manual so missing FFmpeg does not block import, preview, rename, or delete.
- Phase 2.5 moved secondary row actions behind an ellipsis menu to keep sound cards calm as actions grow.
- Phase 2.6 added Lucide icons, icon-aware buttons, accessible icon buttons, reusable spinner/progress components, clearer audio-tool status rows, and more stable sound-card action states.
- Phase 2.6 intentionally deferred one-click FFmpeg/dependency installation to a later setup/integrations phase.

## Phase 3 — Rules MVP

Status: Complete.

Goal: Let users attach sounds to command outcomes.

Tasks:

- [x] Create rules model with name, enabled state, event type, sound assignment, optional command matching fields, and timestamps.
- [x] Add MVP event types for command, Git, test, and build success/failure outcomes.
- [x] Add rule create, edit, enable/disable, delete, and test-sound UI.
- [x] Show missing sound state when a referenced sound is deleted.
- [x] Add backend rule CRUD methods and config persistence tests.
- [x] Add internal rule matcher and in-app event simulator.
- [ ] Generate lightweight matcher cache for shell integrations.

Done when:

- [x] User can create a rule like `npm test` success/fail and test the assigned sound in the UI.

Notes:

- Rules are definitions only in Phase 3; real terminal detection, helper playback, shell hooks, and daemon behavior are intentionally deferred.
- Playlist/group assignment remains open from Phase 2, so Phase 3 rules assign a single sound.
- Phase 3.1 matcher priority is deterministic: enabled rules matching the event type are considered, exact/regex and command-specific rules outrank broad any rules, exit-code-specific rules get a small priority bump, and equal priority keeps first-created/config order.
- Phase 3.1 recent simulated events are in memory only; persistent logging belongs to the helper/integration phase.

## Phase 4 — Helper + Shell Integration

Status: Phase 4.2 macOS zsh helper MVP complete; automatic install and cross-platform shell support remain.

Goal: Make rules react to real terminal commands without requiring the UI to stay open.

Tasks:

- [x] Create `prodtag-helper` Go process.
- [x] Add CLI command for sending events to helper.
- [x] Implement async backend playback from matched local sound files on macOS through `afplay`.
- [x] Add local event handling API that evaluates events, starts backend playback, and records recent handled events.
- [x] Add listening/muted/playback config handling for the backend event path.
- [x] Add stop-current-audio backend/UI action.
- [x] Add shell-ready Dashboard readiness display.
- [x] Clean up Rules information architecture before shell integration.
- [x] Reorganize Integrations so shell setup is primary and diagnostics are secondary.
- [x] Clarify runtime settings for listening, rule engine, muted audio, and backend playback.
- [x] Add persistent handled-event JSONL log for helper/CLI events.
- [x] Add manual macOS zsh integration script.
- [x] Add manual macOS zsh setup documentation.
- [ ] Add helper status display in Dashboard.
- [x] Add zsh integration first.
- [ ] Add bash integration.
- [ ] Add PowerShell integration.
- [ ] Generate lightweight matcher cache for shell integrations.
- [ ] Add install/uninstall buttons in Integrations page.
- [ ] Add doctor checks for installed shell integrations.

Done when:

- [ ] Running a matching command in zsh, bash, or PowerShell can trigger the assigned rule sound through the helper.

Notes:

- Phase 4.0 adds the local app backend path only: `HandleTerminalEvent` evaluates a terminal event, selects the matched sound, starts backend playback when enabled, and records an in-memory recent event entry.
- Backend playback currently supports macOS via `afplay`; Windows/Linux playback methods are structured for later implementation.
- Phase 4.1 addressed the UX audit: Rules now keeps management primary with simulator/logs secondary, New/Edit opens in a modal, Integrations is setup-first, Settings groups runtime controls, Dashboard shows readiness, and paths are compacted.
- Phase 4.2 added a CLI/helper-first macOS zsh MVP. The helper is built manually with `go build -tags prodtaghelper -o build/bin/prodtag-helper .`, and `scripts/prodtag.zsh` is sourced manually for the current shell session.
- The zsh script captures command text, exit code, cwd, and duration, infers MVP event types, and calls `prodtag-helper emit` in the background.
- Local HTTP intake is intentionally not implemented. Automatic shell file edits, one-click install/uninstall, bash/PowerShell support, and daemon/tray behavior remain deferred.

## Phase 5 — Rule Presets and Matching Polish

Goal: Make command matching easier and safer after the helper path works.

Tasks:

- [ ] Add built-in presets for common Git, test, and build commands.
- [ ] Add custom command rule templates.
- [ ] Add rule test/matcher simulation with command and exit code input.
- [ ] Support current-shell install and global install if still needed.

Done when:

- [ ] User can quickly create and verify common command rules without understanding matcher internals.

## Phase 6 — Hotkeys and Startup

Goal: Make ProdTag feel like a real background utility.

Tasks:

- [ ] Add configurable global hotkeys.
- [ ] Add stop-current-audio hotkey.
- [ ] Add toggle-listening hotkey.
- [ ] Add mute/unmute hotkey.
- [ ] Add launch-helper-at-startup setting.
- [ ] Add helper start/stop/restart buttons.

Done when:

- [ ] User can control ProdTag without opening the UI.
- [ ] Helper can start automatically on login.

## Phase 7 — Packaging and Polish

Goal: Prepare v1 for friends/open-source release.

Tasks:

- [ ] Add app icons and basic branding.
- [ ] Add empty-state UI and helpful tooltips.
- [ ] Add error handling for missing FFmpeg/helper/shell issues.
- [ ] Add guided dependency setup or one-click install flow for external tools like FFmpeg.
- [ ] Add README with install/use instructions.
- [ ] Add GitHub release build workflow if possible.
- [ ] Build/test on macOS, Windows, and Linux.

Done when:

- [ ] A friend can install the app, import sounds, add a rule, and get a sound after a command succeeds or fails.

## Keep Out of v1

Do not build these unless the core app is already working:

- Sound marketplace
- Cloud accounts/sync
- Built-in AI audio generation
- Git GUI integration
- AI agent completion detection
- Advanced per-repo profiles
- Analytics dashboard
