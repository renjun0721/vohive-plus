package esim

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
	"github.com/yibaiba/hideck/pkg/logger"
)

// Only retry identifier/compatibility errors, never card policy, busy state,
// transport failures or the expected reset after an accepted refresh.
func shouldRetryProfileByAID(err error, operation sgp22.ProfileOperation) bool {
	var result *sgp22.ProfileOperationError
	if errors.As(err, &result) {
		return result.Operation == operation &&
			(result.Result == sgp22.ProfileOperationResultICCIDOrAIDNotFound ||
				result.Result == sgp22.ProfileOperationResultUndefinedError || result.Result == 7)
	}
	// Older euicc-go maps vendor commandError (7) to ErrUndefined.
	return errors.Is(err, sgp22.ErrUndefined)
}

func validProfileAID(value string) sgp22.ISDPAID {
	aid, err := hex.DecodeString(strings.TrimSpace(value))
	if err != nil || len(aid) == 0 || len(aid) > 16 {
		return nil
	}
	return sgp22.ISDPAID(aid)
}

func (m *Manager) cachedProfileAID(iccid string, euiccAID []byte) sgp22.ISDPAID {
	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()
	if m.overviewCache == nil {
		return nil
	}
	for _, group := range m.overviewCache.Profiles {
		if !strings.EqualFold(group.AIDHex, hex.EncodeToString(euiccAID)) {
			continue
		}
		for _, profile := range group.Profiles {
			if normalizeICCIDForSwitchCompare(profile.ICCID) == normalizeICCIDForSwitchCompare(iccid) {
				return validProfileAID(profile.ProfileAID)
			}
		}
	}
	return nil
}

func (m *Manager) profileOperationWithAIDFallback(ctx context.Context, client *lpa.Client,
	iccid sgp22.ICCID, euiccAID []byte, operation sgp22.ProfileOperation, refresh bool) error {
	var run func(any, bool) error
	switch operation {
	case sgp22.EnableProfile:
		run = client.EnableProfile
	case sgp22.DisableProfile:
		run = client.DisableProfile
	default:
		return fmt.Errorf("不支持的 eSIM 兼容操作: %s", operation)
	}
	err := run(iccid, refresh)
	if !shouldRetryProfileByAID(err, operation) {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	aid := m.cachedProfileAID(iccid.String(), euiccAID)
	if len(aid) == 0 {
		// Query the selected eUICC using the existing logical channel. The
		// frontend's aid_hex selects ISD-R; it is NOT the profile's ISD-P AID.
		profiles, lookupErr := client.ListProfile(nil, nil)
		if lookupErr != nil {
			return fmt.Errorf("%w（读取 Profile AID 失败: %v）", err, lookupErr)
		}
		for _, profile := range profiles {
			if profile != nil && normalizeICCIDForSwitchCompare(profile.ICCID.String()) == normalizeICCIDForSwitchCompare(iccid.String()) {
				aid = validProfileAID(hex.EncodeToString(profile.ISDPAID))
				break
			}
		}
	}
	if len(aid) == 0 {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	logger.Warn("ICCID 操作不兼容，按 Profile AID 重试",
		"device", m.deviceID, "operation", operation.String(), "err", err)
	// Keep the caller's refresh flag and perform exactly one AID attempt.
	if retryErr := run(aid, refresh); retryErr != nil {
		return fmt.Errorf("按 Profile AID 重试失败: %w", retryErr)
	}
	return nil
}
