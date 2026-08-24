package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	pw := strings.Join([]string{"zyuctest", "zyuc123"}, " ")
	h, err := Hash(pw)
	assert.Nil(t, err)
	assert.Nil(t, VerifyPassword(string(h), pw))
}
