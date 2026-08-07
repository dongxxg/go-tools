package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"unitechs.com/unios-dice/uni-base/lib/stringx"
)

func TestFloat2Decimal(t *testing.T) {
	cases := []struct {
		a      float64
		expect float64
	}{
		{
			a:      0.12345,
			expect: 0.12,
		},
		{
			a:      2.1065,
			expect: 2.11,
		},
		{
			a:      1.9999,
			expect: 2.00,
		},
	}

	for _, each := range cases {
		t.Run(stringx.Rand(), func(t *testing.T) {
			actual := Float2Decimal(each.a)
			assert.Equal(t, each.expect, actual)
		})
	}
}
