package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
