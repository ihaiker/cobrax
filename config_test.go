package cobrax_test

import (
	"github.com/ihaiker/cobrax"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type TestConfigSuite struct {
	suite.Suite
}

func (p *TestConfigSuite) TestConfig() {
	type tls struct {
		Enable bool    `ngx:"enable" help:"Use TLS"`
		CaCert string  `ngx:"ca-cert" help:"Trust certs signed only by this CA" def:"~/.docker/ca.pem"`
		Cert   string  `ngx:"cert" help:"Path to TLS certificate file" def:"~/.docker/cert.pem"`
		Key    *string `ngx:"key" help:"Path to TLS key file" def:"~/.docker/key.pem"`
	}

	type config struct {
		Version         string   `ngx:"version" help:"docker version" json:"version"`
		DataRoot        *string  `ngx:"data-root" help:"docker data root"`
		DaemonJson      string   `ngx:"daemon-json" help:"docker cfg file"`
		RegistryMirrors []string `ngx:"registry-mirrors" help:"preferred Docker registry mirror"`
		StorageDriver   []string
		StraitVersion   bool `ngx:"strait-version" help:"Strict check DOCKER version if inconsistent will upgrade" def:"false"`
		TLS             *tls `ngx:"tls" flag:"tls"`

		Envs map[string]string
	}

	os.Args = []string{
		"test",
		"--conf", "./_testdata/config.json",
	}

	cmd := &cobra.Command{
		Use: "test", Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cfg := new(config)
	err := cobrax.ConfigJson(cmd, cobrax.GetFlags, cfg, "TEST_CONF", "./config.json", "/etc/config.json")
	p.Nil(err)

	p.Nil(cmd.Help())
	p.Nil(cmd.Execute())
	p.Equal("v1.21.2", cfg.Version)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}
