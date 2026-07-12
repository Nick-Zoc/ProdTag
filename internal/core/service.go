package core

type PlaybackStarter func(path string, stopPrevious bool) error
type EventAppender func(RecentEventRecord) error

func HandleTerminalEvent(config AppConfig, event TerminalEvent, start PlaybackStarter, appendEvent EventAppender) (RuleMatchResult, error) {
	event = NormalizeTerminalEvent(event)
	if !config.EventEngineEnabled {
		result := RuleMatchResult{EventEngineEnabled: false, PlaybackEnabled: config.PlaybackEnabled, Message: "Event engine is disabled", Event: event}
		return recordResult(result, appendEvent)
	}
	if !config.Listening {
		result := RuleMatchResult{EventEngineEnabled: true, PlaybackEnabled: config.PlaybackEnabled && !config.Muted, Message: "Listening is paused", Event: event}
		return recordResult(result, appendEvent)
	}
	result := EvaluateEvent(config, event)
	result.EventEngineEnabled = true
	result.PlaybackEnabled = config.PlaybackEnabled && !config.Muted
	if result.Matched && result.Sound != nil && !result.MissingSound {
		if !config.PlaybackEnabled {
			result.Message = "Matched rule; playback is disabled"
		} else if config.Muted {
			result.Message = "Matched rule; app is muted"
		} else {
			result.PlaybackAttempted = true
			if err := start(result.SoundPath, config.StopPreviousOnNewEvent); err != nil {
				result.PlaybackError = err.Error()
				result.Message = "Matched rule; playback failed"
			} else {
				result.PlaybackStarted = true
				result.Message = "Matched rule; playback started"
			}
		}
	}
	return recordResult(result, appendEvent)
}
func recordResult(result RuleMatchResult, appendEvent EventAppender) (RuleMatchResult, error) {
	if appendEvent == nil {
		return result, nil
	}
	if err := appendEvent(NewRecentEventRecord(result)); err != nil {
		return result, err
	}
	return result, nil
}
func HandleEventFromDisk(event TerminalEvent) (RuleMatchResult, error) {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return RuleMatchResult{}, err
	}
	return HandleTerminalEvent(snapshot.Config, event, StartPlayback, AppendHandledEventLog)
}
