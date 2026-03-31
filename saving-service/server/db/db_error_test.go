package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundErr_Error(t *testing.T) {
	err := NotFound("ExchangeRate", "er-uuid-123")
	assert.Equal(t, "ExchangeRate with id 'er-uuid-123' not found", err.Error())
}

func TestNotFoundErr_Fields(t *testing.T) {
	err := NotFound("Branch", "br-456")
	assert.Equal(t, "Branch", err.ResourceType)
	assert.Equal(t, "br-456", err.ID)
}
