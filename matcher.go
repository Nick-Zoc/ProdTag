package main

import (
	"ProdTag/internal/core"
	"sync"
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
	return evaluateEvent(snapshot.Config, event), nil
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
	var recorded RecentEventRecord
	result, err := core.HandleTerminalEvent(snapshot.Config, event, startPlayback, func(record RecentEventRecord) error { recorded = record; return appendHandledEventLog(record) })
	if recorded.ID != "" {
		addRecentEventRecord(recorded)
	}
	return result, err
}
func (a *App) ListRecentEvents() ([]RecentEventRecord, error) {
	recentEvents.Lock()
	defer recentEvents.Unlock()
	items := make([]RecentEventRecord, len(recentEvents.items))
	copy(items, recentEvents.items)
	return items, nil
}
func (a *App) ClearRecentEvents() ([]RecentEventRecord, error) {
	recentEvents.Lock()
	defer recentEvents.Unlock()
	recentEvents.items = []RecentEventRecord{}
	return []RecentEventRecord{}, nil
}
func evaluateEvent(config AppConfig, event TerminalEvent) RuleMatchResult {
	return core.EvaluateEvent(config, event)
}
func normalizeTerminalEvent(event TerminalEvent) TerminalEvent {
	return core.NormalizeTerminalEvent(event)
}
func addRecentEvent(result RuleMatchResult) RecentEventRecord {
	record := core.NewRecentEventRecord(result)
	addRecentEventRecord(record)
	return record
}
func addRecentEventRecord(record RecentEventRecord) {
	recentEvents.Lock()
	defer recentEvents.Unlock()
	recentEvents.items = append([]RecentEventRecord{record}, recentEvents.items...)
	if len(recentEvents.items) > recentEventLimit {
		recentEvents.items = recentEvents.items[:recentEventLimit]
	}
}
func newRecentEventRecord(result RuleMatchResult) RecentEventRecord {
	return core.NewRecentEventRecord(result)
}
