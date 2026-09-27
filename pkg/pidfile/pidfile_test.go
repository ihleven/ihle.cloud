package pidfile_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ihleven/ihlvn/pkg/pidfile"
)

func tempfile(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "server.pid")
}

// Claiming a free path writes the pid, and reading it back gives it again.
func TestWriteThenRead(t *testing.T) {
	path := tempfile(t)

	if err := pidfile.Write(path, os.Getpid()); err != nil {
		t.Fatal(err)
	}

	got, err := pidfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != os.Getpid() {
		t.Errorf("read %d, want this process, %d", got, os.Getpid())
	}
}

// The refusal the package exists for: a second instance does not get the file
// while the first is up.
func TestWriteRefusesWhileTheOwnerLives(t *testing.T) {
	path := tempfile(t)

	if err := pidfile.Write(path, os.Getpid()); err != nil {
		t.Fatal(err)
	}
	if err := pidfile.Write(path, os.Getpid()); err == nil {
		t.Fatal("claimed a file already held by a running process")
	}
}

// A file left behind by a process that has since died is not a claim.
//
// The pid is a child's, waited for and therefore gone. It could in principle be
// reused by the time this asks, which is a race the kernel wins slowly enough
// that no test has lost it — and it is the same race the real thing runs.
func TestWriteTakesOverFromTheDead(t *testing.T) {
	path := tempfile(t)

	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not run a throwaway process: %v", err)
	}
	dead := cmd.Process.Pid

	if err := os.WriteFile(path, []byte(strconv.Itoa(dead)), 0o644); err != nil {
		t.Fatal(err)
	}

	if got, err := pidfile.Read(path); err != nil || got != 0 {
		t.Fatalf("Read of a dead pid = %d, %v; want 0 and no error", got, err)
	}
	if err := pidfile.Write(path, os.Getpid()); err != nil {
		t.Fatalf("refused to take over a stale file: %v", err)
	}
}

// Contents that are not a number are void rather than broken — the same answer
// as a dead process, because the file makes no claim either way.
func TestRubbishInTheFileIsNoClaim(t *testing.T) {
	path := tempfile(t)

	if err := os.WriteFile(path, []byte("wer weiss\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := pidfile.Read(path)
	if err != nil {
		t.Errorf("Read of a malformed file errored: %v", err)
	}
	if got != 0 {
		t.Errorf("Read = %d, want 0", got)
	}
	if err := pidfile.Write(path, os.Getpid()); err != nil {
		t.Errorf("refused to overwrite a malformed file: %v", err)
	}
}

// A file that is not there is an error, so a caller that cannot read its own
// directory hears about it instead of concluding the coast is clear.
func TestReadingWhatIsNotThere(t *testing.T) {
	if _, err := pidfile.Read(tempfile(t)); !os.IsNotExist(err) {
		t.Errorf("Read of a missing file gave %v, want a not-exist error", err)
	}
}

// Write is asked for a pid, and zero or a negative number is not one — kill(2)
// reads those as process groups, so letting them through would ask about
// something else.
func TestWriteRefusesSomethingThatIsNotAPid(t *testing.T) {
	for _, pid := range []int{0, -1, -1000} {
		if err := pidfile.Write(tempfile(t), pid); err == nil {
			t.Errorf("Write accepted %d as a pid", pid)
		}
	}
}
