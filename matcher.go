package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

const recentEventLimit = 20

var recentEvents = struct {
	sync.Mutex
	items []RecentEventRecord
}{}

func (a *App) EvaluateEvent(event TerminalEvent) (RuleMatchResult, error) {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return RuleMatchResult{}, err
	}
	return evaluateEvent(snapshot.Config, normalizeTerminalEvent(event)), nil
}

func (a *App) SimulateEvent(event TerminalEvent) (RuleMatchResult, error) {
	result, err := a.EvaluateEvent(event)
	if err != nil {
		return RuleMatchResult{}, err
	}
	addRecentEvent(result)
	return result, nil
}

func (a *App) HandleTerminalEvent(event TerminalEvent) (RuleMatchResult, error) {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		return RuleMatchResult{}, err
	}

	event = normalizeTerminalEvent(event)
	if !snapshot.Config.EventEngineEnabled {
		result := RuleMatchResult{
			Matched:            false,
			EventEngineEnabled: false,
			PlaybackEnabled:    snapshot.Config.PlaybackEnabled,
			Message:            "Event engine is disabled",
			Event:              event,
		}
		if err := recordHandledEvent(result); err != nil {
			return result, err
		}
		return result, nil
	}
	if !snapshot.Config.Listening {
		result := RuleMatchResult{
			Matched:            false,
			EventEngineEnabled: true,
			PlaybackEnabled:    snapshot.Config.PlaybackEnabled && !snapshot.Config.Muted,
			Message:            "Listening is paused",
			Event:              event,
		}
		if err := recordHandledEvent(result); err != nil {
			return result, err
		}
		return result, nil
	}

	result := evaluateEvent(snapshot.Config, event)
	result.EventEngineEnabled = true
	result.PlaybackEnabled = snapshot.Config.PlaybackEnabled && !snapshot.Config.Muted

	if result.Matched && result.Sound != nil && !result.MissingSound {
		if !snapshot.Config.PlaybackEnabled {
			result.Message = "Matched rule; playback is disabled"
		} else if snapshot.Config.Muted {
			result.Message = "Matched rule; app is muted"
		} else {
			result.PlaybackAttempted = true
			if err := startPlayback(result.SoundPath, snapshot.Config.StopPreviousOnNewEvent); err != nil {
				result.PlaybackError = err.Error()
				result.Message = "Matched rule; playback failed"
			} else {
				result.PlaybackStarted = true
				result.Message = "Matched rule; playback started"
			}
		}
	}

	if err := recordHandledEvent(result); err != nil {
		return result, err
	}
	return result, nil
}

func (a *App) ListRecentEvents() ([]RecentEventRecord, error) {
	recentEvents.Lock()
	defer recentEvents.Unlock()

	events := make([]RecentEventRecord, len(recentEvents.items))
	copy(events, recentEvents.items)
	return events, nil
}

func (a *App) ClearRecentEvents() ([]RecentEventRecord, error) {
	recentEvents.Lock()
	defer recentEvents.Unlock()

	recentEvents.items = []RecentEventRecord{}
	return []RecentEventRecord{}, nil
}

func evaluateEvent(config AppConfig, event TerminalEvent) RuleMatchResult {
	bestIndex := -1
	bestScore := -1

	for index, rule := range config.Rules {
		score, ok := ruleMatchScore(rule, event)
		if !ok {
			continue
		}
		if score > bestScore {
			bestScore = score
			bestIndex = index
		}
	}

	if bestIndex == -1 {
		return RuleMatchResult{
			Matched: false,
			Message: "No matching rule",
			Event:   event,
		}
	}

	rule := config.Rules[bestIndex]
	soundIndex := findSoundIndex(config.Sounds, rule.SoundID)
	if soundIndex == -1 {
		return RuleMatchResult{
			Matched:      true,
			Rule:         &rule,
			MissingSound: true,
			Message:      "Matched rule, but assigned sound is missing",
			Event:        event,
		}
	}

	sound := config.Sounds[soundIndex]
	return RuleMatchResult{
		Matched:   true,
		Rule:      &rule,
		Sound:     &sound,
		SoundPath: previewPath(sound),
		Message:   "Matched rule",
		Event:     event,
	}
}

func ruleMatchScore(rule RuleRecord, event TerminalEvent) (int, bool) {
	if !rule.Enabled {
		return 0, false
	}
	if strings.TrimSpace(rule.EventType) != strings.TrimSpace(event.EventType) {
		return 0, false
	}
	if rule.ExitCode != nil {
		if event.ExitCode == nil || *event.ExitCode != *rule.ExitCode {
			return 0, false
		}
	}

	commandScore, ok := commandMatchScore(rule.MatchMode, rule.CommandPattern, event.Command)
	if !ok {
		return 0, false
	}

	score := commandScore
	if rule.ExitCode != nil {
		score += 5
	}
	return score, true
}

func commandMatchScore(mode string, pattern string, command string) (int, bool) {
	mode = strings.TrimSpace(mode)
	pattern = strings.TrimSpace(pattern)
	command = strings.TrimSpace(command)
	if mode == "" {
		mode = "any"
	}

	if pattern == "" || mode == "any" {
		return 0, true
	}
	if command == "" {
		return 0, false
	}

	switch mode {
	case "exact":
		return 40, command == pattern
	case "regex":
		matched, err := regexp.MatchString(pattern, command)
		if err != nil {
			return 0, false
		}
		return 40, matched
	case "startsWith":
		return 30, strings.HasPrefix(command, pattern)
	case "endsWith":
		return 30, strings.HasSuffix(command, pattern)
	case "contains":
		return 30, strings.Contains(command, pattern)
	default:
		return 0, false
	}
}

func normalizeTerminalEvent(event TerminalEvent) TerminalEvent {
	event.EventType = strings.TrimSpace(event.EventType)
	event.Command = strings.TrimSpace(event.Command)
	event.Cwd = strings.TrimSpace(event.Cwd)
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	return event
}

func recordHandledEvent(result RuleMatchResult) error {
	record := addRecentEvent(result)
	return appendHandledEventLog(record)
}

func addRecentEvent(result RuleMatchResult) RecentEventRecord {
	record := newRecentEventRecord(result)

	recentEvents.Lock()
	defer recentEvents.Unlock()

	recentEvents.items = append([]RecentEventRecord{record}, recentEvents.items...)
	if len(recentEvents.items) > recentEventLimit {
		recentEvents.items = recentEvents.items[:recentEventLimit]
	}

	return record
}

func newRecentEventRecord(result RuleMatchResult) RecentEventRecord {
	id, err := newSoundID()
	if err != nil {
		id = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}

	record := RecentEventRecord{
		ID:                id,
		Event:             result.Event,
		Matched:           result.Matched,
		MissingSound:      result.MissingSound,
		PlaybackAttempted: result.PlaybackAttempted,
		PlaybackStarted:   result.PlaybackStarted,
		PlaybackError:     result.PlaybackError,
		Message:           result.Message,
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
	}
	if result.Rule != nil {
		record.RuleID = result.Rule.ID
		record.RuleName = result.Rule.Name
	}
	if result.Sound != nil {
		record.SoundID = result.Sound.ID
		record.SoundName = result.Sound.Name
	}

	if !record.PlaybackStarted && record.PlaybackError == "" {
		record.PlaybackSkipped = true
		record.PlaybackSkipReason = result.Message
	}

	return record
}
