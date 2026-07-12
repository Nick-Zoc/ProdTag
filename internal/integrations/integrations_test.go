package integrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellMarkerInstallRemovePreservesExistingConfig(t *testing.T) {
	for _, shell := range []string{"bash", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "profile")
			original := "export KEEP=this\nfunction keep_me { echo yes; }\n"
			if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
				t.Fatal(err)
			}
			block := MarkerStart + "\n# " + shell + " integration\n" + MarkerEnd
			if err := InstallMarkedBlock(path, block); err != nil {
				t.Fatal(err)
			}
			if err := InstallMarkedBlock(path, block); err != nil {
				t.Fatal(err)
			}
			data, _ := os.ReadFile(path)
			if strings.Count(string(data), MarkerStart) != 1 {
				t.Fatalf("duplicate marker: %s", data)
			}
			if _, err := os.Stat(path + ".prodtag-backup"); err != nil {
				t.Fatalf("backup missing: %v", err)
			}
			if err := RemoveMarkedBlock(path); err != nil {
				t.Fatal(err)
			}
			data, _ = os.ReadFile(path)
			if string(data) != original {
				t.Fatalf("profile changed: %q", data)
			}
		})
	}
}
func TestPartialMarkerIsDetectedAndProtected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile")
	if err := os.WriteFile(path, []byte("keep\n"+MarkerStart+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if state := DetectMarkerState(path); state != "partial" {
		t.Fatalf("state=%q", state)
	}
	if err := InstallMarkedBlock(path, MarkerStart+"\n"+MarkerEnd); err == nil {
		t.Fatal("expected install refusal")
	}
	if err := RemoveMarkedBlock(path); err == nil {
		t.Fatal("expected removal refusal")
	}
}
