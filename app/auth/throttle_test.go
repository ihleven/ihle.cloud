package auth

import (
	"testing"
	"time"
)

func TestThrottleBlocksAfterRepeatedFailures(t *testing.T) {
	tr := newThrottle(3, time.Minute)

	if tr.blocked("matt") {
		t.Fatal("blocked before any failure")
	}
	tr.fail("matt", "10.0.0.1")
	tr.fail("matt", "10.0.0.1")
	if tr.blocked("matt") {
		t.Error("blocked below the limit")
	}
	tr.fail("matt", "10.0.0.1")
	if !tr.blocked("matt") {
		t.Error("not blocked at the limit")
	}
	// The address is limited too, so one client cannot sweep other auth.
	if !tr.blocked("10.0.0.1") {
		t.Error("the client address was not limited alongside the account")
	}
	if tr.blocked("wolfgang") {
		t.Error("an untouched account was blocked")
	}

	tr.succeed("matt", "10.0.0.1")
	if tr.blocked("matt") || tr.blocked("10.0.0.1") {
		t.Error("a successful sign-in did not clear the failures")
	}
}

func TestThrottleForgetsAfterTheWindow(t *testing.T) {
	tr := newThrottle(1, time.Millisecond)
	tr.fail("matt")
	if !tr.blocked("matt") {
		t.Fatal("not blocked immediately after the limit")
	}

	time.Sleep(5 * time.Millisecond)
	if tr.blocked("matt") {
		t.Error("still blocked after the window passed")
	}

	tr.fail("gone")
	time.Sleep(5 * time.Millisecond)
	tr.sweep()
	if len(tr.failures) != 0 {
		t.Errorf("sweep left %d expired entries", len(tr.failures))
	}
}
