package ipsec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestSanitizeSOCKS5RelayLookupErrorPreservesOnlySafeClassification(t *testing.T) {
	const secret = "syntheticproxysecret"
	privateCause := errors.New(secret)
	tests := []struct {
		name          string
		cause         error
		wantContext   error
		wantDNS       bool
		wantTimeout   bool
		wantTemporary bool
		wantNotFound  bool
	}{
		{name: "canceled", cause: fmt.Errorf("nested: %w", context.Canceled), wantContext: context.Canceled},
		{name: "deadline", cause: fmt.Errorf("nested: %w", context.DeadlineExceeded), wantContext: context.DeadlineExceeded},
		{name: "unknown", cause: privateCause},
		{name: "timeout", cause: &net.DNSError{Err: secret, Name: secret, Server: secret, UnwrapErr: privateCause,
			IsTimeout: true, IsTemporary: true}, wantDNS: true, wantTimeout: true, wantTemporary: true},
		{name: "not found", cause: &net.DNSError{Err: secret, Name: secret, Server: secret, UnwrapErr: privateCause,
			IsNotFound: true}, wantDNS: true, wantNotFound: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := fmt.Errorf("%s: %w", secret, test.cause)
			got := sanitizeSOCKS5RelayLookupError(original)
			if got == nil || strings.Contains(fmt.Sprintf("%+v", got), secret) {
				t.Fatal("sanitized lookup error retained sensitive text")
			}
			if test.wantContext != nil && !errors.Is(got, test.wantContext) {
				t.Fatalf("lost context classification: %v", got)
			}
			if errors.Is(got, original) || errors.Is(got, privateCause) {
				t.Fatal("sanitized lookup error retained an original error chain")
			}
			for current := got; current != nil; current = errors.Unwrap(current) {
				if current == original || current == test.cause {
					t.Fatal("original error remains accessible through Unwrap")
				}
			}
			var dns *net.DNSError
			if errors.As(got, &dns) {
				if dns.Name != "" || dns.Server != "" || dns.UnwrapErr != nil || strings.Contains(dns.Err, secret) {
					t.Fatal("sanitized DNS error retained sensitive fields")
				}
				if test.wantDNS && (dns.IsTimeout != test.wantTimeout || dns.IsTemporary != test.wantTemporary ||
					dns.IsNotFound != test.wantNotFound) {
					t.Fatal("sanitized DNS error changed classification")
				}
			} else if test.wantDNS {
				t.Fatal("lost DNS error classification")
			}
			if test.wantTimeout {
				var network net.Error
				if !errors.As(got, &network) || !network.Timeout() || network.Temporary() != test.wantTemporary {
					t.Fatal("lost network timeout classification")
				}
			}
		})
	}
}
