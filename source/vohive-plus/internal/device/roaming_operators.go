package device

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/iniwex5/quectel-qmi-go/pkg/qmi"
	"github.com/yibaiba/hideck/internal/backend"
	"github.com/yibaiba/hideck/pkg/logger"
)

const (
	roamingAttemptWait = 90 * time.Second
	roamingRetryWait   = 5 * time.Minute
	roamingLossGrace   = 30 * time.Second
)

type roamingOperatorPolicy struct {
	ICCID       string                                `json:"iccid"`
	Operators   []backend.SetOperatorSelectionRequest `json:"operators"`
	fingerprint [32]byte
}

type roamingOperatorFile struct {
	Version  int                     `json:"version"`
	Policies []roamingOperatorPolicy `json:"policies"`
}

func parseRoamingOperatorPolicies(data []byte) (map[string]roamingOperatorPolicy, error) {
	var file roamingOperatorFile
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&file); err != nil {
		return nil, fmt.Errorf("漫游策略 JSON 无效: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, errors.New("漫游策略包含多余内容")
	}
	if file.Version != 1 {
		return nil, errors.New("漫游策略版本必须为 1")
	}
	policies := make(map[string]roamingOperatorPolicy)
	for _, policy := range file.Policies {
		policy.ICCID = normalizeSIMIdentityForCompare(policy.ICCID)
		if len(policy.ICCID) < 18 || len(policy.ICCID) > 22 || strings.Trim(policy.ICCID, "0123456789") != "" {
			return nil, errors.New("漫游策略必须绑定有效 ICCID")
		}
		if _, duplicate := policies[policy.ICCID]; duplicate {
			return nil, errors.New("同一卡不能有重复漫游策略")
		}
		if len(policy.Operators) == 0 || len(policy.Operators) > 8 {
			return nil, errors.New("漫游白名单需要 1 至 8 个运营商")
		}
		seen := make(map[string]bool)
		for i, op := range policy.Operators {
			// Auto is intentionally unrepresentable: exhaustion must never leave
			// the allowlist. The file stores no modem/SIM mutation commands.
			if op.Mode != "" && op.Mode != backend.OperatorSelectionManual {
				return nil, errors.New("漫游白名单不允许自动选网")
			}
			op.Mode = backend.OperatorSelectionManual
			if len(op.PLMN) != 5 {
				return nil, errors.New("漫游白名单当前仅支持五位 PLMN")
			}
			if _, err := backend.NormalizeSetOperatorSelectionRequest(op); err != nil {
				return nil, errors.New("漫游白名单 PLMN 无效")
			}
			if op.RAT != backend.OperatorRATLTE {
				return nil, errors.New("漫游白名单当前仅支持 LTE")
			}
			if seen[op.PLMN] {
				return nil, errors.New("漫游白名单 PLMN 重复")
			}
			seen[op.PLMN] = true
			policy.Operators[i] = op
		}
		encoded, _ := json.Marshal(policy.Operators)
		policy.fingerprint = sha256.Sum256(encoded)
		policies[policy.ICCID] = policy
	}
	return policies, nil
}

// State is local to the loop. Time-driven decisions are pure and tested with
// a fake clock; no tests need real SIMs, scanning, registration or messages.
type roamingOperatorState struct {
	owner       workerLifecycleOwner
	iccid       string
	policy      [32]byte
	initialized bool
	pending     bool
	index       int
	deadline    time.Time
	retryAt     time.Time
	lostAt      time.Time
}

func (s *roamingOperatorState) next(now time.Time, registered bool, plmn string, operators []backend.SetOperatorSelectionRequest) (int, bool) {
	if !s.initialized {
		s.initialized, s.pending, s.index = true, true, 0
		s.deadline = now.Add(roamingAttemptWait)
		return 0, true
	}
	if s.pending {
		if registered && plmn == operators[s.index].PLMN {
			s.pending, s.lostAt, s.retryAt = false, time.Time{}, time.Time{}
			return 0, false
		}
		if now.Before(s.deadline) {
			return 0, false
		}
		s.index++
		if s.index < len(operators) {
			s.deadline = now.Add(roamingAttemptWait)
			return s.index, true
		}
		s.pending = false
		s.retryAt = now.Add(roamingRetryWait)
		return 0, false
	}
	for _, op := range operators {
		if registered && plmn == op.PLMN {
			s.lostAt, s.retryAt = time.Time{}, time.Time{}
			return 0, false // Keep a working fallback; do not periodically disrupt SMS.
		}
	}
	if now.Before(s.retryAt) {
		return 0, false
	}
	if !registered && s.retryAt.IsZero() {
		if s.lostAt.IsZero() {
			s.lostAt = now
			return 0, false
		}
		if now.Sub(s.lostAt) < roamingLossGrace {
			return 0, false
		}
	}
	s.pending, s.index, s.lostAt = true, 0, time.Time{}
	s.deadline = now.Add(roamingAttemptWait)
	return 0, true
}

type roamingOperatorManager struct {
	pool     *Pool
	path     string
	mu       sync.RWMutex
	policies map[string]roamingOperatorPolicy
	digest   [32]byte
	lastBad  string
	states   map[string]*roamingOperatorState
}

func (p *Pool) ConfigureRoamingOperatorPolicies(path string) error {
	m := &roamingOperatorManager{pool: p, path: path, policies: make(map[string]roamingOperatorPolicy), states: make(map[string]*roamingOperatorState)}
	if err := m.reload(); err != nil {
		return err
	}
	p.roamingOperators = m // Configure once, before starting pool goroutines.
	return nil
}

func (m *roamingOperatorManager) reload() error {
	f, err := os.Open(m.path)
	if os.IsNotExist(err) {
		return nil
	} // A missing file never discards a loaded allowlist.
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		return err
	}
	if len(data) > 65536 {
		return errors.New("漫游策略文件过大")
	}
	digest := sha256.Sum256(data)
	if digest == m.digest {
		return nil
	}
	policies, err := parseRoamingOperatorPolicies(data)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.policies = policies
	m.mu.Unlock()
	m.digest = digest
	logger.Info("已加载按卡漫游白名单", "policies", len(policies))
	return nil
}

func (m *roamingOperatorManager) policy(iccid string) (roamingOperatorPolicy, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.policies[normalizeSIMIdentityForCompare(iccid)]
	return p, ok
}

func (w *Worker) hasRoamingOperatorPolicy() bool {
	if w == nil || w.Pool == nil || w.Pool.roamingOperators == nil || w.Backend == nil || w.Backend.Mode() != backend.BackendQMI {
		return false
	}
	_, ok := w.Pool.roamingOperators.policy(w.CurrentICCID())
	return ok
}

func (w *Worker) roamingOperatorDataReady(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ss, err := w.Backend.GetServingSystem(ctx)
	if err != nil {
		return err
	}
	policy, ok := w.Pool.roamingOperators.policy(w.CurrentICCID())
	if ok && ss != nil && (ss.RegStatus == 1 || ss.RegStatus == 5) && ss.PSAttached {
		plmn := fmt.Sprintf("%03d%02d", ss.MCC, ss.MNC)
		for _, op := range policy.Operators {
			if op.PLMN == plmn {
				return nil
			}
		}
	}
	return errors.New("正在等待白名单运营商完成驻网，暂缓连接数据")
}

func (m *roamingOperatorManager) run() {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-m.pool.ctx.Done():
			return
		case <-t.C:
			if err := m.reload(); err != nil {
				if err.Error() != m.lastBad {
					logger.Warn("漫游策略更新无效，保留上次有效白名单", "err", err)
					m.lastBad = err.Error()
				}
			} else {
				m.lastBad = ""
			}
			for _, w := range m.pool.healthCheckWorkerSnapshot() {
				m.reconcile(w)
			}
		}
	}
}

func (m *roamingOperatorManager) reconcile(w *Worker) {
	if w == nil || w.Backend == nil || w.Backend.Mode() != backend.BackendQMI {
		return
	}
	transition := m.pool.policyTransitionFor(w.ID)
	if !transition.TryLock() {
		return
	}
	defer transition.Unlock()
	owner := m.pool.captureWorkerLifecycleOwner(w.ID)
	if owner.worker != w || w.shouldAbortCellularRegistration() || m.pool.IsESIMSwitching(w.ID) || w.IsOperatorScanActive() {
		return
	}
	if m.pool.voiceGateway != nil && m.pool.voiceGateway.ActiveCall(w.ID) != nil {
		return
	}
	if m.pool.volteCtl != nil && m.pool.volteCtl.ActiveCall(w.ID) != nil {
		return
	}
	w.cacheMu.RLock()
	ready := w.state.Identity.Ready
	w.cacheMu.RUnlock()
	if !ready {
		return
	}
	iccid := normalizeSIMIdentityForCompare(w.CurrentICCID())
	if iccid == "" {
		return
	}
	policy, managed := m.policy(iccid)
	old := m.states[w.ID]
	if !managed && old == nil {
		return
	}
	ctx, cancel := context.WithTimeout(m.pool.ctx, 12*time.Second)
	defer cancel()
	reader, ok := w.Backend.(liveSIMIdentityReader)
	if !ok {
		return
	}
	live, err := reader.GetICCIDLive(ctx)
	if err != nil || normalizeSIMIdentityForCompare(live) != iccid {
		return
	}
	provider, ok := w.Backend.(backend.OperatorSelectionProvider)
	if !ok {
		return
	}
	if !managed {
		// A different/disabled SIM must inherit the device's original choice,
		// never this card's allowlist. We never persist our targets in Config.
		req := backend.SetOperatorSelectionRequest{Mode: backend.OperatorSelectionAutomatic}
		if w.Config.OperatorSelectionMode == "manual" {
			req.Mode = backend.OperatorSelectionManual
			req.PLMN = w.Config.OperatorSelectionPLMN
			req.RAT = backend.OperatorAccessTechnology(w.Config.OperatorSelectionRAT)
		}
		if _, err := provider.SetOperatorSelection(ctx, req); err == nil {
			delete(m.states, w.ID)
			logger.Info("本卡不再匹配漫游白名单，已恢复设备原选网方式", "device", w.ID)
		}
		return
	}
	if old == nil || old.owner != owner || old.iccid != iccid || old.policy != policy.fingerprint {
		if err := w.cancelRadioRegistrationReconcile(ctx, "按卡漫游白名单接管"); err != nil {
			return
		}
		old = &roamingOperatorState{owner: owner, iccid: iccid, policy: policy.fingerprint}
		m.states[w.ID] = old
	}
	ss, err := w.Backend.GetServingSystem(ctx)
	if err != nil || ss == nil {
		return
	}
	plmn := fmt.Sprintf("%03d%02d", ss.MCC, ss.MNC)
	now := time.Now()
	wasPending := old.pending
	index, act := old.next(now, ss.RegStatus == 1 || ss.RegStatus == 5, plmn, policy.Operators)
	if wasPending && !old.pending {
		if now.Before(old.retryAt) {
			logger.Warn("漫游白名单暂时均不可用，等待重试且不退回自动选网", "device", w.ID, "retry_seconds", 300)
		} else {
			logger.Info("漫游白名单驻网成功", "device", w.ID, "plmn", plmn)
		}
	}
	if !act {
		return
	}
	live, err = reader.GetICCIDLive(ctx)
	if err != nil || normalizeSIMIdentityForCompare(live) != iccid {
		return
	}
	if !m.pool.workerLifecycleOwnerCurrent(owner) || normalizeSIMIdentityForCompare(w.CurrentICCID()) != iccid || m.pool.IsESIMSwitching(w.ID) {
		return
	}
	// The manager owns NAS registration while this policy is active. A
	// power-cycle duration avoids leaving a permanent target on another SIM.
	controller, ok := w.Backend.(interface {
		NASInitiateNetworkRegister(context.Context, backend.NASRegisterRequest) error
	})
	if !ok {
		return
	}
	sel, err := backend.NormalizeSetOperatorSelectionRequest(policy.Operators[index])
	if err != nil {
		return
	}
	req, err := backend.BuildManualNASRegisterRequest(sel)
	if err != nil {
		return
	}
	req.ChangeDuration = qmi.NASChangeDurationPowerCycle
	req.HasChangeDuration = true
	logger.Info("按卡漫游策略开始手动驻网", "device", w.ID, "target_plmn", sel.PLMN, "priority", index+1, "wait_seconds", 90)
	if err := controller.NASInitiateNetworkRegister(ctx, req); err != nil {
		old.deadline = now.Add(20 * time.Second)
		logger.Warn("白名单运营商注册请求失败，稍后尝试下一项", "device", w.ID, "target_plmn", sel.PLMN, "err", err)
	}
}
