package imscore

import (
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/iniwex5/vowifi-go/internal/smscodec"
)

func TestRPReportFallsBackToRegisteredFlowWhenPeerWriteFails(t *testing.T) {
	t.Run("closed_pipe", func(t *testing.T) { testRPReportPeerFallback(t, false) })
	t.Run("fatal_peer_write", func(t *testing.T) { testRPReportPeerFallback(t, true) })
}

type fatalRPReportPeer struct{ net.Conn }

func (c fatalRPReportPeer) Write([]byte) (int, error)        { return 0, syscall.EPIPE }
func (c fatalRPReportPeer) SetWriteDeadline(time.Time) error { return nil }

func testRPReportPeerFallback(t *testing.T, fatal bool) {
	s, _, _ := newInboundSMSTestService(t)
	registered := startProtectedTestTCPOutbound(t, s)
	s.mu.Lock()
	s.cfg.LocalIP = net.IPv4(127, 0, 0, 1)
	s.protectedClientPort, s.protectedServerPort = 5062, 5064
	s.mu.Unlock()
	peer, network := net.Pipe()
	_ = network.Close()
	t.Cleanup(func() { _ = peer.Close() })
	var reportPeer net.Conn = peer
	if fatal {
		reportPeer = fatalRPReportPeer{peer}
	}

	raw := inboundSMSRequest(t, imsSMSContentType, inboundRPData(t, 0x44, "+447700900123", "peer"))
	attempts, err := s.sendRPReportWithRetryPolicy(rpReportRequest{
		Inbound: raw, Body: smscodec.BuildRPAck(0x44), RPMR: 0x44, PeerConn: reportPeer,
	}, 0, 0)
	if err != nil || attempts != 2 {
		t.Fatalf("attempts=%d error=%v", attempts, err)
	}
	report := waitForOutboundSMSControl(t, registered)
	if body, err := rawSIPBody(report); err != nil || string(body) != string(smscodec.BuildRPAck(0x44)) {
		t.Fatalf("RP report = %x, %v", body, err)
	}
	if !strings.HasPrefix(sipHeaderValue(report, "Via"), "SIP/2.0/TCP ") {
		t.Fatal("RP report did not use the registered TCP flow")
	}

	// The next report starts on the registered flow instead of waiting on a
	// peer that has already proven silent.
	nextPeer, nextNetwork := net.Pipe()
	_ = nextNetwork.Close()
	t.Cleanup(func() { _ = nextPeer.Close() })
	raw = inboundSMSRequest(t, imsSMSContentType, inboundRPData(t, 0x45, "+447700900123", "next"))
	attempts, err = s.sendRPReportWithRetryPolicy(rpReportRequest{
		Inbound: raw, Body: smscodec.BuildRPAck(0x45), RPMR: 0x45, PeerConn: nextPeer,
	}, 0, 0)
	if err != nil || attempts != 1 {
		t.Fatalf("next report attempts=%d error=%v", attempts, err)
	}
	waitForOutboundSMSControl(t, registered)
}
