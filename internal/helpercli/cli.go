package helpercli

import (
	"ProdTag/internal/core"
	"ProdTag/internal/integrations"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const Version = "ProdTag helper Phase 4.4"

func Run(args []string, stdout, stderr io.Writer, starter core.PlaybackStarter) int {
	if len(args) == 0 {
		usage(stdout)
		return 0
	}
	switch args[0] {
	case "emit":
		return emit(args[1:], stdout, stderr, starter)
	case "doctor":
		return doctor(stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, Version)
		return 0
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
}
func emit(args []string, stdout, stderr io.Writer, starter core.PlaybackStarter) int {
	flags := flag.NewFlagSet("emit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	eventType := flags.String("event-type", "", "ProdTag event type")
	command := flags.String("command", "", "completed shell command")
	exitValue := flags.String("exit-code", "", "command exit code")
	cwd := flags.String("cwd", "", "working directory")
	durationValue := flags.String("duration-ms", "", "command duration in milliseconds")
	timestamp := flags.String("timestamp", "", "event timestamp")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	exitCode, err := optionalInt(*exitValue)
	if err != nil {
		fmt.Fprintf(stderr, "invalid --exit-code: %v\n", err)
		return 2
	}
	duration, err := optionalInt64(*durationValue)
	if err != nil {
		fmt.Fprintf(stderr, "invalid --duration-ms: %v\n", err)
		return 2
	}
	cwdValue := strings.TrimSpace(*cwd)
	if cwdValue == "" {
		cwdValue, _ = os.Getwd()
	}
	event := core.TerminalEvent{EventType: strings.TrimSpace(*eventType), Command: strings.TrimSpace(*command), ExitCode: exitCode, Cwd: cwdValue, Timestamp: strings.TrimSpace(*timestamp), DurationMs: duration}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	snapshot, err := core.LoadConfigSnapshot()
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper emit failed: %v\n", err)
		return 1
	}
	result, err := core.HandleTerminalEvent(snapshot.Config, event, starter, core.AppendHandledEventLog)
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper emit failed: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, FormatEmitResult(result))
	return 0
}
func doctor(stdout, stderr io.Writer) int {
	result, err := integrations.Doctor()
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper doctor failed: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, Version)
	fmt.Fprintf(stdout, "overall: %s\n", healthState(result.OK))
	fmt.Fprintf(stdout, "platform: %s\n", result.Platform)
	fmt.Fprintf(stdout, "config: %s\n", result.ConfigPath)
	fmt.Fprintf(stdout, "matcher cache: %s\n", state(result.MatcherCacheValid))
	fmt.Fprintf(stdout, "rules: %d\n", result.RuleCount)
	fmt.Fprintf(stdout, "listening: %t\n", result.Listening)
	fmt.Fprintf(stdout, "event engine: %t\n", result.EventEngineEnabled)
	fmt.Fprintf(stdout, "muted: %t\n", result.Muted)
	fmt.Fprintf(stdout, "playback enabled: %t\n", result.PlaybackEnabled)
	fmt.Fprintf(stdout, "playback: %s (%s)\n", result.Playback.Method, result.Playback.Message)
	if len(result.Playback.Alternatives) > 0 {
		fmt.Fprintf(stdout, "playback alternatives: %s\n", strings.Join(result.Playback.Alternatives, ", "))
	}
	for _, suggestion := range result.Playback.Suggestions {
		if suggestion.Command != "" {
			fmt.Fprintf(stdout, "suggestion [%s]: %s\n", suggestion.Platform, suggestion.Command)
		}
	}
	for _, integration := range result.Integrations {
		fmt.Fprintf(stdout, "%s: executable=%t helper=%t script=%t profile=%s\n", integration.DisplayName, integration.ShellExecutableFound, integration.HelperExecutable, integration.ScriptInstalled, integration.ProfileState)
	}
	return 0
}
func FormatEmitResult(result core.RuleMatchResult) string {
	if !result.Matched {
		return "no match: " + result.Message
	}
	if result.Rule == nil {
		return "matched: " + result.Message
	}
	if result.MissingSound {
		return fmt.Sprintf("matched rule %q (%s); missing assigned sound", result.Rule.Name, result.Rule.ID)
	}
	sound := "unknown sound"
	if result.Sound != nil {
		sound = result.Sound.Name
	}
	playback := result.Message
	if result.PlaybackStarted {
		playback = "playback started"
	} else if result.PlaybackError != "" {
		playback = "playback error: " + result.PlaybackError
	}
	return fmt.Sprintf("matched rule %q (%s); sound %q; %s", result.Rule.Name, result.Rule.ID, sound, playback)
}
func optionalInt(value string) (*int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func optionalInt64(value string) (*int64, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func state(ok bool) string {
	if ok {
		return "valid"
	}
	return "invalid"
}
func healthState(ok bool) string {
	if ok {
		return "healthy"
	}
	return "needs attention"
}
func usage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  prodtag-helper emit --command \"ls -a\" --exit-code 0 --cwd \"$PWD\" --duration-ms 120")
	fmt.Fprintln(output, "  prodtag-helper doctor")
	fmt.Fprintln(output, "  prodtag-helper version")
}
