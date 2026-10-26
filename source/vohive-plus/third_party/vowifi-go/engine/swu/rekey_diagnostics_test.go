package swu

import (
	"strings"
	"testing"

	"github.com/iniwex5/vowifi-go/engine/ikev2"
)

// Adapted from HiDeck f619660 (BFNOC); kept separate from newer SA lifecycle tests.
func TestIKERekeyResponseErrorNamesReceivedPayloads(t *testing.T) {
	session := &Session{}
	for _, tc := range []struct {
		name     string
		payloads []ikev2.Payload
		want     string
	}{
		{"empty", nil, "(response: none)"},
		{"notify", []ikev2.Payload{&ikev2.EncryptedPayloadNotify{NotifyType: ikev2.NO_ADDITIONAL_SAS}}, "(response: N(35))"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := session.completeInitiatedIKESARekey(initiatedIKERekey{payloads: tc.payloads})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}
