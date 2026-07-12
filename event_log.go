package main

import "ProdTag/internal/core"

const handledEventLogFileName = core.HandledEventLogFileName
const handledEventLogLimit = core.HandledEventLogLimit
const handledEventLogMaxLines = core.HandledEventLogMaxLines

func (a *App) ListHandledEventLog() ([]RecentEventRecord, error) {
	return core.ReadHandledEventLog(handledEventLogLimit)
}
func appendHandledEventLog(record RecentEventRecord) error { return core.AppendHandledEventLog(record) }
func readHandledEventLog(limit int) ([]RecentEventRecord, error) {
	return core.ReadHandledEventLog(limit)
}
func handledEventLogPath() (string, error) { return core.HandledEventLogPath() }
