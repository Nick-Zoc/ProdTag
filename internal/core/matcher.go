package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func InferTerminalEventType(command string, exitCode *int) string {
	command = strings.ToLower(strings.TrimSpace(command))
	suffix := "success"
	if exitCode != nil && *exitCode != 0 {
		suffix = "failure"
	}
	switch {
	case strings.HasPrefix(command, "git commit"):
		return "git_commit_" + suffix
	case strings.HasPrefix(command, "git push"):
		return "git_push_" + suffix
	case isTestCommand(command):
		return "test_" + suffix
	case isBuildCommand(command):
		return "build_" + suffix
	default:
		return "command_" + suffix
	}
}
func isTestCommand(c string) bool {
	for _, p := range []string{"npm test", "npm run test", "pnpm test", "pnpm run test", "yarn test", "pytest", "go test", "cargo test", "flutter test"} {
		if strings.HasPrefix(c, p) {
			return true
		}
	}
	return false
}
func isBuildCommand(c string) bool {
	for _, p := range []string{"npm run build", "pnpm build", "pnpm run build", "yarn build", "go build", "cargo build", "flutter build"} {
		if strings.HasPrefix(c, p) {
			return true
		}
	}
	return false
}
func NormalizeTerminalEvent(e TerminalEvent) TerminalEvent {
	e.EventType = strings.TrimSpace(e.EventType)
	e.Command = strings.TrimSpace(e.Command)
	e.Cwd = strings.TrimSpace(e.Cwd)
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if e.EventType == "" {
		e.EventType = InferTerminalEventType(e.Command, e.ExitCode)
	}
	return e
}
func EvaluateEvent(config AppConfig, event TerminalEvent) RuleMatchResult {
	event = NormalizeTerminalEvent(event)
	best, bestScore := -1, -1
	for i, rule := range config.Rules {
		score, ok := ruleMatchScore(rule, event)
		if ok && score > bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 {
		return RuleMatchResult{Message: "No matching rule", Event: event}
	}
	rule := config.Rules[best]
	soundIndex := -1
	for i := range config.Sounds {
		if config.Sounds[i].ID == rule.SoundID {
			soundIndex = i
			break
		}
	}
	if soundIndex < 0 {
		return RuleMatchResult{Matched: true, Rule: &rule, MissingSound: true, Message: "Matched rule, but assigned sound is missing", Event: event}
	}
	sound := config.Sounds[soundIndex]
	path := sound.OriginalPath
	if sound.ProcessedPath != nil && *sound.ProcessedPath != "" {
		path = *sound.ProcessedPath
	}
	return RuleMatchResult{Matched: true, Rule: &rule, Sound: &sound, SoundPath: path, Message: "Matched rule", Event: event}
}
func ruleMatchScore(rule RuleRecord, event TerminalEvent) (int, bool) {
	if !rule.Enabled || strings.TrimSpace(rule.EventType) != strings.TrimSpace(event.EventType) {
		return 0, false
	}
	if rule.ExitCode != nil && (event.ExitCode == nil || *event.ExitCode != *rule.ExitCode) {
		return 0, false
	}
	score, ok := commandMatchScore(rule.MatchMode, rule.CommandPattern, event.Command)
	if rule.ExitCode != nil {
		score += 5
	}
	return score, ok
}
func commandMatchScore(mode, pattern, command string) (int, bool) {
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
		ok, err := regexp.MatchString(pattern, command)
		return 40, err == nil && ok
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
func NewRecentEventRecord(result RuleMatchResult) RecentEventRecord {
	id := fmt.Sprintf("event-%d", time.Now().UnixNano())
	r := RecentEventRecord{ID: id, Event: result.Event, Matched: result.Matched, MissingSound: result.MissingSound, PlaybackAttempted: result.PlaybackAttempted, PlaybackStarted: result.PlaybackStarted, PlaybackError: result.PlaybackError, Message: result.Message, Timestamp: time.Now().UTC().Format(time.RFC3339Nano)}
	if result.Rule != nil {
		r.RuleID = result.Rule.ID
		r.RuleName = result.Rule.Name
	}
	if result.Sound != nil {
		r.SoundID = result.Sound.ID
		r.SoundName = result.Sound.Name
	}
	if !r.PlaybackStarted && r.PlaybackError == "" {
		r.PlaybackSkipped = true
		r.PlaybackSkipReason = result.Message
	}
	return r
}
