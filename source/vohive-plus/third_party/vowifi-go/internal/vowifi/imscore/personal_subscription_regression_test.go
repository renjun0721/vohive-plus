package imscore

import (
	"testing"
	"time"
)

// Exercise Timer N through the existing lifecycle on both versions.
func TestPersonalSubscriptionFailureDelayIndependentOfExpires(t *testing.T) {
	for _, mwi := range []bool{false, true} {
		s := newSubscriptionLifecycleTestService(t, nil)
		result := beginProtocolSubscription(t, s, mwi)
		completeProtocolSubscription(t, s, mwi, result)
		f := s.subscriptionFieldsLocked(mwi)
		deadline := f.lifecycle.notifyDeadline
		s.mu.Lock()
		s.expireSubscriptionTimersLocked(deadline)
		s.mu.Unlock()
		want := 30 * time.Second
		if mwi {
			want = 5 * time.Minute
		}
		if got := f.lifecycle.retryAt.Sub(deadline); got != want {
			t.Errorf("mwi=%t retry=%s, want %s independent of requested Expires", mwi, got, want)
		}
	}
}
