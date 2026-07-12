package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatcherCacheUpdatesAfterRuleMutations(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "cache.wav")
	created := createTestRule(t, app, RuleRequest{Name: "Tests", Enabled: true, EventType: "test_success", SoundID: sound.ID, MatchMode: "contains", CommandPattern: "npm test"})

	cache := loadTestMatcherCache(t)
	if len(cache.EnabledEventTypes) != 1 || cache.EnabledEventTypes[0] != "test_success" || len(cache.Candidates) != 1 {
		t.Fatalf("unexpected created cache: %+v", cache)
	}
	if cache.Candidates[0].RuleID != created.ID || cache.Candidates[0].Pattern != "npm test" {
		t.Fatalf("unexpected cache candidate: %+v", cache.Candidates[0])
	}

	if _, err := app.ToggleRule(created.ID, false); err != nil {
		t.Fatalf("ToggleRule() error = %v", err)
	}
	cache = loadTestMatcherCache(t)
	if len(cache.EnabledEventTypes) != 0 || len(cache.Candidates) != 0 {
		t.Fatalf("disabled rule remained in cache: %+v", cache)
	}

	if _, err := app.ToggleRule(created.ID, true); err != nil {
		t.Fatalf("ToggleRule() error = %v", err)
	}
	if _, err := app.DeleteRule(created.ID); err != nil {
		t.Fatalf("DeleteRule() error = %v", err)
	}
	cache = loadTestMatcherCache(t)
	if len(cache.EnabledEventTypes) != 0 || len(cache.Candidates) != 0 {
		t.Fatalf("deleted rule remained in cache: %+v", cache)
	}
}

func TestZshrcInstallIsIdempotentAndRemovalPreservesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	original := "export EDITOR=vim\nalias ll='ls -la'\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	block := zshMarkerStart + "\nsource '/tmp/prodtag.zsh'\n" + zshMarkerEnd
	if err := installZshrcBlock(path, block); err != nil {
		t.Fatalf("installZshrcBlock() error = %v", err)
	}
	if err := installZshrcBlock(path, block); err != nil {
		t.Fatalf("second installZshrcBlock() error = %v", err)
	}
	data, _ := os.ReadFile(path)
	if strings.Count(string(data), zshMarkerStart) != 1 {
		t.Fatalf("marker duplicated: %s", data)
	}
	backups, _ := filepath.Glob(path + ".prodtag-backup*")
	if len(backups) != 1 {
		t.Fatalf("expected one first-modification backup, got %d", len(backups))
	}
	if err := removeZshrcBlock(path); err != nil {
		t.Fatalf("removeZshrcBlock() error = %v", err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != original {
		t.Fatalf("unrelated zshrc content changed: %q", data)
	}
}

func TestZshrcPartialMarkerIsDetectedAndRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	if err := os.WriteFile(path, []byte("alias gs='git status'\n"+zshMarkerStart+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := detectZshrcMarkerState(path); got != "partial" {
		t.Fatalf("state = %q, want partial", got)
	}
	if err := installZshrcBlock(path, zshMarkerStart+"\n"+zshMarkerEnd); err == nil {
		t.Fatal("expected partial marker install refusal")
	}
	if err := removeZshrcBlock(path); err == nil {
		t.Fatal("expected partial marker removal refusal")
	}
}

func TestZshIntegrationStatusDetectsInstalledFiles(t *testing.T) {
	isolateUserDirs(t)
	if err := EnsureAppData(); err != nil {
		t.Fatal(err)
	}
	paths, _ := GetAppDataPaths()
	if err := os.WriteFile(paths.HelperBinary, []byte("helper"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ZshScript, []byte("# script"), 0o755); err != nil {
		t.Fatal(err)
	}
	zshrc := filepath.Join(os.Getenv("HOME"), ".zshrc")
	block := zshMarkerStart + "\nsource '" + paths.ZshScript + "'\n" + zshMarkerEnd
	if err := os.WriteFile(zshrc, []byte(block+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := getZshIntegrationStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !status.HelperInstalled || !status.HelperExecutable || !status.ScriptInstalled || status.ProfileState != "configured" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestHandledEventLogRetention(t *testing.T) {
	isolateUserDirs(t)
	for index := 0; index < handledEventLogMaxLines+25; index++ {
		record := RecentEventRecord{ID: fmt.Sprintf("event-%d", index), Timestamp: "2026-01-01T00:00:00Z"}
		if err := appendHandledEventLog(record); err != nil {
			t.Fatalf("appendHandledEventLog() error = %v", err)
		}
	}
	path, _ := handledEventLogPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != handledEventLogMaxLines {
		t.Fatalf("retained lines = %d, want %d", len(lines), handledEventLogMaxLines)
	}
	var first RecentEventRecord
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatal(err)
	}
	if first.ID != "event-25" {
		t.Fatalf("oldest retained ID = %q, want event-25", first.ID)
	}
}

func loadTestMatcherCache(t *testing.T) MatcherCache {
	t.Helper()
	paths, err := GetAppDataPaths()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := readMatcherCache(paths.MatcherCacheFile)
	if err != nil {
		t.Fatal(err)
	}
	return cache
}

func TestRunHelperEmitMatchesAndLogsWithoutRealPlayback(t *testing.T) {
	isolateUserDirs(t)
	app := NewApp()
	sound := importTestSound(t, "success.wav")
	rule := createTestRule(t, app, RuleRequest{
		Name:      "Command success",
		Enabled:   true,
		EventType: "command_success",
		SoundID:   sound.ID,
	})

	var playedPath string
	restorePlayback := stubPlayback(t, func(path string, stopPrevious bool) error {
		playedPath = path
		return nil
	})
	defer restorePlayback()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runHelper([]string{
		"emit",
		"--event-type", "command_success",
		"--command", "ls -a",
		"--exit-code", "0",
		"--cwd", "/tmp/prodtag-test",
		"--duration-ms", "120",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runHelper emit code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), rule.Name) {
		t.Fatalf("expected helper output to mention matched rule, got %q", stdout.String())
	}
	if playedPath != sound.OriginalPath {
		t.Fatalf("expected playback path %q, got %q", sound.OriginalPath, playedPath)
	}

	records, err := readHandledEventLog(10)
	if err != nil {
		t.Fatalf("readHandledEventLog() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(records))
	}
	record := records[0]
	if record.RuleID != rule.ID || record.SoundID != sound.ID || !record.PlaybackStarted {
		t.Fatalf("unexpected log record: %+v", record)
	}
	if record.Event.Command != "ls -a" || record.Event.DurationMs == nil || *record.Event.DurationMs != 120 {
		t.Fatalf("expected command and duration in log record, got %+v", record.Event)
	}
}

func TestHandledEventLogRoundTrip(t *testing.T) {
	isolateUserDirs(t)

	exitCode := 1
	durationMs := int64(77)
	record := RecentEventRecord{
		ID: "event-1",
		Event: TerminalEvent{
			EventType:  "command_failure",
			Command:    "false",
			ExitCode:   &exitCode,
			Cwd:        "/tmp/project",
			Timestamp:  "2026-01-01T00:00:00Z",
			DurationMs: &durationMs,
		},
		Matched:            false,
		PlaybackSkipped:    true,
		PlaybackSkipReason: "No matching rule",
		Message:            "No matching rule",
		Timestamp:          "2026-01-01T00:00:01Z",
	}

	if err := appendHandledEventLog(record); err != nil {
		t.Fatalf("appendHandledEventLog() error = %v", err)
	}

	records, err := readHandledEventLog(10)
	if err != nil {
		t.Fatalf("readHandledEventLog() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(records))
	}
	if records[0].Event.EventType != "command_failure" || records[0].Event.Command != "false" {
		t.Fatalf("unexpected log record: %+v", records[0])
	}

	path, err := handledEventLogPath()
	if err != nil {
		t.Fatalf("handledEventLogPath() error = %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join("logs", handledEventLogFileName)) {
		t.Fatalf("unexpected handled event log path: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
}

func TestInferTerminalEventType(t *testing.T) {
	success := 0
	failure := 2

	tests := []struct {
		command  string
		exitCode *int
		want     string
	}{
		{command: "git commit -m init", exitCode: &success, want: "git_commit_success"},
		{command: "git push origin main", exitCode: &success, want: "git_push_success"},
		{command: "npm test -- --runInBand", exitCode: &failure, want: "test_failure"},
		{command: "go test ./...", exitCode: &success, want: "test_success"},
		{command: "npm run build", exitCode: &failure, want: "build_failure"},
		{command: "ls -a", exitCode: &success, want: "command_success"},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			if got := InferTerminalEventType(test.command, test.exitCode); got != test.want {
				t.Fatalf("InferTerminalEventType() = %q, want %q", got, test.want)
			}
		})
	}
}
