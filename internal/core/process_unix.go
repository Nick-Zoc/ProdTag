//go:build !windows

package core

import "os/exec"

func configureBackgroundProcess(cmd *exec.Cmd) {}
