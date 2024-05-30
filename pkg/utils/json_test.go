package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintJSON(t *testing.T) {
	testCases := []struct {
		name        string
		json        string
		expected    string
		expectedErr string
	}{
		{
			name:        "empty",
			json:        "",
			expectedErr: "unexpected end of JSON input",
		},
		{
			name: "simple",
			json: `{"foo": "bar"}`,
			expected: `{
  "foo": "bar"
}`,
		},
		{
			name: "complex",
			json: `{"foo": "bar", "baz": {"qux": "quux"}}`,
			expected: `{
  "foo": "bar",
  "baz": {
    "qux": "quux"
  }
}`,
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PrettyPrintJSON([]byte(tc.json))

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}

			assert.Equal(t, tc.expected, string(got))
		})
	}
}

func TestPrettyPrintToJSON(t *testing.T) {
	testCases := []struct {
		name        string
		obj         interface{}
		expected    string
		expectedErr string
	}{
		{
			name:     "empty",
			obj:      nil,
			expected: "null",
		},
		{
			name: "simple",
			obj:  map[string]string{"foo": "bar"},
			expected: `{
  "foo": "bar"
}`,
		},
		{
			name: "complex",
			obj: map[string]interface{}{
				"foo": "bar",
				"baz": map[string]string{"qux": "quux"},
			},
			expected: `{
  "baz": {
    "qux": "quux"
  },
  "foo": "bar"
}`,
		},
		{
			name:        "invalid",
			obj:         make(chan int),
			expected:    "",
			expectedErr: "json: unsupported type: chan int",
		},
		{
			name: "containing HTML escape characters",
			obj:  map[string]interface{}{"foo": `A~@#$%^&*\!\() _+=-;:\"/?.>,<%Z`},
			expected: `{
  "foo": "A~@#$%^&*\\!\\() _+=-;:\\\"/?.>,<%Z"
}`,
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PrettyPrintToJSON(tc.obj)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}
