# ProdTag UX Audit - Phase 4

Date: 2026-07-02

## Executive Summary

ProdTag's core functionality is now strong, but the UI has crossed from "local control center" into "developer diagnostics console" in a few places. The Sounds page still has the clearest product shape. Rules and Integrations now expose too many concepts at once: rules, matcher simulation, backend event handling, recent logs, shell setup placeholders, playback status, and config state. A beginner can technically use the app, but they may not know which area matters next or whether an action changed anything.

The next cleanup phase should focus on information architecture and interaction containment, not new capability. The biggest product move is to separate "make rules" from "debug the engine." Creation/editing should be focused and visible. Diagnostics should be collapsible, tabbed, or moved lower behind an advanced section.

Computer-use GUI inspection was successfully used on the running Wails app. No code behavior changes were made. I did not create/delete a temporary rule because the audit is report-only and the New rule form appeared outside the visible viewport after clicking.

## Biggest UX Issues

1. The Rules page mixes primary workflow and diagnostics in one long page.
2. New rule creation appears off-screen or below other sections, so clicking New rule may look like it did nothing.
3. Integrations leads with engine diagnostics while shell setup, the user's likely goal on that page, is pushed below recent events.
4. Full filesystem paths are visually noisy and make cards feel taller and more technical than necessary.
5. Several pages show implementation state rather than user intent: "event engine," "backend playback," "local intake," and "Handle + play" are accurate but not beginner-friendly.

## Beginner-Friendliness Problems

- The difference between Listening, Event engine, Muted, and Backend playback requires prior knowledge.
- "Simulate" versus "Handle + play" is not self-explanatory. A beginner may not know which one is safer or more realistic.
- Hotkeys shows "Not set" rows without explaining that hotkeys are not implemented yet.
- Integrations says shell integrations are not installed, but does not show what setup will eventually involve or what is currently usable.
- Dashboard shows counts but does not answer "Is ProdTag ready to react to commands?"

## Page-by-Page Observations

### Dashboard

What works:

- The page is clean and visually calm.
- Counts for Sounds, Rules, Enabled rules, and Shells are useful.
- The top state cards are easy to scan.

Issues:

- It does not show Event engine or Backend playback status, even though those are now key to the app working.
- Helper says "Not started," but Phase 4.0 currently has backend playback inside the app. This can feel contradictory.
- Local app data paths dominate the lower first viewport and feel more like diagnostics than dashboard content.

Recommendations:

- Add one compact "Readiness" band: Sounds, Enabled rules, Event engine, Playback, Shell setup.
- Move app data paths into a collapsible "Local files" diagnostics section.
- Reword Helper to "Background helper: Not built yet" or hide it until Phase 4.1/4.2.

### Sounds

What works:

- This is the strongest page. It has a clear job: import/manage sounds.
- Import, Normalize all, drag/drop, cards, badges, preview, and ellipsis actions feel coherent.
- The ellipsis menu works well for Rename, Probe duration, Normalize.

Issues:

- Audio tools block takes prime space even after tools are installed.
- Full original and processed paths make sound cards tall and visually technical.
- Delete is very prominent as a red button on every card.
- Select all reveals Delete selected clearly, but the destructive action becomes the loudest thing in the section.

Recommendations:

- Collapse Audio tools into a small status row once both tools are installed.
- Show paths as compact chips: "Original copy" and "Processed WAV" with reveal/copy actions.
- Keep Preview visible; move Delete into the ellipsis menu or make it less visually dominant.
- Consider a table/list density toggle later if users import many sounds.

### Rules

What works:

- Rule count and available sounds summary is useful.
- Simulator fields are understandable individually.
- Recent event output confirms the matcher/backend path exists.
- Rule cards show missing/matched sound state and actions.

Issues:

- Rules, Simulator, Recent simulated events, Create/Edit form, and rule list compete on one page.
- New rule appears below the visible area after clicking; there is no scroll, focus, or visible state change in the first viewport.
- The existing rule list is below simulator and recent events, even though rules are the primary object.
- Recent events can become long and bury actual rules.
- "Handle + play" is accurate but unclear for beginners.
- Duplicate no-match feedback appears as both toast and in-card state.

Recommendations:

- Split into tabs or segmented control:
  - Rules
  - Simulator
  - Recent events
  - Advanced diagnostics
- Move create/edit into a modal or side drawer so it appears immediately and focuses the task.
- Keep the Rules list above Simulator by default.
- Rename "Handle + play" to "Test full event flow" with helper text.
- Make recent events capped/collapsible on the Rules page.

### Integrations

What works:

- Event engine and playback status are clear for technical users.
- Backend playback method is visible.
- Recent handled events show that full-path events are reaching the backend.

Issues:

- Shell integrations are the likely reason a user opens this page, but they are below engine status and recent events.
- "Local intake deferred to Phase 4.1" is useful for development but not user-facing language.
- Recent handled events can bury shell setup.
- Engine status, playback, shell setup, and diagnostics are all mixed together.

Recommendations:

- Reorganize into:
  - Setup status
  - Shell integrations
  - Event engine
  - Playback
  - Recent handled events
  - Advanced diagnostics
- Put zsh/bash/PowerShell cards near the top.
- Move recent handled events into a collapsible diagnostics block.
- Replace "Deferred to Phase 4.1" with "Coming next: local shell event receiver."

### Settings

What works:

- Toggle cards are visually consistent and readable.
- The copy is short enough to scan.

Issues:

- Listening, Event engine, Muted, Backend playback, and Stop previous sound are too similar without grouping.
- There is no dependency explanation between states. For example, playback can be enabled while muted is on.
- Start helper at login appears beside working controls even though it is reserved/future.
- Config paths again take a lot of visual space.

Recommendations:

- Group toggles:
  - Event intake: Listening, Event engine
  - Audio behavior: Muted, Backend playback, Stop previous sound
  - Startup: Start helper at login
- Add small state hints for contradictory combinations: Muted on + Playback enabled.
- Move future-only toggles into "Coming later" or disable with clearer label.
- Collapse config paths.

### Hotkeys

What works:

- The page is simple and uncluttered.

Issues:

- "Not set" reads like missing configuration rather than not-yet-implemented behavior.
- No explanatory empty state or phase boundary.

Recommendations:

- Add an empty state: "Hotkeys are planned for Phase 6."
- Show expected future controls as disabled cards with purpose text.
- Add "Stop audio" as the most important future hotkey.

## Specific Interaction Issues Found

- New rule click did not produce visible feedback in the current viewport.
- Rules page has too much vertical content before the actual rule list.
- Simulator no-match feedback appears twice.
- Recent event lists can grow visually noisy.
- Full paths wrap aggressively and become the visual center of cards.
- Destructive actions are too prominent on Sounds and Rules.
- Stop playback is disabled but not explained when no sound is playing.

## Information Architecture Problems

- Primary workflows and diagnostics are mixed:
  - Rules creation is beside matcher simulation and recent logs.
  - Integrations setup is below event-engine diagnostics.
  - Dashboard diagnostic paths compete with readiness summary.
- Phase language leaks into UI in places where product language would help more.
- The app lacks a clear "ready/not ready" mental model after Phase 4.0.

## Recommended Layout Changes

- Dashboard:
  - Add "ProdTag readiness" summary.
  - Move paths into collapsible diagnostics.
- Rules:
  - Default tab: Rules list and New rule.
  - Simulator tab: event simulation and full event flow.
  - Recent tab: event logs.
  - Advanced tab: matcher details.
- Integrations:
  - Put shell setup first.
  - Keep engine/playback status second.
  - Put recent handled events last and collapsed by default.
- Settings:
  - Group controls by Event intake, Audio behavior, Startup, Diagnostics.

## Recommended Modal/Drawer/Collapsible Usage

- New/Edit rule should be a modal or right-side drawer.
- Rule simulator should be a separate tab or collapsible panel.
- Recent events should be collapsible on Rules and Integrations.
- Audio tool details and filesystem paths should be collapsible.
- Delete confirmations should remain modal.

## Recommended Component Improvements

- Add a compact `PathChip` component:
  - label, filename/folder, copy/reveal buttons, expandable full path.
- Add `SectionHeader` with optional description and right-side actions.
- Add `DiagnosticPanel` for collapsible technical details.
- Add `EmptyPhaseState` for pages/features planned later.
- Add action hierarchy:
  - Primary: Import, New rule, Test full event flow.
  - Secondary: Preview, Normalize all, Simulate.
  - Menu/overflow: Rename, Probe duration, Delete, Advanced diagnostics.

## Prioritized Fixes

### P0: Must Fix Before Shell Integration

- Make New/Edit rule appear immediately in a modal or drawer.
- Reorganize Rules so rules list is primary and simulator/recent events are secondary.
- Reorganize Integrations so shell setup/status appears before diagnostics.
- Add a clear Dashboard readiness summary for sounds, rules, event engine, playback, and shell setup.
- Clarify Listening vs Event engine vs Muted vs Backend playback.

### P1: Should Fix Soon

- Collapse recent event logs by default.
- Collapse or compact filesystem paths.
- Reduce visible destructive action prominence.
- Rename "Handle + play" to beginner-friendly language.
- Add Hotkeys planned-state empty copy.

### P2: Polish / Finalization

- Add density mode for sound/rule lists.
- Add tooltips for advanced statuses.
- Add persistent visual grouping across pages.
- Add improved copy for future-only controls.
- Add subtle success/error inline feedback beside the action that caused it.

## Proposed Next Phase Plan

Recommended phase name: **Phase 4.1 UX Cleanup and Shell-Ready IA**

Suggested scope:

1. Rules page information architecture cleanup.
2. New/Edit rule modal or drawer.
3. Integrations page setup-first layout.
4. Dashboard readiness summary.
5. Settings grouping and clearer state language.
6. Collapsible diagnostics for paths, tools, and recent events.
7. Hotkeys planned-state empty page.

Do not add shell hooks in this cleanup phase. The goal should be to make the current app understandable before adding real terminal integration.
