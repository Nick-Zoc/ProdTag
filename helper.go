package main

import (
	"ProdTag/internal/core"
	"ProdTag/internal/helpercli"
	"io"
)

const helperVersion = helpercli.Version

func runHelper(args []string, stdout, stderr io.Writer) int {
	return helpercli.Run(args, stdout, stderr, startPlayback)
}
func formatEmitResult(result RuleMatchResult) string { return helpercli.FormatEmitResult(result) }
func InferTerminalEventType(command string, exitCode *int) string {
	return core.InferTerminalEventType(command, exitCode)
}
