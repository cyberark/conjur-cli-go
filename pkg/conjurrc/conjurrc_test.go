package conjurrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConjurrc(t *testing.T) {
	data := generateConjurrc("test-account", "test-appliance-url")

	assert.Equal(t, string(data), `account: test-account
appliance_url: test-appliance-url
plugins: []
`)
}
