package tlsutil

import "testing"

func TestTLSCipherSuits(t *testing.T) {
	suits := TLSCipherSuits()
	if len(suits) <= 0 {
		t.Fatalf("Get TLSCipherSuits failed")
	}
}
