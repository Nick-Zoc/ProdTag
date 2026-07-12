package main

import "ProdTag/internal/core"

const configVersion = core.ConfigVersion

type AppDataPaths = core.AppDataPaths
type ConfigSnapshot = core.ConfigSnapshot
type AppConfig = core.AppConfig
type SoundRecord = core.SoundRecord
type PlaylistRecord = core.PlaylistRecord
type RuleRecord = core.RuleRecord
type HotkeySettings = core.HotkeySettings
type IntegrationSettings = core.IntegrationSettings
type ShellIntegrationState = core.ShellIntegrationState
type MatcherCache = core.MatcherCache
type MatcherCacheCandidate = core.MatcherCacheCandidate
type TerminalEvent = core.TerminalEvent
type RuleMatchResult = core.RuleMatchResult
type RecentEventRecord = core.RecentEventRecord
type DependencySuggestion = core.DependencySuggestion

func (a *App) LoadConfig() (ConfigSnapshot, error)                 { return core.LoadConfigSnapshot() }
func (a *App) SaveConfig(config AppConfig) (ConfigSnapshot, error) { return core.SaveConfig(config) }
func LoadConfigSnapshot() (ConfigSnapshot, error)                  { return core.LoadConfigSnapshot() }
func EnsureAppData() error                                         { return core.EnsureAppData() }
func GetAppDataPaths() (AppDataPaths, error)                       { return core.GetAppDataPaths() }
func normalizeConfig(config AppConfig) AppConfig                   { return core.NormalizeConfig(config) }
func defaultConfig() AppConfig                                     { return core.DefaultConfig() }
func defaultMatcherCache() MatcherCache                            { return core.DefaultMatcherCache() }
func readConfig(path string) (AppConfig, error)                    { return core.ReadConfig(path) }
func writeJSON(path string, value any) error                       { return core.WriteJSON(path, value) }
