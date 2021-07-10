package cobrax

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	name := camel2Case("FlagNameAge", '-')
	assert.Equal(t, name, "flag-name-age")
}

func TestEnv(t *testing.T) {
	name := envName("FlagNameAge")
	assert.Equal(t, name, "FLAG_NAME_AGE")
}
