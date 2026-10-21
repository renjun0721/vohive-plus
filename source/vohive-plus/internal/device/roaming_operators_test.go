package device

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/iniwex5/quectel-qmi-go/pkg/qmi"
	"github.com/yibaiba/hideck/internal/backend"
	qmicore "github.com/yibaiba/hideck/internal/qmi"
)

const roamingTestICCID = "8985000000000000001"

func testRoamingOperators() []backend.SetOperatorSelectionRequest {
	return []backend.SetOperatorSelectionRequest{{Mode: "manual", PLMN: "46001", RAT: "lte"}, {Mode: "manual", PLMN: "46000", RAT: "lte"}}
}

func roamingTestJSON() []byte {
	b, _ := json.Marshal(roamingOperatorFile{Version: 1, Policies: []roamingOperatorPolicy{{ICCID: roamingTestICCID, Operators: testRoamingOperators()}}})
	return b
}

func TestRoamingOperatorRejectsInvalidAllowlist(t *testing.T) {
	for _, text := range []string{
		`{"version":2,"policies":[]}`,
		`{"version":1,"policies":[{"iccid":"8985000000000000001","operators":[]}]}`,
		`{"version":1,"policies":[{"iccid":"wrong","operators":[{"plmn":"46001","rat":"lte"}]}]}`,
		`{"version":1,"policies":[{"iccid":"8985000000000000001","operators":[{"mode":"automatic","plmn":"46001","rat":"lte"}]}]}`,
		`{"version":1,"policies":[{"iccid":"8985000000000000001","operators":[{"plmn":"46001","rat":"lte"},{"plmn":"46001","rat":"lte"}]}]}`,
		`{"version":1,"policies":[{"iccid":"8985000000000000001","operators":[{"plmn":"46001","rat":"bogus"}]}]}`,
		`{"version":1,"policies":[]} {}`,
		`{"version":1,"policies":[],"unknown":true}`,
	} {
		if _, err := parseRoamingOperatorPolicies([]byte(text)); err == nil {
			t.Fatalf("accepted invalid configuration: %s", text)
		}
	}
	p, err := parseRoamingOperatorPolicies(roamingTestJSON())
	if err != nil || len(p) != 1 {
		t.Fatalf("valid policy: %v %v", p, err)
	}
}

func TestRoamingOperatorBootCorrectsTelecomAndWaitsBeforeFallback(t *testing.T) {
	now := time.Unix(1000, 0)
	ops := testRoamingOperators()
	s := roamingOperatorState{}
	if i, act := s.next(now, true, "46011", ops); !act || i != 0 {
		t.Fatal("already registered Telecom must still select preferred Unicom")
	}
	if _, act := s.next(now.Add(80*time.Second), true, "46011", ops); act {
		t.Fatal("must allow registration to settle")
	}
	if i, act := s.next(now.Add(90*time.Second), true, "46011", ops); !act || i != 1 {
		t.Fatal("must try Mobile after Unicom deadline")
	}
	s.next(now.Add(100*time.Second), true, "46000", ops)
	if _, act := s.next(now.Add(time.Hour), true, "46000", ops); act {
		t.Fatal("must keep working Mobile instead of repeatedly interrupting SMS")
	}
	s.next(now.Add(time.Hour+time.Second), false, "", ops)
	if _, act := s.next(now.Add(time.Hour+20*time.Second), false, "", ops); act {
		t.Fatal("short coverage loss must not trigger a switch")
	}
	if i, act := s.next(now.Add(time.Hour+40*time.Second), false, "", ops); !act || i != 0 {
		t.Fatal("recovery must prefer Unicom again")
	}
}

func TestRoamingOperatorExhaustionNeverSelectsAutomaticOrTelecom(t *testing.T) {
	now := time.Unix(1000, 0)
	ops := testRoamingOperators()
	s := roamingOperatorState{}
	for second := 0; second < 480; second += 10 {
		i, act := s.next(now.Add(time.Duration(second)*time.Second), false, "", ops)
		if act && (ops[i].Mode != "manual" || (ops[i].PLMN != "46001" && ops[i].PLMN != "46000")) {
			t.Fatal("escaped allowlist")
		}
		if act && second != 0 && second != 90 {
			t.Fatalf("unexpected retry at %d", second)
		}
	}
	if i, act := s.next(now.Add(480*time.Second), true, "46011", ops); !act || i != 0 {
		t.Fatal("after five-minute backoff restart with Unicom")
	}
}

type roamingTestBackend struct {
	mockOperatorProvider
	iccid    string
	serving  backend.ServingSystem
	requests []backend.NASRegisterRequest
	restore  []backend.SetOperatorSelectionRequest
}

func (b *roamingTestBackend) Mode() string                                 { return backend.BackendQMI }
func (b *roamingTestBackend) GetICCIDLive(context.Context) (string, error) { return b.iccid, nil }
func (b *roamingTestBackend) GetIMSILive(context.Context) (string, error) {
	return "454030000000001", nil
}
func (b *roamingTestBackend) GetServingSystem(context.Context) (*backend.ServingSystem, error) {
	v := b.serving
	return &v, nil
}
func (b *roamingTestBackend) NASInitiateNetworkRegister(_ context.Context, r backend.NASRegisterRequest) error {
	b.requests = append(b.requests, r)
	return nil
}
func (b *roamingTestBackend) SetOperatorSelection(_ context.Context, r backend.SetOperatorSelectionRequest) (backend.OperatorSelection, error) {
	b.restore = append(b.restore, r)
	return backend.OperatorSelection{}, nil
}

func roamingFixture(t *testing.T) (*Pool, *Worker, *roamingTestBackend) {
	t.Helper()
	p := NewPool(nil)
	t.Cleanup(p.cancel)
	file := filepath.Join(t.TempDir(), "roaming-operators.json")
	if err := os.WriteFile(file, roamingTestJSON(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := p.ConfigureRoamingOperatorPolicies(file); err != nil {
		t.Fatal(err)
	}
	b := &roamingTestBackend{iccid: roamingTestICCID, serving: backend.ServingSystem{MCC: 460, MNC: 11, RegStatus: 5}}
	w := &Worker{ID: "test", Pool: p, Backend: b, generation: 1}
	w.state.Identity.Ready = true
	w.state.Identity.ICCID = roamingTestICCID
	p.workers[w.ID] = w
	return p, w, b
}

func TestRoamingOperatorCorrectsRegisteredTelecomWithoutPersistingDeviceTarget(t *testing.T) {
	p, w, b := roamingFixture(t)
	p.roamingOperators.reconcile(w)
	if len(b.requests) != 1 || b.requests[0].Mode != "manual" || b.requests[0].MNC != 1 || !b.requests[0].HasChangeDuration || b.requests[0].ChangeDuration != qmi.NASChangeDurationPowerCycle {
		t.Fatalf("wrong registration request: %+v", b.requests)
	}
	if w.Config.OperatorSelectionMode != "" || w.Config.OperatorSelectionPLMN != "" {
		t.Fatal("SIM-specific target leaked into device config")
	}
	if _, err := w.SetOperatorSelection(context.Background(), backend.SetOperatorSelectionRequest{Mode: "automatic"}); err == nil {
		t.Fatal("manual API must not escape active allowlist")
	}
	// QMI recovery must not re-enable auto even if the pool's default is auto.
	if !w.hasRoamingOperatorPolicy() {
		t.Fatal("missing policy guard")
	}
	w.QMICore = &qmicore.Manager{}
	if err := w.ensureQMIRegistration(context.Background(), false); err != errQMIRegistrationSkipped {
		t.Fatalf("generic idle recovery not skipped: %v", err)
	}
	if err := w.ensureQMIRegistration(context.Background(), true); err == nil {
		t.Fatal("must not start data on forbidden Telecom")
	}
	b.serving.MNC = 1
	b.serving.PSAttached = true
	if err := w.ensureQMIRegistration(context.Background(), true); err != nil {
		t.Fatalf("allowed data registration blocked: %v", err)
	}
	w.setCellularRadioSuppressed(true)
	if err := w.ensureQMIRegistration(context.Background(), true); err != errQMIRegistrationSkipped {
		t.Fatalf("flight/VoWiFi suppression must take priority: %v", err)
	}
}

func TestRoamingOperatorSwitchingSIMRestoresOriginalDeviceMode(t *testing.T) {
	p, w, b := roamingFixture(t)
	p.roamingOperators.reconcile(w)
	w.cacheMu.Lock()
	w.state.Identity.ICCID = "8985000000000000002"
	w.cacheMu.Unlock()
	b.iccid = "8985000000000000002"
	p.roamingOperators.reconcile(w)
	if w.hasRoamingOperatorPolicy() || len(b.requests) != 1 || len(b.restore) != 1 || b.restore[0].Mode != "automatic" {
		t.Fatalf("SIM isolation failed: requests=%+v restore=%+v", b.requests, b.restore)
	}
}

func TestRoamingOperatorSkipsStaleIdentitySuppressionAndBusyTransition(t *testing.T) {
	for _, reason := range []string{"identity-mismatch", "not-ready", "suppressed", "removed", "transition"} {
		t.Run(reason, func(t *testing.T) {
			p, w, b := roamingFixture(t)
			switch reason {
			case "identity-mismatch":
				b.iccid = "8985000000000000002"
			case "not-ready":
				w.state.Identity.Ready = false
			case "suppressed":
				w.setCellularRadioSuppressed(true)
			case "removed":
				delete(p.workers, w.ID)
			case "transition":
				mu := p.policyTransitionFor(w.ID)
				mu.Lock()
				defer mu.Unlock()
			}
			p.roamingOperators.reconcile(w)
			if len(b.requests) != 0 {
				t.Fatal("unsafe selection attempted")
			}
		})
	}
}

func TestRoamingOperatorReloadRetainsAllowlistOnBadOrMissingFile(t *testing.T) {
	p, w, _ := roamingFixture(t)
	m := p.roamingOperators
	if err := os.WriteFile(m.path, []byte(`{"version":1,"policies":[`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.reload(); err == nil {
		t.Fatal("invalid reload accepted")
	}
	if !w.hasRoamingOperatorPolicy() {
		t.Fatal("invalid update cleared allowlist")
	}
	if err := os.Remove(m.path); err != nil {
		t.Fatal(err)
	}
	if err := m.reload(); err != nil {
		t.Fatal(err)
	}
	if !w.hasRoamingOperatorPolicy() {
		t.Fatal("missing file cleared allowlist")
	}
	if err := os.WriteFile(m.path, []byte(`{"version":1,"policies":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.reload(); err != nil {
		t.Fatal(err)
	}
	if w.hasRoamingOperatorPolicy() {
		t.Fatal("explicit empty policy should disable management")
	}
}

func TestRoamingOperatorConcurrentLookupDuringReload(t *testing.T) {
	p, w, _ := roamingFixture(t)
	m := p.roamingOperators
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = w.hasRoamingOperatorPolicy()
			}
		}()
	}
	for i := 0; i < 10; i++ {
		m.digest = [32]byte{}
		if err := m.reload(); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}
