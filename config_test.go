package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureAppDataCreatesInitialFiles(t *testing.T) {
	isolateUserDirs(t)

	if err := EnsureAppData(); err != nil {
		t.Fatalf("EnsureAppData() error = %v", err)
	}

	paths, err := GetAppDataPaths()
	if err != nil {
		t.Fatalf("GetAppDataPaths() error = %v", err)
	}

	for _, path := range []string{
		paths.ConfigFile,
		paths.MatcherCacheFile,
		paths.OriginalSoundsDir,
		paths.ProcessedSoundsDir,
		paths.LogsDir,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestSaveConfigPersistsConfig(t *testing.T) {
	isolateUserDirs(t)

	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() error = %v", err)
	}

	snapshot.Config.Listening = false
	snapshot.Config.Muted = true

	app := NewApp()
	if _, err := app.SaveConfig(snapshot.Config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	reloaded, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() after save error = %v", err)
	}

	if reloaded.Config.Listening {
		t.Fatal("expected listening to persist as false")
	}
	if !reloaded.Config.Muted {
		t.Fatal("expected muted to persist as true")
	}
}

func TestImportSoundPathsCopiesFileAndPersistsRecord(t *testing.T) {
	isolateUserDirs(t)

	sourcePath := filepath.Join(t.TempDir(), "Producer Tag.MP3")
	if err := os.WriteFile(sourcePath, []byte("fake audio"), 0o644); err != nil {
		t.Fatalf("write source sound: %v", err)
	}

	snapshot, err := importSoundPaths([]string{sourcePath})
	if err != nil {
		t.Fatalf("importSoundPaths() error = %v", err)
	}

	if len(snapshot.Config.Sounds) != 1 {
		t.Fatalf("expected 1 sound, got %d", len(snapshot.Config.Sounds))
	}

	sound := snapshot.Config.Sounds[0]
	if sound.ID == "" {
		t.Fatal("expected sound id")
	}
	if sound.Name != "Producer Tag" {
		t.Fatalf("expected sound name Producer Tag, got %q", sound.Name)
	}
	if sound.Status != "imported" {
		t.Fatalf("expected imported status, got %q", sound.Status)
	}
	if sound.CreatedAt == "" {
		t.Fatal("expected createdAt")
	}
	if sound.ProcessedPath != nil {
		t.Fatal("expected processedPath to be nil before normalization")
	}
	if sound.DurationMs != nil {
		t.Fatal("expected durationMs to be nil before metadata probing")
	}
	if !strings.HasPrefix(sound.OriginalPath, snapshot.Paths.OriginalSoundsDir) {
		t.Fatalf("expected copied file in originals dir, got %s", sound.OriginalPath)
	}
	if _, err := os.Stat(sound.OriginalPath); err != nil {
		t.Fatalf("expected copied sound to exist: %v", err)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("expected original selected file to remain untouched: %v", err)
	}
}

func TestRenameAndDeleteSound(t *testing.T) {
	isolateUserDirs(t)

	sourcePath := filepath.Join(t.TempDir(), "tag.wav")
	if err := os.WriteFile(sourcePath, []byte("fake audio"), 0o644); err != nil {
		t.Fatalf("write source sound: %v", err)
	}

	snapshot, err := importSoundPaths([]string{sourcePath})
	if err != nil {
		t.Fatalf("importSoundPaths() error = %v", err)
	}

	app := NewApp()
	sound := snapshot.Config.Sounds[0]
	renamed, err := app.RenameSound(RenameSoundRequest{ID: sound.ID, Name: "Build Drop"})
	if err != nil {
		t.Fatalf("RenameSound() error = %v", err)
	}
	if renamed.Config.Sounds[0].Name != "Build Drop" {
		t.Fatalf("expected renamed sound, got %q", renamed.Config.Sounds[0].Name)
	}

	deleted, err := app.DeleteSound(sound.ID)
	if err != nil {
		t.Fatalf("DeleteSound() error = %v", err)
	}
	if len(deleted.Config.Sounds) != 0 {
		t.Fatalf("expected no sounds after delete, got %d", len(deleted.Config.Sounds))
	}
	if _, err := os.Stat(sound.OriginalPath); !os.IsNotExist(err) {
		t.Fatalf("expected copied sound file to be removed, stat err = %v", err)
	}
}

func TestDeleteSoundsRemovesMultipleRecordsAndFiles(t *testing.T) {
	isolateUserDirs(t)

	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.mp3")
	secondPath := filepath.Join(dir, "second.ogg")
	if err := os.WriteFile(firstPath, []byte("first"), 0o644); err != nil {
		t.Fatalf("write first sound: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("second"), 0o644); err != nil {
		t.Fatalf("write second sound: %v", err)
	}

	snapshot, err := importSoundPaths([]string{firstPath, secondPath})
	if err != nil {
		t.Fatalf("importSoundPaths() error = %v", err)
	}

	app := NewApp()
	first := snapshot.Config.Sounds[0]
	second := snapshot.Config.Sounds[1]
	deleted, err := app.DeleteSounds([]string{first.ID, second.ID})
	if err != nil {
		t.Fatalf("DeleteSounds() error = %v", err)
	}
	if len(deleted.Config.Sounds) != 0 {
		t.Fatalf("expected no sounds after bulk delete, got %d", len(deleted.Config.Sounds))
	}
	for _, path := range []string{first.OriginalPath, second.OriginalPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected copied sound file to be removed, stat err = %v", err)
		}
	}
}

func TestImportSoundPathsRejectsUnsupportedFile(t *testing.T) {
	isolateUserDirs(t)

	sourcePath := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(sourcePath, []byte("not audio"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if _, err := importSoundPaths([]string{sourcePath}); err == nil {
		t.Fatal("expected unsupported file error")
	}
}

func TestCheckAudioToolsDoesNotError(t *testing.T) {
	app := NewApp()
	if _, err := app.CheckAudioTools(); err != nil {
		t.Fatalf("CheckAudioTools() error = %v", err)
	}
}

func TestProcessSoundWithoutFFmpegMarksFailed(t *testing.T) {
	isolateUserDirs(t)

	sourcePath := filepath.Join(t.TempDir(), "tag.wav")
	if err := os.WriteFile(sourcePath, []byte("fake audio"), 0o644); err != nil {
		t.Fatalf("write source sound: %v", err)
	}

	snapshot, err := importSoundPaths([]string{sourcePath})
	if err != nil {
		t.Fatalf("importSoundPaths() error = %v", err)
	}

	originalLookPath := lookPath
	lookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}
	t.Cleanup(func() {
		lookPath = originalLookPath
	})

	app := NewApp()
	processed, err := app.ProcessSound(snapshot.Config.Sounds[0].ID)
	if err != nil {
		t.Fatalf("ProcessSound() error = %v", err)
	}

	sound := processed.Config.Sounds[0]
	if sound.Status != "failed" {
		t.Fatalf("expected failed status, got %q", sound.Status)
	}
	if sound.Error == nil || !strings.Contains(*sound.Error, "ffmpeg") {
		t.Fatalf("expected ffmpeg error, got %v", sound.Error)
	}
	if sound.ProcessedPath != nil {
		t.Fatalf("expected no processed path, got %q", *sound.ProcessedPath)
	}
}

func TestCreateUpdateDeleteRulePersistsConfig(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")

	created, err := app.CreateRule(RuleRequest{
		Name:      "Tests passed",
		Enabled:   true,
		EventType: "test_success",
		SoundID:   sound.ID,
		MatchMode: "contains",
	})
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}
	if len(created.Config.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(created.Config.Rules))
	}

	rule := created.Config.Rules[0]
	if rule.ID == "" {
		t.Fatal("expected rule id")
	}
	if rule.CreatedAt == "" || rule.UpdatedAt == "" {
		t.Fatal("expected rule timestamps")
	}
	if rule.SoundID != sound.ID {
		t.Fatalf("expected rule sound id %q, got %q", sound.ID, rule.SoundID)
	}

	updated, err := app.UpdateRule(RuleRequest{
		ID:             rule.ID,
		Name:           "Build failed",
		Enabled:        false,
		EventType:      "build_failure",
		SoundID:        sound.ID,
		MatchMode:      "startsWith",
		CommandPattern: "npm run build",
	})
	if err != nil {
		t.Fatalf("UpdateRule() error = %v", err)
	}
	if updated.Config.Rules[0].Name != "Build failed" {
		t.Fatalf("expected updated name, got %q", updated.Config.Rules[0].Name)
	}
	if updated.Config.Rules[0].Enabled {
		t.Fatal("expected updated rule to be disabled")
	}

	toggled, err := app.ToggleRule(rule.ID, true)
	if err != nil {
		t.Fatalf("ToggleRule() error = %v", err)
	}
	if !toggled.Config.Rules[0].Enabled {
		t.Fatal("expected toggled rule to be enabled")
	}

	reloaded, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() error = %v", err)
	}
	if len(reloaded.Config.Rules) != 1 {
		t.Fatalf("expected persisted rule, got %d", len(reloaded.Config.Rules))
	}

	deleted, err := app.DeleteRule(rule.ID)
	if err != nil {
		t.Fatalf("DeleteRule() error = %v", err)
	}
	if len(deleted.Config.Rules) != 0 {
		t.Fatalf("expected no rules after delete, got %d", len(deleted.Config.Rules))
	}
}

func TestCreateRuleValidation(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")

	if _, err := app.CreateRule(RuleRequest{Name: "", EventType: "test_success", SoundID: sound.ID}); err == nil {
		t.Fatal("expected empty name validation error")
	}
	if _, err := app.CreateRule(RuleRequest{Name: "Bad event", EventType: "unknown", SoundID: sound.ID}); err == nil {
		t.Fatal("expected event type validation error")
	}
	if _, err := app.CreateRule(RuleRequest{Name: "Missing sound", EventType: "test_success", SoundID: "missing"}); err == nil {
		t.Fatal("expected missing sound validation error")
	}
}

func TestRuleSurvivesDeletedSoundReference(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")

	created, err := app.CreateRule(RuleRequest{
		Name:      "Git push",
		Enabled:   true,
		EventType: "git_push_success",
		SoundID:   sound.ID,
	})
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	if _, err := app.DeleteSound(sound.ID); err != nil {
		t.Fatalf("DeleteSound() error = %v", err)
	}

	reloaded, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() error = %v", err)
	}
	if len(reloaded.Config.Rules) != 1 {
		t.Fatalf("expected rule to remain after sound delete, got %d", len(reloaded.Config.Rules))
	}
	if reloaded.Config.Rules[0].ID != created.Config.Rules[0].ID {
		t.Fatalf("expected same rule id, got %q", reloaded.Config.Rules[0].ID)
	}
	if _, err := app.TestRuleSound(created.Config.Rules[0].ID); err == nil {
		t.Fatal("expected TestRuleSound to report missing sound")
	}
}

func TestEvaluateEventMatchesEventType(t *testing.T) {
	sound := testSoundRecord("sound-1", "/tmp/original.wav", nil)
	config := AppConfig{
		Sounds: []SoundRecord{sound},
		Rules: []RuleRecord{
			testRuleRecord("rule-1", "test_success", sound.ID, "", "", nil, true),
		},
	}

	result := evaluateEvent(config, TerminalEvent{EventType: "test_success"})
	if !result.Matched || result.Rule == nil || result.Rule.ID != "rule-1" {
		t.Fatalf("expected rule-1 match, got %+v", result)
	}

	result = evaluateEvent(config, TerminalEvent{EventType: "test_failure"})
	if result.Matched {
		t.Fatalf("expected no match, got %+v", result)
	}
}

func TestEvaluateEventCommandMatchModes(t *testing.T) {
	sound := testSoundRecord("sound-1", "/tmp/original.wav", nil)
	tests := []struct {
		name    string
		mode    string
		pattern string
		command string
		matched bool
	}{
		{name: "contains", mode: "contains", pattern: "test", command: "npm test -- --watch", matched: true},
		{name: "startsWith", mode: "startsWith", pattern: "npm", command: "npm test", matched: true},
		{name: "exact", mode: "exact", pattern: "npm test", command: "npm test", matched: true},
		{name: "exact miss", mode: "exact", pattern: "npm test", command: "npm run test", matched: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := AppConfig{
				Sounds: []SoundRecord{sound},
				Rules: []RuleRecord{
					testRuleRecord("rule-1", "command_success", sound.ID, test.mode, test.pattern, nil, true),
				},
			}

			result := evaluateEvent(config, TerminalEvent{EventType: "command_success", Command: test.command})
			if result.Matched != test.matched {
				t.Fatalf("expected matched=%v, got %+v", test.matched, result)
			}
		})
	}
}

func TestEvaluateEventIgnoresDisabledRules(t *testing.T) {
	sound := testSoundRecord("sound-1", "/tmp/original.wav", nil)
	config := AppConfig{
		Sounds: []SoundRecord{sound},
		Rules: []RuleRecord{
			testRuleRecord("rule-1", "command_success", sound.ID, "", "", nil, false),
		},
	}

	result := evaluateEvent(config, TerminalEvent{EventType: "command_success"})
	if result.Matched {
		t.Fatalf("expected disabled rule to be ignored, got %+v", result)
	}
}

func TestEvaluateEventMissingSoundHandledSafely(t *testing.T) {
	config := AppConfig{
		Rules: []RuleRecord{
			testRuleRecord("rule-1", "command_success", "missing-sound", "", "", nil, true),
		},
	}

	result := evaluateEvent(config, TerminalEvent{EventType: "command_success"})
	if !result.Matched || !result.MissingSound || result.Sound != nil {
		t.Fatalf("expected missing sound match result, got %+v", result)
	}
}

func TestEvaluateEventPrefersProcessedPath(t *testing.T) {
	dir := t.TempDir()
	originalPath := filepath.Join(dir, "original.wav")
	processedPath := filepath.Join(dir, "processed.wav")
	if err := os.WriteFile(originalPath, []byte("original"), 0o644); err != nil {
		t.Fatalf("write original: %v", err)
	}
	if err := os.WriteFile(processedPath, []byte("processed"), 0o644); err != nil {
		t.Fatalf("write processed: %v", err)
	}

	sound := testSoundRecord("sound-1", originalPath, &processedPath)
	config := AppConfig{
		Sounds: []SoundRecord{sound},
		Rules: []RuleRecord{
			testRuleRecord("rule-1", "command_success", sound.ID, "", "", nil, true),
		},
	}

	result := evaluateEvent(config, TerminalEvent{EventType: "command_success"})
	if result.SoundPath != processedPath {
		t.Fatalf("expected processed path %q, got %q", processedPath, result.SoundPath)
	}
}

func TestEvaluateEventDeterministicPriority(t *testing.T) {
	exitCode := 0
	sound := testSoundRecord("sound-1", "/tmp/original.wav", nil)
	config := AppConfig{
		Sounds: []SoundRecord{sound},
		Rules: []RuleRecord{
			testRuleRecord("broad-first", "command_success", sound.ID, "any", "", nil, true),
			testRuleRecord("contains-second", "command_success", sound.ID, "contains", "test", nil, true),
			testRuleRecord("exact-third", "command_success", sound.ID, "exact", "npm test", nil, true),
			testRuleRecord("exact-with-exit-fourth", "command_success", sound.ID, "exact", "npm test", &exitCode, true),
		},
	}

	result := evaluateEvent(config, TerminalEvent{EventType: "command_success", Command: "npm test", ExitCode: &exitCode})
	if result.Rule == nil || result.Rule.ID != "exact-with-exit-fourth" {
		t.Fatalf("expected exact rule with exit code priority, got %+v", result)
	}

	config.Rules = []RuleRecord{
		testRuleRecord("first-created", "command_success", sound.ID, "exact", "npm test", nil, true),
		testRuleRecord("second-created", "command_success", sound.ID, "exact", "npm test", nil, true),
	}
	result = evaluateEvent(config, TerminalEvent{EventType: "command_success", Command: "npm test"})
	if result.Rule == nil || result.Rule.ID != "first-created" {
		t.Fatalf("expected first-created tie winner, got %+v", result)
	}
}

func TestHandleTerminalEventStartsPlaybackOnMatch(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")
	createTestRule(t, app, RuleRequest{Name: "Command success", Enabled: true, EventType: "command_success", SoundID: sound.ID})

	var playedPath string
	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		playedPath = path
		if !stopPrevious {
			t.Fatal("expected stopPrevious to default true")
		}
		return nil
	})
	defer restorePlayback()

	result, err := app.HandleTerminalEvent(TerminalEvent{EventType: "command_success"})
	if err != nil {
		t.Fatalf("HandleTerminalEvent() error = %v", err)
	}
	if !result.Matched || !result.PlaybackStarted {
		t.Fatalf("expected matched playback result, got %+v", result)
	}
	if playedPath != sound.OriginalPath {
		t.Fatalf("expected playback path %q, got %q", sound.OriginalPath, playedPath)
	}
}

func TestHandleTerminalEventNoMatch(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		t.Fatal("playback should not start for no match")
		return nil
	})
	defer restorePlayback()

	result, err := app.HandleTerminalEvent(TerminalEvent{EventType: "command_success"})
	if err != nil {
		t.Fatalf("HandleTerminalEvent() error = %v", err)
	}
	if result.Matched || result.PlaybackAttempted {
		t.Fatalf("expected no match without playback, got %+v", result)
	}
}

func TestHandleTerminalEventMissingSound(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")
	created := createTestRule(t, app, RuleRequest{Name: "Command success", Enabled: true, EventType: "command_success", SoundID: sound.ID})
	if _, err := app.DeleteSound(sound.ID); err != nil {
		t.Fatalf("DeleteSound() error = %v", err)
	}
	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		t.Fatal("playback should not start for missing sound")
		return nil
	})
	defer restorePlayback()

	result, err := app.HandleTerminalEvent(TerminalEvent{EventType: "command_success"})
	if err != nil {
		t.Fatalf("HandleTerminalEvent() error = %v", err)
	}
	if !result.Matched || !result.MissingSound || result.Rule == nil || result.Rule.ID != created.ID {
		t.Fatalf("expected missing sound match, got %+v", result)
	}
}

func TestHandleTerminalEventPlaybackDisabled(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")
	createTestRule(t, app, RuleRequest{Name: "Command success", Enabled: true, EventType: "command_success", SoundID: sound.ID})

	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() error = %v", err)
	}
	snapshot.Config.PlaybackEnabled = false
	if _, err := app.SaveConfig(snapshot.Config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		t.Fatal("playback should not start when disabled")
		return nil
	})
	defer restorePlayback()

	result, err := app.HandleTerminalEvent(TerminalEvent{EventType: "command_success"})
	if err != nil {
		t.Fatalf("HandleTerminalEvent() error = %v", err)
	}
	if !result.Matched || result.PlaybackAttempted || result.PlaybackEnabled {
		t.Fatalf("expected matched result with playback disabled, got %+v", result)
	}
}

func TestHandleTerminalEventListeningPaused(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")
	createTestRule(t, app, RuleRequest{Name: "Command success", Enabled: true, EventType: "command_success", SoundID: sound.ID})

	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatalf("LoadConfigSnapshot() error = %v", err)
	}
	snapshot.Config.Listening = false
	if _, err := app.SaveConfig(snapshot.Config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		t.Fatal("playback should not start while listening is paused")
		return nil
	})
	defer restorePlayback()

	result, err := app.HandleTerminalEvent(TerminalEvent{EventType: "command_success"})
	if err != nil {
		t.Fatalf("HandleTerminalEvent() error = %v", err)
	}
	if result.Matched || result.Message != "Listening is paused" {
		t.Fatalf("expected listening paused result, got %+v", result)
	}
}

func importTestSound(t *testing.T, name string) SoundRecord {
	t.Helper()

	sourcePath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(sourcePath, []byte("fake audio"), 0o644); err != nil {
		t.Fatalf("write source sound: %v", err)
	}

	snapshot, err := importSoundPaths([]string{sourcePath})
	if err != nil {
		t.Fatalf("importSoundPaths() error = %v", err)
	}
	return snapshot.Config.Sounds[len(snapshot.Config.Sounds)-1]
}

func createTestRule(t *testing.T, app *App, request RuleRequest) RuleRecord {
	t.Helper()

	snapshot, err := app.CreateRule(request)
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}
	return snapshot.Config.Rules[len(snapshot.Config.Rules)-1]
}

func stubPlayback(t *testing.T, start func(path string, stopPrevious bool) error) func() {
	t.Helper()

	originalStart := startPlayback
	originalStop := stopCurrentPlayback
	startPlayback = start
	stopCurrentPlayback = func() error { return nil }
	return func() {
		startPlayback = originalStart
		stopCurrentPlayback = originalStop
	}
}

func testSoundRecord(id string, originalPath string, processedPath *string) SoundRecord {
	return SoundRecord{
		ID:            id,
		Name:          id,
		OriginalPath:  originalPath,
		ProcessedPath: processedPath,
		Format:        "wav",
		Status:        "ready",
		CreatedAt:     "2026-01-01T00:00:00Z",
	}
}

func testRuleRecord(id string, eventType string, soundID string, matchMode string, commandPattern string, exitCode *int, enabled bool) RuleRecord {
	return RuleRecord{
		ID:             id,
		Name:           id,
		Enabled:        enabled,
		EventType:      eventType,
		SoundID:        soundID,
		MatchMode:      matchMode,
		CommandPattern: commandPattern,
		ExitCode:       exitCode,
		CreatedAt:      "2026-01-01T00:00:00Z",
		UpdatedAt:      "2026-01-01T00:00:00Z",
	}
}

func isolateUserDirs(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", filepath.Join(root, "AppData", "Roaming"))
	case "linux":
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, ".config"))
		t.Setenv("XDG_DATA_HOME", filepath.Join(root, ".local", "share"))
	default:
		t.Setenv("HOME", root)
	}
}
