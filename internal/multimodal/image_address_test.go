package multimodal

import (
	"net"
	"testing"
)

// TestIsPublicImageIPRejectsSharedAndTranslatedRanges is the M5 regression.
// The SSRF filter used IsPrivate/IsGlobalUnicast, which pass CGNAT (100.64/10),
// benchmark (198.18/15), class E (240/4) and NAT64 (64:ff9b::/96) addresses.
func TestIsPublicImageIPRejectsSharedAndTranslatedRanges(t *testing.T) {
	blocked := []string{
		"100.64.0.1",
		"100.127.255.254",
		"198.18.0.1",
		"198.19.255.1",
		"240.0.0.1",
		"255.255.255.254",
		"64:ff9b::7f00:1",
		"64:ff9b::808:808",
	}
	for _, raw := range blocked {
		if isPublicImageIP(net.ParseIP(raw)) {
			t.Errorf("%s must not be fetchable as a public image host", raw)
		}
	}

	allowed := []string{
		"8.8.8.8",
		"100.63.255.1",
		"100.128.0.1",
		"198.20.0.1",
		"2606:4700:4700::1111",
	}
	for _, raw := range allowed {
		if !isPublicImageIP(net.ParseIP(raw)) {
			t.Errorf("%s is public and should stay fetchable", raw)
		}
	}
}
