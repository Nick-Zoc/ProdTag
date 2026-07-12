package core

type AppDataPaths struct {
	ConfigDir          string `json:"configDir"`
	DataDir            string `json:"dataDir"`
	ConfigFile         string `json:"configFile"`
	MatcherCacheFile   string `json:"matcherCacheFile"`
	OriginalSoundsDir  string `json:"originalSoundsDir"`
	ProcessedSoundsDir string `json:"processedSoundsDir"`
	LogsDir            string `json:"logsDir"`
	BinDir             string `json:"binDir"`
	IntegrationsDir    string `json:"integrationsDir"`
	HelperBinary       string `json:"helperBinary"`
	ZshScript          string `json:"zshScript"`
	BashScript         string `json:"bashScript"`
	PowerShellScript   string `json:"powerShellScript"`
}

type ConfigSnapshot struct {
	Config AppConfig    `json:"config"`
	Paths  AppDataPaths `json:"paths"`
}
type AppConfig struct {
	Version                int                 `json:"version"`
	Listening              bool                `json:"listening"`
	Muted                  bool                `json:"muted"`
	EventEngineEnabled     bool                `json:"eventEngineEnabled"`
	PlaybackEnabled        bool                `json:"playbackEnabled"`
	StopPreviousOnNewEvent bool                `json:"stopPreviousSoundOnNewEvent"`
	LocalEventPort         int                 `json:"localEventPort,omitempty"`
	LaunchHelperAtStartup  bool                `json:"launchHelperAtStartup"`
	Sounds                 []SoundRecord       `json:"sounds"`
	Playlists              []PlaylistRecord    `json:"playlists"`
	Rules                  []RuleRecord        `json:"rules"`
	Hotkeys                HotkeySettings      `json:"hotkeys"`
	Integrations           IntegrationSettings `json:"integrations"`
	UpdatedAt              string              `json:"updatedAt"`
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
	ID             string `json:"id"`
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled"`
	EventType      string `json:"eventType"`
	SoundID        string `json:"soundId"`
	MatchMode      string `json:"matchMode,omitempty"`
	CommandPattern string `json:"commandPattern,omitempty"`
	ExitCode       *int   `json:"exitCode,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	PlaylistID     string `json:"playlistId,omitempty"`
	CooldownMs     int64  `json:"cooldownMs,omitempty"`
	Probability    int    `json:"probability,omitempty"`
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
	Version           int                     `json:"version"`
	Complete          bool                    `json:"complete"`
	EnabledEventTypes []string                `json:"enabledEventTypes"`
	BroadEventTypes   []string                `json:"broadEventTypes"`
	Candidates        []MatcherCacheCandidate `json:"candidates"`
	UpdatedAt         string                  `json:"updatedAt"`
}
type MatcherCacheCandidate struct {
	RuleID      string `json:"ruleId"`
	EventType   string `json:"eventType"`
	Pattern     string `json:"pattern,omitempty"`
	MatchMode   string `json:"matchMode"`
	HasExitCode bool   `json:"hasExitCode"`
}
type TerminalEvent struct {
	EventType  string `json:"eventType"`
	Command    string `json:"command,omitempty"`
	ExitCode   *int   `json:"exitCode,omitempty"`
	Cwd        string `json:"cwd,omitempty"`
	Timestamp  string `json:"timestamp"`
	DurationMs *int64 `json:"durationMs,omitempty"`
}
type RuleMatchResult struct {
	Matched            bool          `json:"matched"`
	Rule               *RuleRecord   `json:"rule,omitempty"`
	Sound              *SoundRecord  `json:"sound,omitempty"`
	SoundPath          string        `json:"soundPath,omitempty"`
	MissingSound       bool          `json:"missingSound"`
	PlaybackAttempted  bool          `json:"playbackAttempted"`
	PlaybackStarted    bool          `json:"playbackStarted"`
	PlaybackError      string        `json:"playbackError,omitempty"`
	EventEngineEnabled bool          `json:"eventEngineEnabled"`
	PlaybackEnabled    bool          `json:"playbackEnabled"`
	Message            string        `json:"message"`
	Event              TerminalEvent `json:"event"`
}
type RecentEventRecord struct {
	ID                 string        `json:"id"`
	Event              TerminalEvent `json:"event"`
	Matched            bool          `json:"matched"`
	RuleID             string        `json:"ruleId,omitempty"`
	RuleName           string        `json:"ruleName,omitempty"`
	SoundID            string        `json:"soundId,omitempty"`
	SoundName          string        `json:"soundName,omitempty"`
	MissingSound       bool          `json:"missingSound"`
	PlaybackAttempted  bool          `json:"playbackAttempted"`
	PlaybackStarted    bool          `json:"playbackStarted"`
	PlaybackSkipped    bool          `json:"playbackSkipped"`
	PlaybackSkipReason string        `json:"playbackSkipReason,omitempty"`
	PlaybackError      string        `json:"playbackError,omitempty"`
	Message            string        `json:"message"`
	Timestamp          string        `json:"timestamp"`
}

type DependencySuggestion struct {
	Platform string `json:"platform"`
	Label    string `json:"label"`
	Command  string `json:"command,omitempty"`
	Detail   string `json:"detail"`
}
type PlaybackStatus struct {
	Supported    bool                   `json:"supported"`
	Platform     string                 `json:"platform"`
	Method       string                 `json:"method"`
	Playing      bool                   `json:"playing"`
	Message      string                 `json:"message"`
	Alternatives []string               `json:"alternatives"`
	Suggestions  []DependencySuggestion `json:"suggestions"`
}
