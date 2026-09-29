package services

import (
	"os/exec"
	"strconv"
)

// killTree makes cancellation terminate the full process tree, since
// yt-dlp.exe is a launcher that spawns its own child processes.
func killTree(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		return exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	}
}
