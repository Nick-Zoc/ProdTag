package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const helperVersion = "ProdTag helper Phase 4.2"

func runHelper(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printHelperUsage(stdout)
		return 0
	}

	switch args[0] {
	case "emit":
		return runHelperEmit(args[1:], stdout, stderr)
	case "doctor":
		return runHelperDoctor(stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, helperVersion)
		return 0
	case "help", "--help", "-h":
		printHelperUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printHelperUsage(stderr)
		return 2
	}
}

func runHelperEmit(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("emit", flag.ContinueOnError)
	flags.SetOutput(stderr)

	eventType := flags.String("event-type", "", "ProdTag event type")
	command := flags.String("command", "", "completed shell command")
	exitCodeValue := flags.String("exit-code", "", "command exit code")
	cwd := flags.String("cwd", "", "working directory")
	durationValue := flags.String("duration-ms", "", "command duration in milliseconds")
	timestamp := flags.String("timestamp", "", "event timestamp")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	exitCode, err := optionalInt(exitCodeValue)
	if err != nil {
		fmt.Fprintf(stderr, "invalid --exit-code: %v\n", err)
		return 2
	}
	durationMs, err := optionalInt64(durationValue)
	if err != nil {
		fmt.Fprintf(stderr, "invalid --duration-ms: %v\n", err)
		return 2
	}

	eventTypeValue := strings.TrimSpace(*eventType)
	if eventTypeValue == "" {
		eventTypeValue = InferTerminalEventType(*command, exitCode)
	}

	cwdValue := strings.TrimSpace(*cwd)
	if cwdValue == "" {
		if currentDir, err := os.Getwd(); err == nil {
			cwdValue = currentDir
		}
	}

	event := TerminalEvent{
		EventType:  eventTypeValue,
		Command:    strings.TrimSpace(*command),
		ExitCode:   exitCode,
		Cwd:        cwdValue,
		Timestamp:  strings.TrimSpace(*timestamp),
		DurationMs: durationMs,
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	result, err := NewApp().HandleTerminalEvent(event)
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper emit failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, formatEmitResult(result))
	return 0
}

func runHelperDoctor(stdout io.Writer, stderr io.Writer) int {
	snapshot, err := LoadConfigSnapshot()
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper doctor failed: %v\n", err)
		return 1
	}
	playbackStatus := getPlaybackStatus()
	logPath, err := handledEventLogPath()
	if err != nil {
		fmt.Fprintf(stderr, "prodtag-helper doctor failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, helperVersion)
	fmt.Fprintf(stdout, "platform: %s\n", runtime.GOOS)
	fmt.Fprintf(stdout, "config: %s\n", snapshot.Paths.ConfigFile)
	fmt.Fprintf(stdout, "logs: %s\n", logPath)
	fmt.Fprintf(stdout, "sounds: %d\n", len(snapshot.Config.Sounds))
	fmt.Fprintf(stdout, "rules: %d\n", len(snapshot.Config.Rules))
	fmt.Fprintf(stdout, "listening: %t\n", snapshot.Config.Listening)
	fmt.Fprintf(stdout, "event engine: %t\n", snapshot.Config.EventEngineEnabled)
	fmt.Fprintf(stdout, "muted: %t\n", snapshot.Config.Muted)
	fmt.Fprintf(stdout, "playback enabled: %t\n", snapshot.Config.PlaybackEnabled)
	fmt.Fprintf(stdout, "playback method: %s (%s)\n", playbackStatus.Method, playbackStatus.Message)
	return 0
}

func printHelperUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  prodtag-helper emit --event-type command_success --command \"ls -a\" --exit-code 0 --cwd \"$PWD\" --duration-ms 120")
	fmt.Fprintln(output, "  prodtag-helper doctor")
	fmt.Fprintln(output, "  prodtag-helper version")
}

func formatEmitResult(result RuleMatchResult) string {
	if !result.Matched {
		return fmt.Sprintf("no match: %s", result.Message)
	}
	if result.Rule == nil {
		return fmt.Sprintf("matched: %s", result.Message)
	}
	if result.MissingSound {
		return fmt.Sprintf("matched rule %q (%s); missing assigned sound", result.Rule.Name, result.Rule.ID)
	}

	soundName := "unknown sound"
	if result.Sound != nil {
		soundName = result.Sound.Name
	}

	playbackState := result.Message
	if result.PlaybackStarted {
		playbackState = "playback started"
	} else if result.PlaybackError != "" {
		playbackState = "playback error: " + result.PlaybackError
	}

	return fmt.Sprintf("matched rule %q (%s); sound %q; %s", result.Rule.Name, result.Rule.ID, soundName, playbackState)
}

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

func isTestCommand(command string) bool {
	return strings.HasPrefix(command, "npm test") ||
		strings.HasPrefix(command, "npm run test") ||
		strings.HasPrefix(command, "pnpm test") ||
		strings.HasPrefix(command, "pnpm run test") ||
		strings.HasPrefix(command, "yarn test") ||
		strings.HasPrefix(command, "pytest") ||
		strings.HasPrefix(command, "go test") ||
		strings.HasPrefix(command, "cargo test") ||
		strings.HasPrefix(command, "flutter test")
}

func isBuildCommand(command string) bool {
	return strings.HasPrefix(command, "npm run build") ||
		strings.HasPrefix(command, "pnpm build") ||
		strings.HasPrefix(command, "pnpm run build") ||
		strings.HasPrefix(command, "yarn build") ||
		strings.HasPrefix(command, "go build") ||
		strings.HasPrefix(command, "cargo build") ||
		strings.HasPrefix(command, "flutter build")
}

func optionalInt(value *string) (*int, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalInt64(value *string) (*int64, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(*value), 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
