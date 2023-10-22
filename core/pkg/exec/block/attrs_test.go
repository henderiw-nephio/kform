package block

import (
	"testing"
)

func TestAttrs(t *testing.T) {
	cases := map[string]struct {
		attrs         Attrs
		wantedLoop    bool
		wantedCount   bool
		wantedForEach bool
	}{
		"NonePresent": {
			attrs: map[string]any{
				"a": "b",
				"c": "d",
			},
			wantedLoop:    false,
			wantedCount:   false,
			wantedForEach: false,
		},
		"CountPresent": {
			attrs: map[string]any{
				"a":     "b",
				"count": 4,
			},
			wantedLoop:    true,
			wantedCount:   true,
			wantedForEach: false,
		},
		"ForEachPresent": {
			attrs: map[string]any{
				"for_each": map[string]any{"x": "y"},
				"a":        "b",
			},
			wantedLoop:    true,
			wantedCount:   false,
			wantedForEach: true,
		},
		"BothPresent": {
			attrs: map[string]any{
				"for_each": map[string]any{"x": "y"},
				"a":        "b",
				"count":    4,
			},
			wantedLoop:    true,
			wantedCount:   true,
			wantedForEach: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var b bool
			b = tc.attrs.isLoopAttrPresent()
			if b != tc.wantedLoop {
				t.Errorf("want %t, got: %t", tc.wantedLoop, b)
			}
			b = tc.attrs.isCountAttrPresent()
			if b != tc.wantedCount {
				t.Errorf("want %t, got: %t", tc.wantedCount, b)
			}
			b = tc.attrs.isForEachAttrPresent()
			if b != tc.wantedForEach {
				t.Errorf("want %t, got: %t", tc.wantedForEach, b)
			}
		})
	}
}
