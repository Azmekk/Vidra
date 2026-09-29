//go:build !windows

package services

import (
	"os/exec"
	"syscall"
)

// killTree runs the command in its own process group so cancellation also
// stops children such as the ffmpeg process yt-dlp spawns for merging.
func killTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
