package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const ConfigVersion = 2

func GetAppDataPaths() (AppDataPaths, error) { return appDataPaths(runtime.GOOS) }

func appDataPaths(goos string) (AppDataPaths, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return AppDataPaths{}, fmt.Errorf("find user config directory: %w", err)
	}
	name := "ProdTag"
	if goos == "linux" {
		name = "prodtag"
	}
	configDir := filepath.Join(configBase, name)
	dataDir := configDir
	if goos == "linux" {
		base, err := userDataDir()
		if err != nil {
			return AppDataPaths{}, err
		}
		dataDir = filepath.Join(base, name)
	}
	helper := "prodtag-helper"
	if goos == "windows" {
		helper += ".exe"
	}
	return AppDataPaths{
		ConfigDir: configDir, DataDir: dataDir, ConfigFile: filepath.Join(configDir, "config.json"), MatcherCacheFile: filepath.Join(configDir, "matcher-cache.json"),
		OriginalSoundsDir: filepath.Join(dataDir, "sounds", "originals"), ProcessedSoundsDir: filepath.Join(dataDir, "sounds", "processed"), LogsDir: filepath.Join(dataDir, "logs"),
		BinDir: filepath.Join(dataDir, "bin"), IntegrationsDir: filepath.Join(dataDir, "integrations"), HelperBinary: filepath.Join(dataDir, "bin", helper),
		ZshScript: filepath.Join(dataDir, "integrations", "prodtag.zsh"), BashScript: filepath.Join(dataDir, "integrations", "prodtag.bash"), PowerShellScript: filepath.Join(dataDir, "integrations", "prodtag.ps1"),
	}, nil
}

func EnsureAppData() error {
	paths, err := GetAppDataPaths()
	if err != nil {
		return err
	}
	for _, dir := range []string{paths.ConfigDir, paths.DataDir, paths.OriginalSoundsDir, paths.ProcessedSoundsDir, paths.LogsDir, paths.BinDir, paths.IntegrationsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	if err := createJSONIfMissing(paths.ConfigFile, DefaultConfig()); err != nil {
		return err
	}
	return createJSONIfMissing(paths.MatcherCacheFile, DefaultMatcherCache())
}

func LoadConfigSnapshot() (ConfigSnapshot, error) {
	if err := EnsureAppData(); err != nil {
		return ConfigSnapshot{}, err
	}
	paths, err := GetAppDataPaths()
	if err != nil {
		return ConfigSnapshot{}, err
	}
	config, err := ReadConfig(paths.ConfigFile)
	if err != nil {
		return ConfigSnapshot{}, err
	}
	if err := EnsureMatcherCache(paths.MatcherCacheFile, config); err != nil {
		return ConfigSnapshot{}, err
	}
	return ConfigSnapshot{Config: config, Paths: paths}, nil
}

func SaveConfig(config AppConfig) (ConfigSnapshot, error) {
	if err := EnsureAppData(); err != nil {
		return ConfigSnapshot{}, err
	}
	paths, err := GetAppDataPaths()
	if err != nil {
		return ConfigSnapshot{}, err
	}
	config = NormalizeConfig(config)
	config.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := WriteJSON(paths.ConfigFile, config); err != nil {
		return ConfigSnapshot{}, err
	}
	if err := WriteMatcherCache(paths.MatcherCacheFile, config); err != nil {
		return ConfigSnapshot{}, err
	}
	return ConfigSnapshot{Config: config, Paths: paths}, nil
}

func ReadConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}
	var c AppConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return AppConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return NormalizeConfig(c), nil
}
func NormalizeConfig(c AppConfig) AppConfig {
	original := c.Version
	if original == 0 {
		original = 1
	}
	if original < 2 {
		c.EventEngineEnabled = true
		c.PlaybackEnabled = true
		c.StopPreviousOnNewEvent = true
	}
	c.Version = ConfigVersion
	if c.Sounds == nil {
		c.Sounds = []SoundRecord{}
	}
	if c.Playlists == nil {
		c.Playlists = []PlaylistRecord{}
	}
	if c.Rules == nil {
		c.Rules = []RuleRecord{}
	}
	if c.UpdatedAt == "" {
		c.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	for i := range c.Sounds {
		if c.Sounds[i].Status == "" {
			c.Sounds[i].Status = "imported"
		}
		if c.Sounds[i].CreatedAt == "" {
			c.Sounds[i].CreatedAt = c.UpdatedAt
		}
	}
	for i := range c.Rules {
		if c.Rules[i].CreatedAt == "" {
			c.Rules[i].CreatedAt = c.UpdatedAt
		}
		if c.Rules[i].UpdatedAt == "" {
			c.Rules[i].UpdatedAt = c.Rules[i].CreatedAt
		}
	}
	return c
}
func DefaultConfig() AppConfig {
	return NormalizeConfig(AppConfig{Version: ConfigVersion, Listening: true, EventEngineEnabled: true, PlaybackEnabled: true, StopPreviousOnNewEvent: true, Hotkeys: HotkeySettings{}, Integrations: IntegrationSettings{}})
}
func DefaultMatcherCache() MatcherCache {
	return MatcherCache{Version: ConfigVersion, Complete: true, EnabledEventTypes: []string{}, BroadEventTypes: []string{}, Candidates: []MatcherCacheCandidate{}, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
}
func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary json: %w", err)
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
func createJSONIfMissing(path string, value any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return WriteJSON(path, value)
}
func userDataDir() (string, error) {
	if value := os.Getenv("XDG_DATA_HOME"); value != "" {
		return value, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("user home directory is empty")
	}
	return filepath.Join(home, ".local", "share"), nil
}
