//go:build windows

package main

import (
	"os"
	"os/exec"
)

func setProcessGroup(cmd *exec.Cmd) {}

func killGroup(p *os.Process) {}
