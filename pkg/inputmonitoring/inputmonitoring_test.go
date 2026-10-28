package inputmonitoring

import "testing"

// TestGrantedIsCallable exercises the Input Monitoring query. Its result depends
// on the host's grant, so only its callability is asserted.
func TestGrantedIsCallable(_ *testing.T) {
	_ = granted()
}
