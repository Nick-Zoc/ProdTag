package main

import "ProdTag/internal/core"

type PlaybackStatus = core.PlaybackStatus

var startPlayback = core.StartPlayback
var stopCurrentPlayback = core.StopPlayback

func (a *App) GetPlaybackStatus() (PlaybackStatus, error) { return getPlaybackStatus(), nil }
func (a *App) StopPlayback() (PlaybackStatus, error) {
	if err := stopCurrentPlayback(); err != nil {
		return getPlaybackStatus(), err
	}
	return getPlaybackStatus(), nil
}
func getPlaybackStatus() PlaybackStatus { return core.GetPlaybackStatus() }
func startNativePlayback(path string, stopPrevious bool) error {
	return core.StartPlayback(path, stopPrevious)
}
func stopNativePlayback() error { return core.StopPlayback() }
