package accessibility

import "testing"

// TestTrustedIsCallable exercises the Accessibility query. Its result depends on
// the host's grant (false in a plain test process, true under a granted daemon),
// so only its callability is asserted. Prompt is intentionally not exercised, as
// it would present a blocking system dialog.
func TestTrustedIsCallable(_ *testing.T) {
	_ = Trusted()
}
