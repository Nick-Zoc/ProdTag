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

Status: MVP complete; normalization/playlists remain for follow-up.

Goal: Let users import, preview, and manage sounds.

Tasks:

- [x] Add drag-and-drop sound import.
- [x] Store original files in app data.
- [ ] Use FFmpeg to normalize and convert imported audio to internal WAV.
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

## Phase 3 — Helper + Playback

Goal: Make sounds play without the UI needing to stay open.

Tasks:

- [ ] Create `prodtag-helper` Go process.
- [ ] Add IPC or CLI command for sending events to helper.
- [ ] Implement async playback from processed WAV files.
- [ ] Add listening/muted state handling.
- [ ] Add stop-current-audio action.
- [ ] Add helper status display in Dashboard.

Done when:

- [ ] Helper can play a sound from a command/event while the UI is closed.

## Phase 4 — Rules and Command Matching

Goal: Let users attach sounds to command outcomes.

Tasks:

- [ ] Create rules model: pattern, match type, condition, playlist, enabled.
- [ ] Add built-in presets for Git, tests, and builds.
- [ ] Add custom command rule creation.
- [ ] Generate lightweight matcher cache for shell integrations.
- [ ] Add rule test button in UI.

Done when:

- [ ] User can create a rule like `npm test` success/fail and trigger the correct sound.

## Phase 5 — Shell Integrations

Goal: Make ProdTag react to real terminal commands.

Tasks:

- [ ] Add zsh integration first.
- [ ] Add bash integration.
- [ ] Add PowerShell integration.
- [ ] Add install/uninstall buttons in Integrations page.
- [ ] Support current-shell install and global install.
- [ ] Add doctor checks for installed shell integrations.

Done when:

- [ ] Running a matching command in zsh, bash, or PowerShell can trigger success/fail sounds.

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
