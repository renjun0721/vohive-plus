package device

import (
	"testing"
	"time"

	"github.com/yibaiba/hideck/internal/config"
)

func replaceWorkerForLifecycleTest(p *Pool, worker *Worker) {
	p.assignWorkerGeneration(worker)
	p.mu.Lock()
	p.workers[worker.ID] = worker
	p.mu.Unlock()
}

func TestAbandonDeviceCancelsWorkerLifecycleTimers(t *testing.T) {
	p := NewPool(&config.Config{})
	deviceID := "dev"
	p.AttachWorkerForTest(&Worker{ID: deviceID})
	owner := p.captureWorkerLifecycleOwner(deviceID)
	simTimer := time.AfterFunc(time.Hour, func() {})
	wakeupTimer := time.AfterFunc(time.Hour, func() {})
	t.Cleanup(func() {
		simTimer.Stop()
		wakeupTimer.Stop()
	})

	p.simEventTimers[deviceID] = simTimer
	p.deviceEventWakeups[deviceID] = &deviceEventRecoverWakeup{
		timer: wakeupTimer, owner: owner, sources: map[string]struct{}{"test": {}},
	}
	p.AbandonDevice(deviceID)
	if p.simEventTimers[deviceID] != nil || p.deviceEventWakeups[deviceID] != nil {
		t.Fatal("abandoned device retained delayed lifecycle callbacks")
	}
}

func TestStaleDeviceEventCannotFlushReplacementWakeup(t *testing.T) {
	p := NewPool(&config.Config{})
	old := &Worker{ID: "dev"}
	p.AttachWorkerForTest(old)
	stale := &deviceEventRecoverWakeup{
		owner: p.captureWorkerLifecycleOwner(old.ID), sources: map[string]struct{}{"old": {}},
	}
	replacement := &Worker{ID: old.ID}
	replaceWorkerForLifecycleTest(p, replacement)
	current := &deviceEventRecoverWakeup{
		owner: p.captureWorkerLifecycleOwner(replacement.ID), sources: map[string]struct{}{"new": {}},
	}
	p.deviceEventWakeups[old.ID] = current

	p.flushDesiredVoWiFiRecoverFromDeviceEvent(old.ID, stale)
	if p.deviceEventWakeups[old.ID] != current {
		t.Fatal("stale device event consumed the replacement wakeup")
	}
}

func TestVoWiFiReadyStartRejectsReplacementWorker(t *testing.T) {
	p := NewPool(&config.Config{})
	old := &Worker{ID: "dev"}
	p.AttachWorkerForTest(old)
	start := p.captureVoWiFiReadyStart(old)

	replacement := &Worker{ID: old.ID}
	replaceWorkerForLifecycleTest(p, replacement)
	p.voWiFiHost().InvalidateRuntime(old.ID, "worker_replaced")
	if err := p.enableVoWiFiWhenReady(start, time.Millisecond, "test"); err == nil {
		t.Fatal("stale automatic VoWiFi start accepted the replacement worker")
	}
}

func TestStaleBootstrapGenerationCannotReplacePublishedWorkerGeneration(t *testing.T) {
	p := NewPool(&config.Config{})
	current := &Worker{ID: "dev"}
	p.assignWorkerGeneration(current)
	if err := p.registerWorkerStarting(current); err != nil {
		t.Fatal(err)
	}

	stale := &Worker{ID: current.ID}
	p.assignWorkerGeneration(stale)
	p.transportRecovery.mu.Lock()
	published := p.transportRecovery.workerGenerations[current.ID]
	p.transportRecovery.mu.Unlock()
	if published != current.generation {
		t.Fatalf("stale generation %d replaced published generation %d", published, current.generation)
	}
}

func TestRemoveWorkerAllowsPendingStartupToPublish(t *testing.T) {
	p := NewPool(&config.Config{})
	t.Cleanup(p.cancel)
	p.ForceRebuildingForTest("dev")
	done := make(chan error, 1)
	go func() { done <- p.RemoveWorker("dev") }()
	// Allow removal to enter its wait for the in-progress initialization.
	time.Sleep(30 * time.Millisecond)
	started := make(chan error, 1)
	go func() {
		err := p.registerWorkerStarting(&Worker{ID: "dev"})
		p.mu.Lock()
		delete(p.rebuilding, "dev")
		p.mu.Unlock()
		started <- err
	}()
	select {
	case err := <-started:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("pending startup could not acquire the lifecycle lock")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("removal did not finish after startup settled")
	}
}

func TestStaleSIMAndWakeEventsDoNotTargetReplacementWorker(t *testing.T) {
	p := NewPool(&config.Config{})
	old := &Worker{ID: "dev"}
	p.AttachWorkerForTest(old)
	owner := p.captureWorkerLifecycleOwner(old.ID)
	replaceWorkerForLifecycleTest(p, &Worker{ID: old.ID})

	p.handleSIMStatusEventForOwner(owner, "stale", nil, "")
	p.wakeDesiredVoWiFiRecoverForOwner(owner, "stale")
	if p.simEventTimers[old.ID] != nil {
		t.Fatal("stale SIM event scheduled a timer for the replacement worker")
	}
	if p.deviceEventWakeups[old.ID] != nil {
		t.Fatal("stale device event scheduled VoWiFi recovery for the replacement worker")
	}
}

func TestStaleATDisconnectDoesNotStartReplacementRecovery(t *testing.T) {
	p := NewPool(&config.Config{})
	old := &Worker{ID: "dev"}
	p.AttachWorkerForTest(old)
	owner := p.captureWorkerLifecycleOwner(old.ID)
	replaceWorkerForLifecycleTest(p, &Worker{ID: old.ID})

	p.scheduleATDisconnectRecovery(owner, "serial_runtime_error")
	p.mu.RLock()
	recovering := p.modemRebootRecovering[old.ID]
	p.mu.RUnlock()
	if recovering {
		t.Fatal("stale AT disconnect started recovery for the replacement worker")
	}
}

func TestInFlightIPRefreshDoesNotApplyToReplacementWorker(t *testing.T) {
	p := NewPool(&config.Config{})
	started := make(chan struct{})
	release := make(chan struct{})
	network := &fakeController{connected: true, publicIPStarted: started, publicIPRelease: release}
	old := &Worker{ID: "dev", netOverride: network}
	old.cachedIP = "198.51.100.1"
	p.AttachWorkerForTest(old)

	p.refreshIPs(old, true)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("public IP refresh did not start")
	}
	replaceWorkerForLifecycleTest(p, &Worker{ID: old.ID})
	close(release)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		old.ipRefreshMu.Lock()
		inFlight := old.ipRefreshInFlight
		old.ipRefreshMu.Unlock()
		if !inFlight {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if old.GetCachedIP() != "198.51.100.1" {
		t.Fatal("retired worker applied an in-flight public IP result")
	}
}
