package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	configVersion = 1
	appName       = "ProdTag"
	linuxAppName  = "prodtag"
)

// AppDataPaths describes the cross-platform local folders ProdTag owns.
type AppDataPaths struct {
	ConfigDir          string `json:"configDir"`
	DataDir            string `json:"dataDir"`
	ConfigFile         string `json:"configFile"`
	MatcherCacheFile   string `json:"matcherCacheFile"`
	OriginalSoundsDir  string `json:"originalSoundsDir"`
	ProcessedSoundsDir string `json:"processedSoundsDir"`
	LogsDir            string `json:"logsDir"`
}

type ConfigSnapshot struct {
	Config AppConfig    `json:"config"`
	Paths  AppDataPaths `json:"paths"`
}

type AppConfig struct {
	Version               int                 `json:"version"`
	Listening             bool                `json:"listening"`
	Muted                 bool                `json:"muted"`
	LaunchHelperAtStartup bool                `json:"launchHelperAtStartup"`
	Sounds                []SoundRecord       `json:"sounds"`
	Playlists             []PlaylistRecord    `json:"playlists"`
	Rules                 []RuleRecord        `json:"rules"`
	Hotkeys               HotkeySettings      `json:"hotkeys"`
	Integrations          IntegrationSettings `json:"integrations"`
	UpdatedAt             string              `json:"updatedAt"`
}

type SoundRecord struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	OriginalPath  string  `json:"originalPath"`
	ProcessedPath *string `json:"processedPath"`
	DurationMs    *int64  `json:"durationMs"`
	Format        string  `json:"format"`
	CreatedAt     string  `json:"createdAt"`
	Status        string  `json:"status"`
	Error         *string `json:"error"`
}

type PlaylistRecord struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	SoundIDs []string `json:"soundIds"`
}

type RuleRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Pattern     string `json:"pattern"`
	MatchType   string `json:"matchType"`
	Condition   string `json:"condition"`
	SoundID     string `json:"soundId,omitempty"`
	PlaylistID  string `json:"playlistId,omitempty"`
	CooldownMs  int64  `json:"cooldownMs,omitempty"`
	Probability int    `json:"probability,omitempty"`
}

type HotkeySettings struct {
	StopAudio       string `json:"stopAudio"`
	ToggleListening string `json:"toggleListening"`
	ToggleMute      string `json:"toggleMute"`
	OpenApp         string `json:"openApp"`
}

type IntegrationSettings struct {
	Zsh        ShellIntegrationState `json:"zsh"`
	Bash       ShellIntegrationState `json:"bash"`
	PowerShell ShellIntegrationState `json:"powerShell"`
}

type ShellIntegrationState struct {
	Installed bool   `json:"installed"`
	Scope     string `json:"scope"`
	LastCheck string `json:"lastCheck"`
}

type MatcherCache struct {
	Version    int                     `json:"version"`
	Candidates []MatcherCacheCandidate `json:"candidates"`
	UpdatedAt  string                  `json:"updatedAt"`
}

type MatcherCacheCandidate struct {
	RuleID    string `json:"ruleId"`
	Pattern   string `json:"pattern"`
	MatchType string `json:"matchType"`
}

func (a *App) LoadConfig() (ConfigSnapshot, error) {
	return LoadConfigSnapshot()
}

func (a *App) SaveConfig(config AppConfig) (ConfigSnapshot, error) {
	if err := EnsureAppData(); err != nil {
		return ConfigSnapshot{}, err
	}

	paths, err := GetAppDataPaths()
	if err != nil {
		return ConfigSnapshot{}, err
	}

	config = normalizeConfig(config)
	config.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := writeJSON(paths.ConfigFile, config); err != nil {
		return ConfigSnapshot{}, err
	}

	return ConfigSnapshot{
		Config: config,
		Paths:  paths,
	}, nil
}

func LoadConfigSnapshot() (ConfigSnapshot, error) {
	if err := EnsureAppData(); err != nil {
		return ConfigSnapshot{}, err
	}

	paths, err := GetAppDataPaths()
	if err != nil {
		return ConfigSnapshot{}, err
	}

	config, err := readConfig(paths.ConfigFile)
	if err != nil {
		return ConfigSnapshot{}, err
	}

	return ConfigSnapshot{
		Config: config,
		Paths:  paths,
	}, nil
}

func EnsureAppData() error {
	paths, err := GetAppDataPaths()
	if err != nil {
		return err
	}

	dirs := []string{
		paths.ConfigDir,
		paths.DataDir,
		paths.OriginalSoundsDir,
		paths.ProcessedSoundsDir,
		paths.LogsDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	if err := createJSONIfMissing(paths.ConfigFile, defaultConfig()); err != nil {
		return err
	}
	if err := createJSONIfMissing(paths.MatcherCacheFile, defaultMatcherCache()); err != nil {
		return err
	}

	return nil
}

func GetAppDataPaths() (AppDataPaths, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return AppDataPaths{}, fmt.Errorf("find user config directory: %w", err)
	}

	configName := appName
	dataName := appName
	if runtime.GOOS == "linux" {
		configName = linuxAppName
		dataName = linuxAppName
	}

	configDir := filepath.Join(configBase, configName)
	dataDir := configDir
	if runtime.GOOS == "linux" {
		dataBase, err := userDataDir()
		if err != nil {
			return AppDataPaths{}, err
		}
		dataDir = filepath.Join(dataBase, dataName)
	}

	return AppDataPaths{
		ConfigDir:          configDir,
		DataDir:            dataDir,
		ConfigFile:         filepath.Join(configDir, "config.json"),
		MatcherCacheFile:   filepath.Join(configDir, "matcher-cache.json"),
		OriginalSoundsDir:  filepath.Join(dataDir, "sounds", "originals"),
		ProcessedSoundsDir: filepath.Join(dataDir, "sounds", "processed"),
		LogsDir:            filepath.Join(dataDir, "logs"),
	}, nil
}

func userDataDir() (string, error) {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return dataHome, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find user home directory: %w", err)
	}
	if home == "" {
		return "", errors.New("user home directory is empty")
	}

	return filepath.Join(home, ".local", "share"), nil
}

func readConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return AppConfig{}, fmt.Errorf("parse config: %w", err)
	}

	return normalizeConfig(config), nil
}

func normalizeConfig(config AppConfig) AppConfig {
	if config.Version == 0 {
		config.Version = configVersion
	}
	if config.Sounds == nil {
		config.Sounds = []SoundRecord{}
	}
	if config.UpdatedAt == "" {
		config.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	for index := range config.Sounds {
		if config.Sounds[index].Status == "" {
			config.Sounds[index].Status = "imported"
		}
		if config.Sounds[index].CreatedAt == "" {
			config.Sounds[index].CreatedAt = config.UpdatedAt
		}
	}
	if config.Playlists == nil {
		config.Playlists = []PlaylistRecord{}
	}
	if config.Rules == nil {
		config.Rules = []RuleRecord{}
	}

	return config
}

func defaultConfig() AppConfig {
	return normalizeConfig(AppConfig{
		Version:   configVersion,
		Listening: true,
		Muted:     false,
		Hotkeys: HotkeySettings{
			StopAudio:       "",
			ToggleListening: "",
			ToggleMute:      "",
			OpenApp:         "",
		},
		Integrations: IntegrationSettings{
			Zsh:        ShellIntegrationState{Installed: false, Scope: ""},
			Bash:       ShellIntegrationState{Installed: false, Scope: ""},
			PowerShell: ShellIntegrationState{Installed: false, Scope: ""},
		},
	})
}

func defaultMatcherCache() MatcherCache {
	return MatcherCache{
		Version:    configVersion,
		Candidates: []MatcherCacheCandidate{},
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func createJSONIfMissing(path string, value any) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check %s: %w", path, err)
	}

	if err := writeJSON(path, value); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	data = append(data, '\n')
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tempPath, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil
}
