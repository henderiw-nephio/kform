package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCelExpression(t *testing.T) {
	cases := map[string]struct {
		testString string
		wanted     bool
	}{
		"Underscore": {
			testString: "a_b_c_d",
			wanted:     false,
		},
		"Dash": {
			testString: "a-b-c-d",
			wanted:     false,
		},
		"DashSeperated": {
			testString: "a + b",
			wanted:     true,
		},
		"&": {
			testString: "$a.b",
			wanted:     true,
		},
		"()": {
			testString: "size(a)",
			wanted:     true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := IsCelExpression(tc.testString)
			if err != nil {
				assert.Error(t, err)
			}

			if b != tc.wanted {
				t.Errorf("want %t, got: %t", tc.wanted, b)
			}
		})
	}
}
