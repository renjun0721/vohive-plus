package esim

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
)

type profileAIDTransmitter struct {
	requests  []*sgp22.ProfileOperationRequest
	profiles  []*sgp22.ProfileInfo
	lists     int
	firstErr  error
	retryErr  error
	lookupErr error
}

func (f *profileAIDTransmitter) Transmit(request bertlv.Marshaler, response bertlv.Unmarshaler) error {
	switch request := request.(type) {
	case *sgp22.ProfileOperationRequest:
		f.requests = append(f.requests, request)
		if len(f.requests) == 1 {
			return f.firstErr
		}
		return f.retryErr
	case *sgp22.ProfileInfoListRequest:
		f.lists++
		response.(*sgp22.ProfileInfoListResponse).ProfileList = f.profiles
		return f.lookupErr
	default:
		return fmt.Errorf("unexpected request %T", request)
	}
}
func (*profileAIDTransmitter) TransmitRaw([]byte) ([]byte, error) { return nil, errors.New("unused") }

func TestProfileAIDFallbackUsesProfileIdentifierAndKeepsRefresh(t *testing.T) {
	for _, operation := range []sgp22.ProfileOperation{sgp22.EnableProfile, sgp22.DisableProfile} {
		for _, refresh := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/refresh=%t", operation, refresh), func(t *testing.T) {
				iccid := mustTestICCID(t, "8986001234567890123")
				profileAID := mustDecodeHex(t, "A0000005591010FFFFFFFF8900000123")
				isdR := mustDecodeHex(t, "A0000005591010FFFFFFFF8900000100")
				transmitter := &profileAIDTransmitter{
					firstErr: &sgp22.ProfileOperationError{Operation: operation, Result: sgp22.ProfileOperationResultICCIDOrAIDNotFound},
					profiles: []*sgp22.ProfileInfo{{ICCID: iccid, ISDPAID: sgp22.ISDPAID(profileAID)}},
				}
				m := &Manager{deviceID: "test"}
				err := m.profileOperationWithAIDFallback(context.Background(), &lpa.Client{APDU: transmitter}, iccid, isdR, operation, refresh)
				if err != nil {
					t.Fatal(err)
				}
				if len(transmitter.requests) != 2 || transmitter.lists != 1 {
					t.Fatalf("requests=%d lists=%d", len(transmitter.requests), transmitter.lists)
				}
				if !transmitter.requests[0].Identifier.Tag.If(bertlv.Application, bertlv.Primitive, 26) {
					t.Fatal("first request must use ICCID")
				}
				if !transmitter.requests[1].Identifier.Tag.If(bertlv.Application, bertlv.Primitive, 15) {
					t.Fatal("fallback must use ISD-P AID")
				}
				if !reflect.DeepEqual(transmitter.requests[1].Identifier.Value, []byte(profileAID)) {
					t.Fatal("used ISD-R instead of profile AID")
				}
				for _, request := range transmitter.requests {
					if request.Refresh != refresh || request.Operation != operation {
						t.Fatal("operation/refresh changed")
					}
				}
			})
		}
	}
}

func TestProfileAIDFallbackDoesNotRetryPolicyBusyResetOrTransportFailures(t *testing.T) {
	for _, err := range []error{nil, errors.New("transport error"),
		&sgp22.ProfileOperationError{Operation: sgp22.EnableProfile, Result: sgp22.ProfileOperationResultDisallowedByPolicy},
		&sgp22.ProfileOperationError{Operation: sgp22.EnableProfile, Result: sgp22.ProfileOperationResultCATBusy},
		&sgp22.ProfileOperationError{Operation: sgp22.EnableProfile, Result: sgp22.ProfileOperationResultWrongProfileReenabling},
	} {
		transmitter := &profileAIDTransmitter{firstErr: err}
		m := &Manager{}
		got := m.profileOperationWithAIDFallback(context.Background(), &lpa.Client{APDU: transmitter}, mustTestICCID(t, "8986001234567890123"), nil, sgp22.EnableProfile, true)
		if got != err || len(transmitter.requests) != 1 || transmitter.lists != 0 {
			t.Fatalf("err=%v got=%v requests=%d lists=%d", err, got, len(transmitter.requests), transmitter.lists)
		}
	}
}

func TestProfileAIDFallbackCachedIdentifierIsScopedToEUICC(t *testing.T) {
	iccid := mustTestICCID(t, "8986001234567890123")
	isdr := mustDecodeHex(t, "A0000005591010FFFFFFFF8900000100")
	m := &Manager{overviewCache: &EsimOverview{Profiles: []EUICCProfiles{
		{AIDHex: "A0000005591010FFFFFFFF8900000199", Profiles: []ProfileItem{{ICCID: iccid.String(), ProfileAID: "A0000005591010FFFFFFFF8900000111"}}},
		{AIDHex: fmt.Sprintf("%X", isdr), Profiles: []ProfileItem{{ICCID: iccid.String(), ProfileAID: "A0000005591010FFFFFFFF8900000123"}}},
	}}}
	transmitter := &profileAIDTransmitter{firstErr: sgp22.ErrUndefined}
	if err := m.profileOperationWithAIDFallback(context.Background(), &lpa.Client{APDU: transmitter}, iccid, isdr, sgp22.EnableProfile, true); err != nil {
		t.Fatal(err)
	}
	if transmitter.lists != 0 || len(transmitter.requests) != 2 {
		t.Fatalf("requests=%d lists=%d", len(transmitter.requests), transmitter.lists)
	}
	if !reflect.DeepEqual(transmitter.requests[1].Identifier.Value, mustDecodeHex(t, "A0000005591010FFFFFFFF8900000123")) {
		t.Fatal("wrong eUICC profile")
	}
}

func TestProfileAIDFallbackStopsWhenLookupMissingOrContextCanceled(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		ctx, cancelCtx := context.WithCancel(context.Background())
		if cancel {
			cancelCtx()
		}
		defer cancelCtx()
		original := &sgp22.ProfileOperationError{Operation: sgp22.DisableProfile, Result: sgp22.ProfileOperationResultUndefinedError}
		transmitter := &profileAIDTransmitter{firstErr: original}
		m := &Manager{}
		err := m.profileOperationWithAIDFallback(ctx, &lpa.Client{APDU: transmitter}, mustTestICCID(t, "8986001234567890123"), nil, sgp22.DisableProfile, true)
		if cancel && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if !cancel && err != original {
			t.Fatal(err)
		}
		if len(transmitter.requests) != 1 {
			t.Fatal("retried without confirmed Profile AID")
		}
	}
}
