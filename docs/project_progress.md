# ProdTag Project Progress

## Current Context

ProdTag is a Wails v2 desktop app with a Go backend/helper direction and a React + Tailwind frontend. The agreed product shape is a local control center for developer terminal sound tags, with shell integrations and a small background helper planned later.

## Completed

- Phase 1: Skeleton app foundation.
  - Wails app runs.
  - React + Tailwind frontend is wired.
  - Dashboard, Sounds, Rules, Hotkeys, Integrations, and Settings pages exist.
  - Go config service loads and saves local JSON config.
  - App data folders are created for config, matcher cache, sounds, processed sounds, and logs.
- Phase 1.1: Cleanup and UI foundation.
  - Main UI was split into reusable components and page-level components.
  - Design tokens and consistent component classes were added.
  - Settings copy now clearly distinguishes Listening from Muted behavior.
  - Loading state now uses ProdTag-specific copy.
  - Path rows now wrap long paths and include disabled Copy/Open placeholders.
- Phase 1.2: Frontend tooling modernization and Phase 1 polish.
  - Vite, the React Vite plugin, Tailwind CSS, and TypeScript were modernized while the app is still small.
  - Tailwind moved from the v3 PostCSS/config-file setup to the v4 Vite plugin setup.
  - The old Tailwind/PostCSS config files and direct PostCSS/autoprefixer dev dependencies were removed.
  - TypeScript config now uses modern Vite-friendly bundler module resolution.
  - `.DS_Store` files are ignored.
  - Manual `wails build` passed after the Codex-side packaging warning: bindings, frontend compile, application compile, packaging, self-signing, and `build/bin/ProdTag.app` creation all completed successfully.
- Phase 2 MVP: Sound Library.
  - Import button opens a Wails file picker and accepts MP3, WAV, M4A, OGG, and FLAC files.
  - Sounds page supports Wails drag-and-drop import.
  - Imported files are copied into the app data originals folder; selected source files are not mutated.
  - Sound records now include id, name, original path, optional processed path/duration, created time, status, and optional error.
  - Sounds page shows imported sound cards with status, format, path, imported time, preview, rename, and delete actions.
  - Preview uses frontend audio playback from a backend-provided data URL, keeping helper/playback architecture for later phases.
  - Delete removes the config record and safely removes the copied app-data file when it is inside the sound library folder.
  - Import progress copy covers selecting, copying, reading metadata, added, and failed states.
- Phase 2.1: Sound Library QA fixes and UX polish.
  - Wails file-drop support is enabled in the app options and the drop zone is marked as a Wails drop target.
  - Drop zone now has visible active feedback during drag enter/over.
  - Delete now uses a custom confirmation dialog instead of native browser confirm.
  - Single and bulk delete both stop preview when needed, remove records, remove copied app-data files safely, and refresh the shared config snapshot.
  - Sound cards now support checkbox selection and a clean "Delete selected" action.
  - A lightweight reusable toast/status component shows import, rename, and delete success/failure.
  - Sound action buttons now have clearer Preview, Stop, Rename, and subtle danger Delete hierarchy.
- Phase 2.5: Sound processing, duration probing, normalization, and polish.
  - Drop import progress now advances through per-file import messages and clears non-error status messages automatically.
  - Preview/Stop controls now keep a stable width so imported metadata rows do not jump when playback state changes.
  - The backend detects `ffmpeg` and `ffprobe` availability without blocking import, preview, rename, or delete when tools are missing.
  - Imported sounds are opportunistically probed for duration with `ffprobe`; existing sounds can be probed from the row action menu.
  - Sounds can be normalized manually per row or with Normalize all when FFmpeg is available.
  - Normalized files are written as internal WAV files under the processed sounds app-data folder while originals are preserved.
  - Preview prefers the processed WAV when present and falls back to the original copy.
  - Delete now removes processed app-data files as well as original app-data copies when safe.
  - Secondary row actions moved into an ellipsis menu; Preview and Delete remain visible.
  - Validation passed with `go test ./...` and `cd frontend && npm run build`; Codex-side `wails build` again reached frontend compile and failed at app compile without useful detail, so user-local manual Wails build remains the packaging source of truth.
- Phase 2.6: UI feedback system, icons, and sound page polish.
  - Added `lucide-react` for subtle stroke icons across navigation, page headers, buttons, status rows, empty states, and sound actions.
  - Shared UI now includes icon-aware buttons, an accessible icon button, a spinner, and a reusable progress bar.
  - Multi-file import and Normalize all now show honest step-based progress instead of only toast copy.
  - Busy actions show loading states and disable repeated clicks while import, rename, probing, normalization, or delete confirmation is in flight.
  - The Sounds page now shows separate ffmpeg and ffprobe status rows with installed/missing state and truncated tool paths.
  - Sound cards were tightened into clearer title, status/duration, path, and action areas while preserving the existing warm local-control-center style.
  - One-click FFmpeg/dependency installation is intentionally deferred to a later setup/integrations phase.
  - Validation passed with `go test ./...` and `cd frontend && npm run build`; Codex-side `wails build` still stops at the sandbox app-compile step after successful bindings/frontend compile.
- Phase 3: Rules MVP.
  - Rule records now store id, name, enabled state, event type, sound id, optional match mode, optional command pattern, optional exit code, and created/updated timestamps.
  - Supported MVP event types cover command success/failure, Git commit/push success, test success/failure, and build success/failure.
  - Backend rule methods now support listing through config, create, update, delete, enable/disable, and test playback for a rule sound.
  - The Rules page now supports create, edit, enable/disable, delete confirmation, selected sound metadata, missing sound references, and test playback.
  - Dashboard now shows total rules and enabled rules.
  - Rules are definitions only in this phase; terminal capture, helper playback, shell hooks, background daemon behavior, real hotkeys, and one-click dependency installs remain deferred.
- Phase 3.1: Rule matcher engine and event simulator.
  - Added a backend terminal event model with event type, optional command, optional exit code, optional cwd, timestamp, and optional duration.
  - Added internal matcher logic for enabled rules, event type, command match modes, optional exit code constraints, missing sound handling, and processed-path preference.
  - Match priority is deterministic: exact/regex and command-specific rules outrank broad any rules, exit-code-specific rules get a small bump, and equal-priority ties keep first-created/config order.
  - Added Wails methods for evaluating events, simulating events, listing recent simulated events, and clearing recent simulated events.
  - Added a compact Rules page event simulator with matched rule/sound display, no-match state, matched-sound playback, and an in-memory recent event log.
  - Recent simulated events are intentionally non-persistent until the helper/shell integration phase defines real logging needs.
- Phase 4.0: Backend playback engine and local event handling path.
  - Added a backend playback service that prefers processed sound files, falls back to originals, and starts playback asynchronously.
  - macOS playback uses `afplay`; Windows/Linux playback methods are explicitly deferred but the service boundary is in place.
  - Added stop-current-playback support and playback status reporting.
  - Added `HandleTerminalEvent`, which reuses the matcher, starts backend playback when enabled, records recent handled events, and returns playback status/error details.
  - Added config fields for `eventEngineEnabled`, `playbackEnabled`, `stopPreviousSoundOnNewEvent`, and optional `localEventPort`; defaults are enabled/enabled/stop previous.
  - Rules simulator now has a full-path "Handle + play" action in addition to matcher-only simulation.
  - Integrations page now shows event engine status, backend playback support/method, stop playback, and recent handled events.
  - Local HTTP intake was deferred to Phase 4.1; the preferred route is local-only CLI/helper intake rather than remote/network exposure.

## Current UX Direction

- Keep the dark sidebar, warm off-white app background, soft white cards, rounded corners, subtle badges, and clean local-control-center feel.
- Keep beginner-friendly copy in the UI.
- Avoid building sound import, helper, playback, hotkeys, or shell integrations before their roadmap phases.

## Process Notes

- After each implementation prompt, update this file with completed work, decisions, and next-step notes.
- Also update `docs/ROADMAP.md` with checkboxes or completion notes when a phase/subphase changes.
- When a dev server is started for verification, stop it before finishing unless the user explicitly asks to leave it running.
- Final responses should include what changed, what to test, build/test results, and a brief retrospective with suggestions or risks.
- Put project markdown under `docs/` when it is documentation/progress context; use root only for repo-standard files like `README.md`.

## Next Up

- Phase 4.1 helper and local shell integration intake planning.
- Playlist/group assignment is still open from Phase 2 if it is needed before moving into helper playback.
- The next technical boundary is sending real terminal events into `HandleTerminalEvent` through a local helper/CLI/shell integration.
- Packaging note: the earlier Codex-side macOS `UTType` linker/package warning is considered resolved for now because the user manually ran `wails build` successfully and produced `build/bin/ProdTag.app`.
- Codex-side `wails build` may still fail at the final macOS app compile step in this sandbox even after frontend and Go tests pass; prefer user-local manual build as the packaging truth for now.
