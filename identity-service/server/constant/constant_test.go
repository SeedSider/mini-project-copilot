package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "00", ResponseCodeSuccess)
	assert.Equal(t, "2006-01-02", FORMAT_DATE)
	assert.Equal(t, "2006-01-02 15:04:05", FORMAT_DATE_TIME)
	assert.Equal(t, "process_id", ProcessIdCtx)
}
