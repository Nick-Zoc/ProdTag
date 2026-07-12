//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestZshDebugLogCreatedWhenCacheSkips(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh unavailable")
	}
	dir := t.TempDir()
	cache := filepath.Join(dir, "matcher-cache.json")
	config := filepath.Join(dir, "config.json")
	log := filepath.Join(dir, "prodtag-zsh-debug.log")
	cacheData := "{\n  \"version\": 2,\n  \"complete\": true,\n  \"enabledEventTypes\": []\n}\n"
	if err := os.WriteFile(config, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte(cacheData), 0o644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(projectRootForTest(t), "scripts", "prodtag.zsh")
	command := "source \"$SCRIPT\"; _prodtag_debug 'captured command=ls exit_code=0 event_type=command_success'; if ! _prodtag_cache_may_match command_success; then _prodtag_debug \"cache_state=$_prodtag_cache_state launch=skipped reason=$_prodtag_cache_reason\"; fi"
	cmd := exec.Command("zsh", "-c", command)
	cmd.Env = append(os.Environ(), "SCRIPT="+script, "PRODTAG_ZSH_DEBUG=1", "PRODTAG_ZSH_DEBUG_LOG="+log, "PRODTAG_MATCHER_CACHE="+cache, "PRODTAG_CONFIG="+config)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zsh debug test: %v: %s", err, output)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("debug log missing: %v", err)
	}
	text := string(data)
	for _, part := range []string{"captured command=ls", "cache_state=valid", "launch=skipped", "no enabled rule"} {
		if !strings.Contains(text, part) {
			t.Fatalf("debug log missing %q: %s", part, text)
		}
	}
}
func projectRootForTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}
