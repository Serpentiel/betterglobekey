package feedback

import "testing"

type fakeNamer struct{ names map[string]string }

func (f fakeNamer) Name(id string) string { return f.names[id] }

func TestNotify(t *testing.T) {
	cases := []struct {
		name           string
		showCollection bool
		names          map[string]string
		source         string
		collection     string
		wantTitle      string
		wantSubtitle   string
	}{
		{
			name:           "resolves localized name",
			showCollection: true,
			names:          map[string]string{"com.apple.keylayout.US": "U.S."},
			source:         "com.apple.keylayout.US",
			collection:     "primary",
			wantTitle:      "U.S.",
			wantSubtitle:   "primary",
		},
		{
			name:           "falls back to id when unnamed",
			showCollection: true,
			names:          nil,
			source:         "com.apple.keylayout.US",
			collection:     "primary",
			wantTitle:      "com.apple.keylayout.US",
			wantSubtitle:   "primary",
		},
		{
			name:           "omits collection when disabled",
			showCollection: false,
			names:          map[string]string{"com.apple.keylayout.US": "U.S."},
			source:         "com.apple.keylayout.US",
			collection:     "primary",
			wantTitle:      "U.S.",
			wantSubtitle:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotTitle, gotSubtitle string

			called := false
			notifier := NewHUD(fakeNamer{names: tc.names}, tc.showCollection)
			notifier.show = func(title, subtitle string) {
				gotTitle, gotSubtitle, called = title, subtitle, true
			}

			notifier.Notify(tc.source, tc.collection)

			if !called {
				t.Fatal("show was not called")
			}

			if gotTitle != tc.wantTitle {
				t.Errorf("title = %q, want %q", gotTitle, tc.wantTitle)
			}

			if gotSubtitle != tc.wantSubtitle {
				t.Errorf("subtitle = %q, want %q", gotSubtitle, tc.wantSubtitle)
			}
		})
	}
}

// TestNewHUDWiresRealShow guards against NewHUD forgetting to set the show
// function, which would nil-panic on the first notification in production.
func TestNewHUDWiresRealShow(t *testing.T) {
	if NewHUD(fakeNamer{}, true).show == nil {
		t.Fatal("NewHUD left show nil")
	}
}
