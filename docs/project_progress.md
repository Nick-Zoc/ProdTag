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

- Phase 2.5: Normalization and richer metadata.
- Add FFmpeg probing/normalization, duration extraction, processed WAV creation, and stronger import progress.
- Packaging note: the earlier Codex-side macOS `UTType` linker/package warning is considered resolved for now because the user manually ran `wails build` successfully and produced `build/bin/ProdTag.app`.
- Codex-side `wails build` may still fail at the final macOS app compile step in this sandbox even after frontend and Go tests pass; prefer user-local manual build as the packaging truth for now.
