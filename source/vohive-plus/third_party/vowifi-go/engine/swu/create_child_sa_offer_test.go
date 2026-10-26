package swu

import (
	"strings"
	"testing"

	"github.com/iniwex5/vowifi-go/engine/ikev2"
)

// Diagnostic-only portion adapted from HiDeck's formatter tests (BFNOC).
// Negotiation/narrowing changes are deliberately outside this personal patch.
func TestCreateChildSAOfferDescribesEveryProposalWithoutSecrets(t *testing.T) {
	first := &ikev2.Proposal{ProposalNum: 1, ProtocolID: ikev2.ProtoESP, SPI: []byte{1, 2, 3, 4}}
	first.AddTransform(ikev2.TransformTypeEncr, 12, 256)
	first.AddTransform(ikev2.TransformTypeInteg, 12, 0)
	second := &ikev2.Proposal{ProposalNum: 2, ProtocolID: ikev2.ProtoESP, SPI: []byte{1, 2, 3, 4}}
	second.AddTransform(ikev2.TransformTypeEncr, 20, 128)
	secret := []byte("secret-nonce-must-not-be-logged")
	got := createChildSAOffer([]ikev2.Payload{
		&ikev2.EncryptedPayloadSA{Proposals: []*ikev2.Proposal{first, second}},
		&ikev2.EncryptedPayloadNotify{NotifyType: ikev2.REKEY_SA},
		&ikev2.EncryptedPayloadNonce{NonceData: secret},
	})
	for _, want := range []string{"SA#1(proto=3 spi=4 1=12/256,3=12)", "SA#2(proto=3 spi=4 1=20/128)", "N(16393)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("offer %q missing %q", got, want)
		}
	}
	if strings.Contains(got, string(secret)) {
		t.Fatal("diagnostic exposed nonce")
	}
	if got := createChildSAOffer(nil); got != "none" {
		t.Fatalf("empty=%q", got)
	}
	if got := createChildSAOffer([]ikev2.Payload{nil, &ikev2.EncryptedPayloadSA{Proposals: []*ikev2.Proposal{nil}}}); got != "nil SA#?(nil)" {
		t.Fatalf("nil payload/proposal=%q", got)
	}
}
