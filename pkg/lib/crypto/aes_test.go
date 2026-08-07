package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAES(t *testing.T) {
	s, err := AesEncryptString("zyuc", "ceshi!23ceshi!23")
	assert.NoError(t, err)
	assert.Equal(t, "Fd4rYAISWJ3c8NGtj9yvrg==", s)

	s2, err := AesDecryptString("erFElY7HQpFxDIxQAEN/SQ==", "ceshi!23")
	assert.NoError(t, err)
	assert.Equal(t, "zyuc", s2)
}
