package main

import (
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
