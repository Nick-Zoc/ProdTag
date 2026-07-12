package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestInferTerminalEventType(t *testing.T) {
	success, failure := 0, 1
	tests := []struct {
		command string
		code    *int
		want    string
	}{{"ls", &success, "command_success"}, {"git commit -m x", &failure, "git_commit_failure"}, {"git push", &success, "git_push_success"}, {"npm test", &success, "test_success"}, {"go build ./...", &failure, "build_failure"}}
	for _, test := range tests {
		if got := InferTerminalEventType(test.command, test.code); got != test.want {
			t.Errorf("InferTerminalEventType(%q)=%q want %q", test.command, got, test.want)
		}
	}
}
func TestWindowsPlaybackSelection(t *testing.T) {
	choice := selectPlayback("windows", lookupWith("pwsh", "powershell"))
	if choice.Method != "pwsh" {
		t.Fatalf("method=%q", choice.Method)
	}
	choice = selectPlayback("windows", lookupWith("powershell"))
	if choice.Method != "powershell" {
		t.Fatalf("fallback=%q", choice.Method)
	}
}
func TestLinuxPlaybackSelection(t *testing.T) {
	choice := selectPlayback("linux", lookupWith("ffplay", "aplay"))
	if choice.Method != "aplay" {
		t.Fatalf("priority=%q", choice.Method)
	}
	choice = selectPlayback("linux", lookupWith("ffplay"))
	if choice.Method != "ffplay" {
		t.Fatalf("fallback=%q", choice.Method)
	}
	choice = selectPlayback("linux", lookupWith())
	if choice.Method != "" || len(choice.Suggestions) == 0 {
		t.Fatalf("missing choice=%+v", choice)
	}
}
func lookupWith(names ...string) commandLookup {
	available := map[string]bool{}
	for _, name := range names {
		available[name] = true
	}
	return func(name string) (string, error) {
		if available[name] {
			return filepath.Join("/tools", name), nil
		}
		return "", errors.New("missing")
	}
}

func TestConcurrentEventLogWritesRemainReadable(t *testing.T) {
	isolateDirs(t)
	const count = 40
	var wg sync.WaitGroup
	errorsCh := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errorsCh <- AppendHandledEventLog(RecentEventRecord{ID: fmt.Sprintf("event-%d", index), Timestamp: "2026-01-01T00:00:00Z"})
		}(i)
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	records, err := ReadHandledEventLog(count)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != count {
		t.Fatalf("records=%d want %d", len(records), count)
	}
}
func TestMatcherCacheInvalidFallbackRegenerates(t *testing.T) {
	isolateDirs(t)
	paths, err := GetAppDataPaths()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureAppData(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.MatcherCacheFile, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := ReadMatcherCache(snapshot.Paths.MatcherCacheFile)
	if err != nil || !cache.Complete {
		t.Fatalf("cache=%+v err=%v", cache, err)
	}
}
func isolateDirs(t *testing.T) {
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
