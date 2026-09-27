// Package pidfile records which process is the running one.
//
// A pid file answers a single question — is an instance already up — and
// answers it badly if it only checks whether the file exists, because a process
// killed with SIGKILL, or a machine that lost power, leaves the file behind.
// So reading one means reading the pid *and* asking whether that pid is still
// alive; a file naming a process that has gone is no claim at all and is
// overwritten.
//
// Alive is asked with signal 0, which POSIX defines as the existence check:
// error handling happens, no signal is delivered. That makes this package
// POSIX-only, which is what this is deployed on.
package pidfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// Read returns the pid held in the file at path if it names a process that is
// still running, and zero otherwise.
//
// A missing or unreadable file is an error; a file whose contents are not a
// number is not. The difference is deliberate: the first means the caller
// cannot tell what is going on, while the second means the claim is void, which
// is the same answer as a dead process and wants the same handling.
func Read(path string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(content)))
	if err != nil {
		return 0, nil
	}
	if alive(pid) {
		return pid, nil
	}

	return 0, nil
}

// Write claims path for pid, and refuses if the process already named there is
// running.
//
// The refusal is the point of the whole package: it is how a second server
// started by mistake stops instead of fighting the first one over a port.
func Write(path string, pid int) error {
	if pid < 1 {
		return fmt.Errorf("pidfile: %d is not a pid", pid)
	}

	switch running, err := Read(path); {
	case err != nil && !os.IsNotExist(err):
		return err
	case running != 0:
		return fmt.Errorf("pidfile: process %d is already running", running)
	}

	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}

// alive reports whether a process with this pid exists.
//
// Signal 0 delivers nothing and reports what would have happened. Permission
// denied is therefore a yes: the process is there, it simply belongs to someone
// else — and treating that as "gone" would let a second instance start.
//
// Only positive pids are asked about, because kill(2) reads zero and negative
// values as process groups and would answer about something else entirely.
func alive(pid int) bool {
	if pid < 1 {
		return false
	}
	err := syscall.Kill(pid, 0)

	return err == nil || errors.Is(err, syscall.EPERM)
}
