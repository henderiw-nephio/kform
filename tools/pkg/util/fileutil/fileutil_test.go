package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	cases := map[string]struct {
		path        string
		isDir       bool
		expectedErr bool
	}{
		"CurrentDir": {
			path:        ".",
			expectedErr: true,
		},
		"PreviousDir": {
			path:        "..",
			expectedErr: true,
		},
		"Relative": {
			path:        "./example",
			expectedErr: false,
			isDir:       true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isDir, err := ValidatePath(tc.path)
			if err != nil {
				if tc.expectedErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				return
			}
			if err == nil && tc.expectedErr {
				t.Errorf("want error, got nil")
			}
			if isDir != tc.isDir {
				t.Errorf("want %t, got: %t", tc.isDir, isDir)
			}
		})
	}
}
