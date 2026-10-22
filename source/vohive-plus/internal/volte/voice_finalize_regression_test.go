package volte

import (
	"context"
	"testing"

	"github.com/iniwex5/quectel-qmi-go/pkg/qmi"
	"github.com/iniwex5/vowifi-go/runtimehost/voicehost"
)

// Cover both QMI terminal forms, inbound/outbound calls and repeated reports.
// All modem operations use FakeModem; no hardware or carrier is contacted.
func TestVoiceFinalizeTerminalPathsExactlyOnce(t *testing.T) {
	for _, inbound := range []bool{false, true} {
		for _, disappeared := range []bool{false, true} {
			for _, answered := range []bool{false, true} {
				name := "outbound"
				if inbound {
					name = "inbound"
				}
				if disappeared {
					name += "/snapshot-disappeared"
				} else {
					name += "/explicit-end"
				}
				if answered {
					name += "/answered"
				} else {
					name += "/unanswered"
				}
				t.Run(name, func(t *testing.T) {
					ctl, host := enableVoice(t)
					var events []voicehost.CallEvent
					ctl.SubscribeCallEvents(func(ev voicehost.CallEvent) { events = append(events, ev) })
					direction := qmiDirMO
					state := qmiCallAlerting
					if inbound {
						direction, state = qmiDirMT, qmiCallIncoming
					} else {
						if _, err := ctl.BeginCall(context.Background(), voicehost.BeginCallRequest{DeviceID: "wwan1", Callee: "10086"}); err != nil {
							t.Fatal(err)
						}
					}
					host.fireVoice(&qmi.VoiceAllCallInfo{
						Calls:              []qmi.VoiceCallInfo{{ID: 1, State: state, Direction: direction}},
						RemotePartyNumbers: []qmi.VoiceRemotePartyNumber{{CallID: 1, Number: "10086"}},
					})
					if answered {
						host.fireVoice(&qmi.VoiceAllCallInfo{Calls: []qmi.VoiceCallInfo{{ID: 1, State: qmiCallConversation, Direction: direction}}})
					}
					end := &qmi.VoiceAllCallInfo{}
					if !disappeared {
						end.Calls = []qmi.VoiceCallInfo{{ID: 1, State: qmiCallEnd, Direction: direction}}
					}
					host.fireVoice(end)
					host.fireVoice(end)
					ended, finalized := 0, 0
					var terminal voicehost.CallEvent
					for i, ev := range events {
						if ev.Type == "CallEnded" || ev.Type == "CallCanceled" {
							ended++
							terminal = ev
						}
						if ev.Type == "CallFinalized" {
							finalized++
							if i == 0 || events[i-1].Type != "CallEnded" {
								t.Fatalf("finalized before terminal: %v", eventTypes(events))
							}
							if ev.CallID != terminal.CallID || ev.DeviceID != "wwan1" || ev.Direction != terminal.Direction || ev.Caller != "10086" || ev.State != "completed" || ev.Time.IsZero() {
								t.Fatalf("finalized metadata mismatch: %+v terminal=%+v", ev, terminal)
							}
						}
					}
					want := 0
					if answered {
						want = 1
					}
					if ended != 1 || finalized != want {
						t.Fatalf("terminal=%d finalized=%d want=%d events=%v", ended, finalized, want, eventTypes(events))
					}
				})
			}
		}
	}
}
