package phone

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// Exercise the same public Create API on the installed baseline and candidate.
// Every DNS request goes to an in-process loopback fixture; no real DNS or modem.
func publicHostDNSFixture(t *testing.T) (*atomic.Uint32, *atomic.Bool) {
	t.Helper()
	server, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var octet atomic.Uint32
	octet.Store(10)
	var unavailable atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, addr, err := server.ReadFrom(buf)
			if err != nil {
				return
			}
			var request dnsmessage.Message
			if request.Unpack(buf[:n]) != nil {
				continue
			}
			response := dnsmessage.Message{
				Header:    dnsmessage.Header{ID: request.ID, Response: true, RecursionAvailable: true},
				Questions: request.Questions,
			}
			if unavailable.Load() {
				response.RCode = dnsmessage.RCodeNameError
			} else {
				for _, q := range request.Questions {
					if q.Type == dnsmessage.TypeA {
						response.Answers = append(response.Answers, dnsmessage.Resource{
							Header: dnsmessage.ResourceHeader{Name: q.Name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET},
							Body:   &dnsmessage.AResource{A: [4]byte{203, 0, 113, byte(octet.Load())}},
						})
					}
				}
			}
			if data, err := response.Pack(); err == nil {
				_, _ = server.WriteTo(data, addr)
			}
		}
	}()
	previous := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "udp4", server.LocalAddr().String())
	}}
	t.Cleanup(func() {
		net.DefaultResolver = previous
		_ = server.Close()
		<-done
	})
	return &octet, &unavailable
}

func TestPublicHostDNSRefreshThroughCreate(t *testing.T) {
	octet, _ := publicHostDNSFixture(t)
	manager, err := NewMediaManager(MediaOptions{UDPAddress: ":0", PublicHost: "media.vohive.invalid"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	browser, _, _ := newPCMUBrowserPeer(t)
	t.Cleanup(func() { _ = browser.Close() })
	offer := completeLocalOffer(t, browser)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first, err := manager.Create(ctx, "synthetic-owner", offer)
	if err != nil || !strings.Contains(first.SDP, " 203.0.113.10 ") {
		t.Fatalf("initial session missing fixture address: %v", err)
	}
	octet.Store(20)
	second, err := manager.Create(ctx, "synthetic-owner", offer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(second.SDP, " 203.0.113.20 ") || strings.Contains(second.SDP, " 203.0.113.10 ") {
		t.Fatal("new session advertises stale public IP after DNS changed")
	}
	if !strings.Contains(manager.Get(first.MediaID).peer.LocalDescription().SDP, " 203.0.113.10 ") {
		t.Fatal("refresh changed the existing session")
	}
}

func TestPublicHostDNSFailureThroughCreate(t *testing.T) {
	_, unavailable := publicHostDNSFixture(t)
	manager, err := NewMediaManager(MediaOptions{UDPAddress: ":0", PublicHost: "media.vohive.invalid"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	browser, _, _ := newPCMUBrowserPeer(t)
	t.Cleanup(func() { _ = browser.Close() })
	offer := completeLocalOffer(t, browser)
	unavailable.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	answer, err := manager.Create(ctx, "synthetic-owner", offer)
	if err == nil || answer != (MediaAnswer{}) || len(manager.sessions) != 0 {
		t.Fatal("DNS failure did not reject new media without stale-address fallback")
	}
}
