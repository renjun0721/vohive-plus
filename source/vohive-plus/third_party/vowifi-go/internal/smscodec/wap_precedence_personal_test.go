package smscodec

import (
	"encoding/hex"
	"github.com/warthog618/sms/encoding/tpdu"
	"strings"
	"testing"
)

// Shared WAP port 2948 must not override a recognized content type or payload.
func TestWAPPushClassificationPrecedence(t *testing.T) {
	tests := []struct {
		name, ct string
		body     []byte
		want     binaryKind
	}{
		{"si", "application/vnd.wap.sic", []byte("https://example.com/si"), binaryKindWAPSI},
		{"sl", "application/vnd.wap.slc", []byte("https://example.com/sl"), binaryKindWAPSL},
		{"mms", "application/vnd.wap.mms-message", []byte{0x8c, 0x82}, binaryKindMMSNotification},
		{"explicit_mms_over_si_hint", "application/vnd.wap.mms-message", []byte{3, 5, 106, 0}, binaryKindMMSNotification},
		{"explicit_sl_over_si_hint", "application/vnd.wap.slc", []byte{3, 5, 106, 0}, binaryKindWAPSL},
		{"explicit_si_over_mms_hint", "application/vnd.wap.sic", []byte{0x8c, 0x82}, binaryKindWAPSI},
		{"explicit_oma_over_si_hint", "application/vnd.wap.connectivity-wbxml", []byte{3, 5, 106, 0}, binaryKindOmaCP},
		{"explicit_oma_over_sl_hint", "application/vnd.wap.connectivity-wbxml", []byte{3, 7, 106, 0}, binaryKindOmaCP},
		{"explicit_oma_over_mms_hint", "application/vnd.wap.connectivity-wbxml", []byte{0x8c, 0x82}, binaryKindOmaCP},
		{"si_payload", "application/octet-stream", []byte{3, 5, 106, 0}, binaryKindWAPSI},
		{"sl_payload", "application/octet-stream", []byte{3, 7, 106, 0}, binaryKindWAPSL},
		{"mms_payload", "application/octet-stream", []byte{0x8c, 0x82}, binaryKindMMSNotification},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := append([]byte{1, 6, byte(len(tc.ct) + 1)}, []byte(tc.ct+"\x00")...)
			msg = append(msg, tc.body...)
			for _, port := range []bool{false, true} {
				tp := &tpdu.TPDU{}
				if port {
					tp.UDH = tpdu.UserDataHeader{{ID: 5, Data: []byte{0x0b, 0x84, 0x23, 0xf0}}}
				}
				got := classifyBinarySMS(tp, msg)
				if got.Kind != tc.want {
					t.Errorf("port=%v kind=%s want=%s", port, got.Kind, tc.want)
				}
				if got.RawHex != hex.EncodeToString(msg) {
					t.Fatal("raw payload changed")
				}
				if got.HasDestPort != port || (port && got.DestPort != 2948) {
					t.Fatal("port metadata changed")
				}
				if tc.want != binaryKindOmaCP && got.IsPossEncrypted {
					t.Fatal("recognized plaintext marked encrypted")
				}
			}
		})
	}
}

// Full SMS-DELIVER wire path, not only the classifier helper.
func TestDecodeDeliverTPDUWAPSharedPort(t *testing.T) {
	for _, tc := range []struct{ ct, label string }{
		{"application/vnd.wap.sic", "WAP SI Push"},
		{"application/vnd.wap.slc", "WAP SL Push"},
		{"application/vnd.wap.mms-message", "MMS Notification"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			msg := append([]byte{1, 6, byte(len(tc.ct) + 1)}, []byte(tc.ct+"\x00")...)
			msg = append(msg, 0x8c, 0x82)
			ud := append([]byte{6, 5, 4, 0x0b, 0x84, 0x23, 0xf0}, msg...)
			// UDHI DELIVER, sender 10086, PID 0, 8-bit DCS, fixed valid SCTS.
			wire := []byte{0x40, 5, 0x81, 0x01, 0x80, 0xf6, 0, 4, 0x62, 0x11, 0x42, 0x01, 0, 0, 0, byte(len(ud))}
			wire = append(wire, ud...)
			sender, text, _, concat, err := DecodeDeliverTPDU(wire)
			if err != nil {
				t.Fatal(err)
			}
			if sender != "10086" || concat.IsConcat {
				t.Fatalf("sender=%q concat=%+v", sender, concat)
			}
			if !strings.HasPrefix(text, "["+tc.label+"]\n") || !strings.Contains(text, "dest_port=2948\n") || !strings.Contains(text, "raw="+hex.EncodeToString(msg)) {
				t.Fatalf("unexpected classification: %s", text)
			}
			if strings.Contains(text, "可能加密") {
				t.Fatal("plaintext marked encrypted")
			}
		})
	}
}
