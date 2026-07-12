package integrations

import (
	"ProdTag/internal/core"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const MarkerStart = "# >>> ProdTag >>>"
const MarkerEnd = "# <<< ProdTag <<<"

type CheckResult struct {
	OK     bool   `json:"ok"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Path   string `json:"path,omitempty"`
}
type IntegrationStatus struct {
	Shell                 string        `json:"shell"`
	DisplayName           string        `json:"displayName"`
	Supported             bool          `json:"supported"`
	PlatformRelevant      bool          `json:"platformRelevant"`
	ShellExecutableFound  bool          `json:"shellExecutableFound"`
	ShellExecutable       string        `json:"shellExecutable,omitempty"`
	HelperInstalled       bool          `json:"helperInstalled"`
	HelperExecutable      bool          `json:"helperExecutable"`
	ScriptInstalled       bool          `json:"scriptInstalled"`
	ProfileConfigured     bool          `json:"profileConfigured"`
	ProfileState          string        `json:"profileState"`
	ProfilePath           string        `json:"profilePath,omitempty"`
	HelperPath            string        `json:"helperPath"`
	ScriptPath            string        `json:"scriptPath"`
	MatcherCachePath      string        `json:"matcherCachePath"`
	CurrentSessionCommand string        `json:"currentSessionCommand"`
	DisableSessionCommand string        `json:"disableSessionCommand"`
	DebugEnableCommand    string        `json:"debugEnableCommand"`
	DebugDisableCommand   string        `json:"debugDisableCommand"`
	DebugLogCommand       string        `json:"debugLogCommand"`
	Problems              []string      `json:"problems"`
	Warnings              []string      `json:"warnings"`
	Checks                []CheckResult `json:"checks"`
}
type DoctorResult struct {
	OK                 bool                `json:"ok"`
	Platform           string              `json:"platform"`
	ConfigPath         string              `json:"configPath"`
	MatcherCacheValid  bool                `json:"matcherCacheValid"`
	RuleCount          int                 `json:"ruleCount"`
	Listening          bool                `json:"listening"`
	EventEngineEnabled bool                `json:"eventEngineEnabled"`
	Muted              bool                `json:"muted"`
	PlaybackEnabled    bool                `json:"playbackEnabled"`
	Playback           core.PlaybackStatus `json:"playback"`
	Integrations       []IntegrationStatus `json:"integrations"`
}

func ListStatuses() ([]IntegrationStatus, error) {
	paths, err := core.GetAppDataPaths()
	if err != nil {
		return nil, err
	}
	shells := []string{"zsh", "bash", "powershell"}
	statuses := make([]IntegrationStatus, 0, len(shells))
	for _, shell := range shells {
		status, err := statusFor(shell, paths)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	sort.SliceStable(statuses, func(i, j int) bool {
		if statuses[i].PlatformRelevant == statuses[j].PlatformRelevant {
			return i < j
		}
		return statuses[i].PlatformRelevant
	})
	return statuses, nil
}
func Doctor() (DoctorResult, error) {
	snapshot, err := core.LoadConfigSnapshot()
	if err != nil {
		return DoctorResult{}, err
	}
	statuses, err := ListStatuses()
	if err != nil {
		return DoctorResult{}, err
	}
	_, cacheErr := core.ReadMatcherCache(snapshot.Paths.MatcherCacheFile)
	playback := core.GetPlaybackStatus()
	ok := cacheErr == nil && playback.Supported
	configuredRelevant := 0
	for _, status := range statuses {
		if status.PlatformRelevant && status.ProfileConfigured {
			configuredRelevant++
			if !status.HelperExecutable || !status.ScriptInstalled || status.ProfileState != "configured" { ok = false }
		}
	}
	if configuredRelevant == 0 { ok = false }
	return DoctorResult{OK: ok, Platform: runtime.GOOS, ConfigPath: snapshot.Paths.ConfigFile, MatcherCacheValid: cacheErr == nil, RuleCount: len(snapshot.Config.Rules), Listening: snapshot.Config.Listening, EventEngineEnabled: snapshot.Config.EventEngineEnabled, Muted: snapshot.Config.Muted, PlaybackEnabled: snapshot.Config.PlaybackEnabled, Playback: playback, Integrations: statuses}, nil
}
func Install(shell string) (DoctorResult, error) {
	paths, err := core.GetAppDataPaths()
	if err != nil {
		return DoctorResult{}, err
	}
	status, err := statusFor(shell, paths)
	if err != nil {
		return DoctorResult{}, err
	}
	if !status.Supported {
		return Doctor()
	}
	if err := installAssets(shell, paths); err != nil {
		return DoctorResult{}, err
	}
	profile, err := profilePath(shell)
	if err != nil {
		return DoctorResult{}, err
	}
	if err := InstallMarkedBlock(profile, profileBlock(shell, paths)); err != nil {
		return DoctorResult{}, err
	}
	return Doctor()
}
func Uninstall(shell string) (DoctorResult, error) {
	profile, err := profilePath(shell)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Doctor()
		}
		return DoctorResult{}, err
	}
	if err := RemoveMarkedBlock(profile); err != nil {
		return DoctorResult{}, err
	}
	return Doctor()
}

func statusFor(shell string, paths core.AppDataPaths) (IntegrationStatus, error) {
	profile, profileErr := profilePath(shell)
	executable, execErr := shellExecutable(shell)
	script := scriptPath(shell, paths)
	helperInfo, helperErr := os.Stat(paths.HelperBinary)
	scriptInfo, scriptErr := os.Stat(script)
	state := "unavailable"
	if profileErr == nil {
		state = DetectMarkerState(profile)
	}
	supported := execErr == nil && profileErr == nil
	problems := []string{}
	warnings := []string{}
	if execErr != nil {
		problems = append(problems, shell+" executable was not found")
	}
	if state == "partial" {
		problems = append(problems, "Profile contains a partial ProdTag marker block")
	}
	if !playbackRelevant(shell) {
		warnings = append(warnings, "This shell is available but is not the primary integration for "+runtime.GOOS)
	}
	status := IntegrationStatus{Shell: shell, DisplayName: displayName(shell), Supported: supported, PlatformRelevant: playbackRelevant(shell), ShellExecutableFound: execErr == nil, ShellExecutable: executable, HelperInstalled: helperErr == nil && !helperInfo.IsDir(), HelperExecutable: helperErr == nil && helperInfo.Mode()&0o111 != 0, ScriptInstalled: scriptErr == nil && !scriptInfo.IsDir(), ProfileConfigured: state == "configured", ProfileState: state, ProfilePath: profile, HelperPath: paths.HelperBinary, ScriptPath: script, MatcherCachePath: paths.MatcherCacheFile, Problems: problems, Warnings: warnings}
	status.CurrentSessionCommand = currentSessionCommand(shell, paths)
	status.DisableSessionCommand = disableCommand(shell)
	status.DebugEnableCommand = debugEnableCommand(shell)
	status.DebugDisableCommand = debugDisableCommand(shell)
	status.DebugLogCommand = debugLogCommand(shell)
	status.Checks = []CheckResult{{OK: status.ShellExecutableFound, Label: status.DisplayName, Detail: availability(status.ShellExecutableFound), Path: executable}, {OK: status.HelperInstalled && status.HelperExecutable, Label: "Helper", Detail: helperDetail(status), Path: paths.HelperBinary}, {OK: status.ScriptInstalled, Label: "Integration script", Detail: availability(status.ScriptInstalled), Path: script}, {OK: status.ProfileConfigured, Label: "Profile", Detail: state, Path: profile}}
	return status, nil
}
func shellExecutable(shell string) (string, error) {
	names := map[string][]string{"zsh": {"zsh"}, "bash": {"bash"}, "powershell": {"pwsh", "powershell.exe", "powershell"}}[shell]
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", exec.ErrNotFound
}
func profilePath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".bash_profile"), nil
		}
		return filepath.Join(home, ".bashrc"), nil
	case "powershell":
		exe, err := shellExecutable(shell)
		if err != nil {
			return "", err
		}
		cmd := exec.Command(exe, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "$PROFILE.CurrentUserAllHosts")
		configureHidden(cmd)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("detect PowerShell profile: %w", err)
		}
		value := strings.TrimSpace(string(output))
		if value == "" {
			return "", errors.New("PowerShell returned an empty profile path")
		}
		return value, nil
	}
	return "", fmt.Errorf("unsupported shell %s", shell)
}
func scriptPath(shell string, paths core.AppDataPaths) string {
	switch shell {
	case "zsh":
		return paths.ZshScript
	case "bash":
		return paths.BashScript
	default:
		return paths.PowerShellScript
	}
}
func installAssets(shell string, paths core.AppDataPaths) error {
	if err := core.EnsureAppData(); err != nil {
		return err
	}
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	if err := copyFile(filepath.Join(root, "scripts", filepath.Base(scriptPath(shell, paths))), scriptPath(shell, paths), 0o755); err != nil {
		return err
	}
	tmp := paths.HelperBinary + ".tmp"
	cmd := exec.Command("go", "build", "-o", tmp, "./cmd/prodtag-helper")
	cmd.Dir = root
	configureHidden(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build helper: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, paths.HelperBinary)
}
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for current := cwd; ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(current, "scripts")); err == nil {
				return current, nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return "", errors.New("project assets unavailable; packaged releases must bundle helper and integration scripts")
}
func copyFile(source, destination string, mode os.FileMode) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	tmp := destination + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, destination)
}

func InstallMarkedBlock(path, block string) error {
	data, err := os.ReadFile(path)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	original := string(data)
	state := markerState(original)
	if state == "partial" {
		return errors.New("profile contains a partial ProdTag marker block")
	}
	if state == "configured" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if exists {
		backup := path + ".prodtag-backup"
		if _, err := os.Stat(backup); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(backup, data, 0o600); err != nil {
				return err
			}
		}
	}
	content := strings.TrimRight(original, "\r\n")
	if content != "" {
		content += "\n\n"
	}
	return os.WriteFile(path, []byte(content+block+"\n"), 0o600)
}
func RemoveMarkedBlock(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	content := string(data)
	if markerState(content) == "partial" {
		return errors.New("profile contains a partial ProdTag marker block")
	}
	start := strings.Index(content, MarkerStart)
	if start < 0 {
		return nil
	}
	relative := strings.Index(content[start:], MarkerEnd)
	end := start + relative + len(MarkerEnd)
	if end < len(content) && content[end] == '\r' {
		end++
	}
	if end < len(content) && content[end] == '\n' {
		end++
	}
	before := strings.TrimRight(content[:start], "\r\n")
	after := strings.TrimLeft(content[end:], "\r\n")
	result := before
	if result != "" && after != "" {
		result += "\n\n"
	}
	result += after
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return os.WriteFile(path, []byte(result), 0o600)
}
func DetectMarkerState(path string) string {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "missing"
	}
	if err != nil {
		return "unreadable"
	}
	return markerState(string(data))
}
func markerState(content string) string {
	start := strings.Contains(content, MarkerStart)
	end := strings.Contains(content, MarkerEnd)
	if start && end {
		return "configured"
	}
	if start || end {
		return "partial"
	}
	return "not_configured"
}
func profileBlock(shell string, paths core.AppDataPaths) string {
	lines := []string{MarkerStart}
	if shell == "powershell" {
		lines = append(lines, "$env:PRODTAG_HELPER = "+psQuote(paths.HelperBinary), "$env:PRODTAG_MATCHER_CACHE = "+psQuote(paths.MatcherCacheFile), "$env:PRODTAG_CONFIG = "+psQuote(paths.ConfigFile), ". "+psQuote(paths.PowerShellScript))
	} else {
		lines = append(lines, "export PRODTAG_HELPER="+shQuote(paths.HelperBinary), "export PRODTAG_MATCHER_CACHE="+shQuote(paths.MatcherCacheFile), "export PRODTAG_CONFIG="+shQuote(paths.ConfigFile), "source "+shQuote(scriptPath(shell, paths)))
	}
	return strings.Join(append(lines, MarkerEnd), "\n")
}
func currentSessionCommand(shell string, paths core.AppDataPaths) string {
	if shell == "powershell" {
		return "$env:PRODTAG_HELPER = " + psQuote(paths.HelperBinary) + "; $env:PRODTAG_MATCHER_CACHE = " + psQuote(paths.MatcherCacheFile) + "; $env:PRODTAG_CONFIG = " + psQuote(paths.ConfigFile) + "; . " + psQuote(paths.PowerShellScript)
	}
	return "export PRODTAG_HELPER=" + shQuote(paths.HelperBinary) + " PRODTAG_MATCHER_CACHE=" + shQuote(paths.MatcherCacheFile) + " PRODTAG_CONFIG=" + shQuote(paths.ConfigFile) + "; source " + shQuote(scriptPath(shell, paths))
}
func disableCommand(shell string) string {
	switch shell {
	case "zsh":
		return "export PRODTAG_ZSH_ENABLED=0"
	case "bash":
		return "export PRODTAG_BASH_ENABLED=0"
	default:
		return "$env:PRODTAG_POWERSHELL_ENABLED = '0'"
	}
}
func debugEnableCommand(shell string) string {
	switch shell {
	case "zsh":
		return "export PRODTAG_ZSH_DEBUG=1"
	case "bash":
		return "export PRODTAG_BASH_DEBUG=1"
	default:
		return "$env:PRODTAG_POWERSHELL_DEBUG = '1'"
	}
}
func debugDisableCommand(shell string) string {
	switch shell {
	case "zsh":
		return "export PRODTAG_ZSH_DEBUG=0"
	case "bash":
		return "export PRODTAG_BASH_DEBUG=0"
	default:
		return "$env:PRODTAG_POWERSHELL_DEBUG = '0'"
	}
}
func debugLogCommand(shell string) string {
	if shell == "powershell" {
		return "Get-Content (Join-Path ([System.IO.Path]::GetTempPath()) 'prodtag-powershell-debug.log') -Tail 50"
	}
	return "tail -n 50 \"${TMPDIR:-/tmp}/prodtag-" + shell + "-debug.log\""
}
func displayName(shell string) string {
	if shell == "powershell" {
		return "PowerShell"
	}
	return shell
}
func playbackRelevant(shell string) bool {
	switch runtime.GOOS {
	case "windows":
		return shell == "powershell"
	case "darwin":
		return shell == "zsh" || shell == "bash"
	case "linux":
		return shell == "bash" || shell == "zsh"
	}
	return false
}
func shQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
func psQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
func availability(ok bool) string {
	if ok {
		return "available"
	}
	return "missing"
}
func helperDetail(status IntegrationStatus) string {
	if !status.HelperInstalled {
		return "missing"
	}
	if !status.HelperExecutable {
		return "not executable"
	}
	return "installed and executable"
}
