package cobrax

import "testing"

func TestName(t *testing.T) {
	name := camel2Case("FlagName.Age", '-')
	t.Log(name)
}

func TestEnv(t *testing.T) {
	name := envName("FlagName.Age")
	t.Log(name)
}
