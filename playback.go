package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

type PlaybackStatus struct {
	Supported bool   `json:"supported"`
	Platform  string `json:"platform"`
	Method    string `json:"method"`
	Playing   bool   `json:"playing"`
	Message   string `json:"message"`
}

var playback = struct {
	sync.Mutex
	cmd *exec.Cmd
}{
	cmd: nil,
}

var startPlayback = startNativePlayback
var stopCurrentPlayback = stopNativePlayback

func (a *App) GetPlaybackStatus() (PlaybackStatus, error) {
	return getPlaybackStatus(), nil
}

func (a *App) StopPlayback() (PlaybackStatus, error) {
	if err := stopCurrentPlayback(); err != nil {
		return getPlaybackStatus(), err
	}
	return getPlaybackStatus(), nil
}

func getPlaybackStatus() PlaybackStatus {
	status := PlaybackStatus{
		Supported: runtime.GOOS == "darwin",
		Platform:  runtime.GOOS,
		Method:    playbackMethod(),
		Message:   playbackStatusMessage(),
	}

	playback.Lock()
	status.Playing = playback.cmd != nil
	playback.Unlock()

	return status
}

func startNativePlayback(path string, stopPrevious bool) error {
	if path == "" {
		return errors.New("sound path is empty")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("sound file unavailable: %w", err)
	}
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("backend playback is not implemented for %s yet", runtime.GOOS)
	}
	if stopPrevious {
		if err := stopNativePlayback(); err != nil {
			return err
		}
	}

	cmd := exec.Command("afplay", path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start afplay: %w", err)
	}

	playback.Lock()
	playback.cmd = cmd
	playback.Unlock()

	go func() {
		_ = cmd.Wait()
		playback.Lock()
		if playback.cmd == cmd {
			playback.cmd = nil
		}
		playback.Unlock()
	}()

	return nil
}

func stopNativePlayback() error {
	playback.Lock()
	cmd := playback.cmd
	playback.cmd = nil
	playback.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("stop playback: %w", err)
	}
	return nil
}

func playbackMethod() string {
	if runtime.GOOS == "darwin" {
		return "afplay"
	}
	return "not implemented"
}

func playbackStatusMessage() string {
	if runtime.GOOS == "darwin" {
		return "Backend playback uses macOS afplay."
	}
	return "Backend playback is structured for this platform but not implemented yet."
}
