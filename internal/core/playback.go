package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

type commandLookup func(string) (string, error)
type playbackChoice struct {
	Method       string
	Executable   string
	Alternatives []string
	Suggestions  []DependencySuggestion
}

var playbackState = struct {
	sync.Mutex
	cmd *exec.Cmd
}{}

func DetectPlayback(goos string, lookup commandLookup) PlaybackStatus {
	choice := selectPlayback(goos, lookup)
	status := PlaybackStatus{Supported: choice.Method != "", Platform: goos, Method: choice.Method, Alternatives: choice.Alternatives, Suggestions: choice.Suggestions}
	if status.Supported {
		status.Message = "Backend playback uses " + choice.Method + "."
	} else {
		status.Method = "unavailable"
		status.Message = missingPlaybackMessage(goos)
	}
	playbackState.Lock()
	status.Playing = playbackState.cmd != nil
	playbackState.Unlock()
	return status
}

func selectPlayback(goos string, lookup commandLookup) playbackChoice {
	candidates := []string{}
	switch goos {
	case "darwin":
		candidates = []string{"afplay"}
	case "windows":
		candidates = []string{"pwsh", "powershell"}
	case "linux":
		candidates = []string{"paplay", "aplay", "ffplay"}
	}
	found := []string{}
	paths := map[string]string{}
	for _, name := range candidates {
		if path, err := lookup(name); err == nil {
			found = append(found, name)
			paths[name] = path
		}
	}
	choice := playbackChoice{Alternatives: found, Suggestions: playbackSuggestions(goos)}
	if len(found) > 0 {
		choice.Method = found[0]
		choice.Executable = paths[found[0]]
	}
	return choice
}

func GetPlaybackStatus() PlaybackStatus { return DetectPlayback(runtime.GOOS, exec.LookPath) }

func StartPlayback(path string, stopPrevious bool) error {
	if path == "" {
		return errors.New("sound path is empty")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("sound file unavailable: %w", err)
	}
	choice := selectPlayback(runtime.GOOS, exec.LookPath)
	if choice.Method == "" {
		return errors.New(missingPlaybackMessage(runtime.GOOS))
	}
	if stopPrevious {
		if err := StopPlayback(); err != nil {
			return err
		}
	}
	cmd := playbackCommand(choice, path)
	configureBackgroundProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", choice.Method, err)
	}
	playbackState.Lock()
	playbackState.cmd = cmd
	playbackState.Unlock()
	go func() {
		_ = cmd.Wait()
		playbackState.Lock()
		if playbackState.cmd == cmd {
			playbackState.cmd = nil
		}
		playbackState.Unlock()
	}()
	return nil
}

func playbackCommand(choice playbackChoice, path string) *exec.Cmd {
	switch choice.Method {
	case "afplay", "paplay", "aplay":
		return exec.Command(choice.Executable, path)
	case "ffplay":
		return exec.Command(choice.Executable, "-nodisp", "-autoexit", "-loglevel", "quiet", path)
	case "pwsh", "powershell":
		script := "$p = New-Object System.Media.SoundPlayer; $p.SoundLocation = '" + escapePowerShell(path) + "'; $p.PlaySync()"
		return exec.Command(choice.Executable, "-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	default:
		return exec.Command(choice.Executable, path)
	}
}
func escapePowerShell(value string) string {
	result := ""
	for _, r := range value {
		if r == '\'' {
			result += "''"
		} else {
			result += string(r)
		}
	}
	return result
}
func StopPlayback() error {
	playbackState.Lock()
	cmd := playbackState.cmd
	playbackState.cmd = nil
	playbackState.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("stop playback: %w", err)
	}
	return nil
}
func playbackSuggestions(goos string) []DependencySuggestion {
	switch goos {
	case "darwin":
		return []DependencySuggestion{{Platform: "macOS", Label: "afplay", Detail: "afplay is included with macOS."}}
	case "windows":
		return []DependencySuggestion{{Platform: "Windows", Label: "PowerShell 7", Command: "winget install --id Microsoft.PowerShell --source winget", Detail: "Windows PowerShell is also used when available."}}
	case "linux":
		return []DependencySuggestion{{Platform: "Debian/Ubuntu", Label: "PulseAudio tools", Command: "sudo apt install pulseaudio-utils", Detail: "Provides paplay."}, {Platform: "Fedora", Label: "PulseAudio tools", Command: "sudo dnf install pulseaudio-utils", Detail: "Provides paplay."}, {Platform: "Arch", Label: "PulseAudio tools", Command: "sudo pacman -S libpulse", Detail: "Provides paplay."}, {Platform: "Linux fallback", Label: "ALSA or FFmpeg", Command: "sudo apt install alsa-utils ffmpeg", Detail: "Provides aplay or ffplay."}}
	}
	return nil
}
func missingPlaybackMessage(goos string) string {
	switch goos {
	case "windows":
		return "Playback unavailable: install PowerShell 7 or enable Windows PowerShell."
	case "linux":
		return "Playback unavailable: install paplay, aplay, or ffplay."
	case "darwin":
		return "Playback unavailable: macOS afplay was not found."
	default:
		return "Backend playback is unavailable on this platform."
	}
}
