package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var allowedAudioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".ogg":  true,
	".flac": true,
}

type RenameSoundRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *App) ImportSoundWithPicker() (ConfigSnapshot, error) {
	paths, err := a.SelectSoundFiles()
	if err != nil {
		return ConfigSnapshot{}, err
	}
	if len(paths) == 0 {
		return LoadConfigSnapshot()
	}

	return importSoundPaths(paths)
}

func (a *App) SelectSoundFiles() ([]string, error) {
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import sounds",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Audio Files (*.mp3, *.wav, *.m4a, *.ogg, *.flac)",
				Pattern:     "*.mp3;*.wav;*.m4a;*.ogg;*.flac",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (a *App) ImportSoundPaths(paths []string) (ConfigSnapshot, error) {
	return importSoundPaths(paths)
}

func (a *App) RenameSound(request RenameSoundRequest) (ConfigSnapshot, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return ConfigSnapshot{}, errors.New("sound name cannot be empty")
	}

	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return ConfigSnapshot{}, err
	}

	found := false
	for index := range snapshot.Config.Sounds {
		if snapshot.Config.Sounds[index].ID == request.ID {
			snapshot.Config.Sounds[index].Name = name
			found = true
			break
		}
	}
	if !found {
		return ConfigSnapshot{}, fmt.Errorf("sound %s not found", request.ID)
	}

	return a.SaveConfig(snapshot.Config)
}

func (a *App) DeleteSound(id string) (ConfigSnapshot, error) {
	return a.DeleteSounds([]string{id})
}

func (a *App) DeleteSounds(ids []string) (ConfigSnapshot, error) {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return ConfigSnapshot{}, err
	}
	if len(ids) == 0 {
		return snapshot, nil
	}

	deleteIDs := make(map[string]bool, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			deleteIDs[id] = true
		}
	}
	if len(deleteIDs) == 0 {
		return snapshot, nil
	}

	nextSounds := make([]SoundRecord, 0, len(snapshot.Config.Sounds))
	removed := make([]SoundRecord, 0, len(deleteIDs))
	for index := range snapshot.Config.Sounds {
		sound := snapshot.Config.Sounds[index]
		if deleteIDs[sound.ID] {
			removed = append(removed, sound)
			continue
		}
		nextSounds = append(nextSounds, sound)
	}
	if len(removed) != len(deleteIDs) {
		return ConfigSnapshot{}, errors.New("one or more selected sounds were not found")
	}

	for _, sound := range removed {
		if err := removeLibraryFile(sound.OriginalPath, snapshot.Paths.OriginalSoundsDir); err != nil {
			return ConfigSnapshot{}, err
		}
	}

	snapshot.Config.Sounds = nextSounds
	return a.SaveConfig(snapshot.Config)
}

func (a *App) GetSoundPreviewDataURL(id string) (string, error) {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return "", err
	}

	var sound *SoundRecord
	for index := range snapshot.Config.Sounds {
		if snapshot.Config.Sounds[index].ID == id {
			sound = &snapshot.Config.Sounds[index]
			break
		}
	}
	if sound == nil {
		return "", fmt.Errorf("sound %s not found", id)
	}

	data, err := os.ReadFile(sound.OriginalPath)
	if err != nil {
		return "", fmt.Errorf("read sound file: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(sound.OriginalPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

func importSoundPaths(paths []string) (ConfigSnapshot, error) {
	if err := EnsureAppData(); err != nil {
		return ConfigSnapshot{}, err
	}

	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return ConfigSnapshot{}, err
	}

	for _, sourcePath := range paths {
		if err := validateSoundSource(sourcePath); err != nil {
			return ConfigSnapshot{}, err
		}
	}

	for _, sourcePath := range paths {
		record, err := importSingleSound(sourcePath, snapshot.Paths.OriginalSoundsDir)
		if err != nil {
			return ConfigSnapshot{}, err
		}
		snapshot.Config.Sounds = append(snapshot.Config.Sounds, record)
	}

	app := NewApp()
	return app.SaveConfig(snapshot.Config)
}

func importSingleSound(sourcePath string, originalsDir string) (SoundRecord, error) {
	sourcePath = strings.TrimSpace(sourcePath)
	if err := validateSoundSource(sourcePath); err != nil {
		return SoundRecord{}, err
	}

	extension := strings.ToLower(filepath.Ext(sourcePath))
	id, err := newSoundID()
	if err != nil {
		return SoundRecord{}, err
	}

	if err := os.MkdirAll(originalsDir, 0o755); err != nil {
		return SoundRecord{}, fmt.Errorf("create originals folder: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	name := strings.TrimSpace(baseName)
	if name == "" {
		name = "Imported sound"
	}

	destination := filepath.Join(originalsDir, fmt.Sprintf("%s-%s%s", id, sanitizeFileName(name), extension))
	if err := copyFile(sourcePath, destination); err != nil {
		return SoundRecord{}, err
	}

	return SoundRecord{
		ID:            id,
		Name:          name,
		OriginalPath:  destination,
		ProcessedPath: nil,
		DurationMs:    nil,
		Format:        strings.TrimPrefix(extension, "."),
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Status:        "imported",
		Error:         nil,
	}, nil
}

func validateSoundSource(sourcePath string) error {
	sourcePath = strings.TrimSpace(sourcePath)
	if sourcePath == "" {
		return errors.New("sound path is empty")
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("open selected sound: %w", err)
	}
	if info.IsDir() {
		return errors.New("selected item is a folder, not an audio file")
	}

	extension := strings.ToLower(filepath.Ext(sourcePath))
	if !allowedAudioExtensions[extension] {
		return fmt.Errorf("unsupported audio type %s", extension)
	}

	return nil
}

func copyFile(sourcePath string, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open selected sound: %w", err)
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create library copy: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("copy sound to library: %w", err)
	}

	return nil
}

func removeLibraryFile(path string, allowedRoot string) error {
	if path == "" {
		return nil
	}

	rel, err := filepath.Rel(allowedRoot, path)
	if err != nil {
		return fmt.Errorf("check copied sound path: %w", err)
	}
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to delete file outside sound library: %s", path)
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete copied sound file: %w", err)
	}

	return nil
}

func newSoundID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("create sound id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func sanitizeFileName(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '-' || char == '_':
			builder.WriteRune(char)
		case char == ' ' || char == '.':
			builder.WriteRune('-')
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "sound"
	}
	return result
}
