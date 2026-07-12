//go:build !windows

package integrations

import "os/exec"

func configureHidden(cmd *exec.Cmd) {}
