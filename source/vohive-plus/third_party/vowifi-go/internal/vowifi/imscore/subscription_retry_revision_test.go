package imscore

import (
	"testing"
	"time"
)

// Adapted from HiDeck 9f2cdefb (BFNOC/yibaiba) for the personal version's
// existing combined registration/retry store. No per-card reset API is added.
func TestPersonalOldServiceNotifyPreservesNewServiceBackoff(t *testing.T) {
	for _, mwi := range []bool{false, true} {
		for _, state := range []string{"active", "pending"} {
			t.Run(subscriptionEventPackage(mwi)+"/"+state, func(t *testing.T) {
				store := NewSubscriptionRegistrationStore()
				old := newSubscriptionLifecycleTestService(t, store)
				initial := beginProtocolSubscription(t, old, mwi)
				completeProtocolSubscription(t, old, mwi, initial)
				current := newSubscriptionLifecycleTestService(t, store)
				failSubscriptionByTimerN(t, current, mwi)
				before := store.subscriptionRetry(current.subscriptionBinding.identity, subscriptionEventPackage(mwi))
				if _, status := old.acceptSubscriptionNotification(notifyForSubscription(initial.request, state+";expires=120", 1)); status != 200 {
					t.Fatalf("old NOTIFY status = %d", status)
				}
				if after := store.subscriptionRetry(current.subscriptionBinding.identity, subscriptionEventPackage(mwi)); after != before {
					t.Fatalf("old NOTIFY changed another Service's retry: before=%+v after=%+v", before, after)
				}
				current.subscriptionFieldsLocked(mwi).lifecycle.retryAt = time.Time{}
				attempt := beginProtocolSubscription(t, current, mwi)
				if _, status := current.acceptSubscriptionNotification(notifyForSubscription(attempt.request, state+";expires=120", 1)); status != 200 {
					t.Fatalf("current NOTIFY status = %d", status)
				}
				if got := subscriptionFailuresForTest(current, mwi); got != 0 {
					t.Fatalf("current attempt left %d failures", got)
				}
			})
		}
	}
}

func TestPersonalQueuedRefreshPreservesOtherServiceBackoff(t *testing.T) {
	store := NewSubscriptionRegistrationStore()
	old := newSubscriptionLifecycleTestService(t, store)
	initial := beginProtocolSubscription(t, old, true)
	completeProtocolSubscription(t, old, true, initial)
	current := newSubscriptionLifecycleTestService(t, store)
	failSubscriptionByTimerN(t, current, true)
	before := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage)
	queued := buildProtocolSubscription(t, old, true)
	if err := old.recordSubscriptionUsageAttempt(queued, true); err != nil {
		t.Fatal(err)
	}
	if _, status := old.acceptSubscriptionNotification(notifyForSubscription(initial.request, "active;expires=120", 1)); status != 200 {
		t.Fatalf("old NOTIFY status = %d", status)
	}
	if after := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage); after != before {
		t.Fatalf("queued refresh confirmed another Service's failure: before=%+v after=%+v", before, after)
	}
}

func TestPersonalOldNotifyKeepsMWIStopped(t *testing.T) {
	for _, reason := range []string{"rejected", "noresource", "invariant"} {
		t.Run(reason, func(t *testing.T) {
			store := NewSubscriptionRegistrationStore()
			old := newSubscriptionLifecycleTestService(t, store)
			initial := beginProtocolSubscription(t, old, true)
			completeProtocolSubscription(t, old, true, initial)
			current := newSubscriptionLifecycleTestService(t, store)
			attempt := beginProtocolSubscription(t, current, true)
			if _, status := current.acceptSubscriptionNotification(notifyForSubscription(attempt.request, "terminated;reason="+reason, 1)); status != 200 {
				t.Fatalf("terminal NOTIFY status = %d", status)
			}
			before := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage)
			if before.stopped == "" {
				t.Fatal("fixture has no stop state")
			}
			if _, status := old.acceptSubscriptionNotification(notifyForSubscription(initial.request, "active;expires=120", 1)); status != 200 {
				t.Fatalf("old NOTIFY status = %d", status)
			}
			if after := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage); after.stopped != before.stopped {
				t.Fatalf("old NOTIFY removed stop: before=%+v after=%+v", before, after)
			}
			rebuilt := newSubscriptionLifecycleTestService(t, store)
			if start, _ := rebuilt.prepareSubscriptionStart(true); start {
				t.Fatal("old NOTIFY restarted stopped MWI")
			}
		})
	}
}

func TestPersonalRetryRevisionNotReusedAfterDeletion(t *testing.T) {
	store := NewSubscriptionRegistrationStore()
	old := newSubscriptionLifecycleTestService(t, store)
	failSubscriptionByTimerN(t, old, true)
	old.subscriptionFieldsLocked(true).lifecycle.retryAt = time.Time{}
	initial := beginProtocolSubscription(t, old, true)
	completeProtocolSubscription(t, old, true, initial)
	if _, status := old.acceptSubscriptionNotification(notifyForSubscription(initial.request, "active;expires=120", 1)); status != 200 {
		t.Fatal(status)
	}
	// Confirmation has deleted the old failure record. A subsequent failure
	// must get a store-wide new revision even if its count is again one.
	current := newSubscriptionLifecycleTestService(t, store)
	failSubscriptionByTimerN(t, current, true)
	before := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage)
	if _, status := old.acceptSubscriptionNotification(notifyForSubscription(initial.request, "active;expires=120", 2)); status != 200 {
		t.Fatal(status)
	}
	if after := store.subscriptionRetry(current.subscriptionBinding.identity, mwiEventPackage); after != before {
		t.Fatalf("deleted retry revision reused: before=%+v after=%+v", before, after)
	}
}
