package main

import (
	"ProdTag/internal/integrations"
	"errors"
)

type CheckResult = integrations.CheckResult
type IntegrationStatus = integrations.IntegrationStatus
type DoctorResult = integrations.DoctorResult

func (a *App) ListIntegrationStatuses() ([]IntegrationStatus, error) {
	return integrations.ListStatuses()
}
func (a *App) InstallShellIntegration(shell string) (DoctorResult, error) {
	return integrations.Install(shell)
}
func (a *App) UninstallShellIntegration(shell string) (DoctorResult, error) {
	return integrations.Uninstall(shell)
}
func (a *App) RunIntegrationDoctor() (DoctorResult, error) { return integrations.Doctor() }

// Phase 4.3 bindings remain as compatibility shims for existing frontend state.
func (a *App) InstallZshIntegration() (DoctorResult, error)   { return integrations.Install("zsh") }
func (a *App) UninstallZshIntegration() (DoctorResult, error) { return integrations.Uninstall("zsh") }

const zshMarkerStart = integrations.MarkerStart
const zshMarkerEnd = integrations.MarkerEnd

func installZshrcBlock(path, block string) error { return integrations.InstallMarkedBlock(path, block) }
func removeZshrcBlock(path string) error         { return integrations.RemoveMarkedBlock(path) }
func detectZshrcMarkerState(path string) string  { return integrations.DetectMarkerState(path) }
func markerState(content string) string {
	if containsStart, containsEnd := contains(content, zshMarkerStart), contains(content, zshMarkerEnd); containsStart && containsEnd {
		return "configured"
	} else if containsStart || containsEnd {
		return "partial"
	}
	return "not_configured"
}
func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
func getZshIntegrationStatus() (IntegrationStatus, error) {
	statuses, err := integrations.ListStatuses()
	if err != nil {
		return IntegrationStatus{}, err
	}
	for _, status := range statuses {
		if status.Shell == "zsh" {
			return status, nil
		}
	}
	return IntegrationStatus{}, errors.New("zsh status unavailable")
}
func runDoctor() (DoctorResult, error) { return integrations.Doctor() }
