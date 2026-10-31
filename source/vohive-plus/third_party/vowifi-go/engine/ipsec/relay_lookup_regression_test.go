package ipsec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"testing"
)

// Exercise the existing public parsers with an offline DNS failure. No proxy,
// modem, or external DNS service is contacted by this regression.
func TestSOCKS5RelayLookupDoesNotExposeProxyData(t *testing.T) {
	const secret = "syntheticproxysecret"
	const domain = "relay." + secret + ".invalid."
	var lookups atomic.Int32
	previous := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(context.Context, string, string) (net.Conn, error) {
		lookups.Add(1)
		return nil, errors.New("untrusted resolver detail: " + secret)
	}}
	t.Cleanup(func() { net.DefaultResolver = previous })
	address := append([]byte{byte(len(domain))}, domain...)
	address = append(address, 0x1e, 0xd9)
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"associate reply", func() error {
			_, err := readSocks5Reply(bytes.NewReader(append([]byte{5, 0, 0, 3}, address...)))
			return err
		}},
		{"UDP datagram", func() error {
			_, err := DecodeSocks5UDPDatagram(append([]byte{0, 0, 0, 3}, address...))
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := lookups.Load()
			err := tc.run()
			if err == nil || lookups.Load() == before {
				t.Fatal("fixture did not reject an unresolved domain through the private resolver")
			}
			var output bytes.Buffer
			slog.New(slog.NewJSONHandler(&output, nil)).Warn("VoWiFi operation failed", "error", err)
			for _, value := range []string{err.Error(), fmt.Sprintf("%+v", err), output.String()} {
				if strings.Contains(value, secret) || strings.Contains(value, domain) || strings.Contains(value, "untrusted resolver detail") {
					t.Fatal("relay lookup failure exposed proxy or resolver data")
				}
			}
			var dns *net.DNSError
			if !errors.As(err, &dns) || dns.Name != "" || dns.Server != "" || dns.UnwrapErr != nil {
				t.Fatal("lookup failure lost safe DNS classification or retained private fields")
			}
		})
	}
}
