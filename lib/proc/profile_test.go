package proc

import (
	"strings"
	"testing"

	"github.com/dongxxg/go-tools/log"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	var buf strings.Builder
	log.SetOutput(&buf)
	profiler := StartProfile()
	// start again should not work
	assert.NotNil(t, StartProfile())
	profiler.Stop()
	// stop twice
	profiler.Stop()
	assert.True(t, strings.Contains(buf.String(), ".pprof"))
}