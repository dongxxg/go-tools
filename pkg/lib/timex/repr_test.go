package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReprOfDuration(t *testing.T) {
	assert.Equal(t, "1000.00ms", ReprOfDuration(time.Second))
	assert.Equal(t, "1111.55ms", ReprOfDuration(
		time.Second+time.Millisecond*111+time.Microsecond*555))
}
