package robot

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRobot_IsSecretCompleted(t *testing.T) {
	ass := assert.New(t)
	end := "."

	tests := []struct {
		name        string
		secretParts []SecretPart
		expected    bool
	}{
		{
			name:        "empty robot",
			secretParts: nil,
			expected:    false,
		},
		{
			name: "single word without end marker",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
			},
			expected: false,
		},
		{
			name: "single word with end marker",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello."},
			},
			expected: true,
		},
		{
			name: "complete secret ordered",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
				{Index: 1, Word: "world."},
			},
			expected: true,
		},
		{
			name: "complete secret unordered",
			secretParts: []SecretPart{
				{Index: 1, Word: "world."},
				{Index: 0, Word: "hello"},
			},
			expected: true,
		},
		{
			name: "missing first word",
			secretParts: []SecretPart{
				{Index: 1, Word: "world."},
			},
			expected: false,
		},
		{
			name: "missing middle word",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
				{Index: 2, Word: "again."},
			},
			expected: false,
		},
		{
			name: "last word present but no end marker",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
				{Index: 1, Word: "world"},
			},
			expected: false,
		},
		{
			name: "end marker present but gap before",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
				{Index: 2, Word: "world."},
			},
			expected: false,
		},
		{
			name: "duplicate indexes with same word (idempotent case)",
			secretParts: []SecretPart{
				{Index: 0, Word: "hello"},
				{Index: 0, Word: "hello"},
				{Index: 1, Word: "world."},
			},
			expected: true,
		},
		{
			name: "non-zero starting index",
			secretParts: []SecretPart{
				{Index: 1, Word: "world."},
				{Index: 2, Word: "again."},
			},
			expected: false,
		},
		{
			name: "large secret complete",
			secretParts: []SecretPart{
				{Index: 0, Word: "this"},
				{Index: 1, Word: "is"},
				{Index: 2, Word: "a"},
				{Index: 3, Word: "complete"},
				{Index: 4, Word: "secret."},
			},
			expected: true,
		},
		{
			name: "large secret missing one part",
			secretParts: []SecretPart{
				{Index: 0, Word: "this"},
				{Index: 1, Word: "is"},
				{Index: 3, Word: "complete"},
				{Index: 4, Word: "secret."},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Robot{
				SecretParts:   tt.secretParts,
				LastUpdatedAt: time.Now(),
			}

			result := r.IsSecretCompleted(end)
			ass.Equal(tt.expected, result,
				"IsSecretCompleted() = %v, expected %v (secretParts=%v)",
				result,
				tt.expected,
				tt.secretParts)
		})
	}
}
