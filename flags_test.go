package cobrax_test

import (
	"github.com/ihaiker/cobrax"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type tls struct {
	Enable bool    `ngx:"enable" help:"Use TLS"`
	CaCert string  `ngx:"ca-cert" help:"Trust certs signed only by this CA" def:"~/.docker/ca.pem"`
	Cert   string  `ngx:"cert" help:"Path to TLS certificate file" def:"~/.docker/cert.pem"`
	Key    *string `ngx:"key" help:"Path to TLS key file" def:"~/.docker/key.pem"`
}

type config struct {
	Version         string   `ngx:"version" help:"docker version"`
	DataRoot        *string  `ngx:"data-root" help:"docker data root"`
	DaemonJson      string   `ngx:"daemon-json" help:"docker cfg file"`
	RegistryMirrors []string `ngx:"registry-mirrors" help:"preferred Docker registry mirror"`
	StorageDriver   []string
	StraitVersion   bool `ngx:"strait-version" help:"Strict check DOCKER version if inconsistent will upgrade" def:"false"`
	TLS             *tls `ngx:"tls" flag:"tls"`

	Envs map[string]string
}

type TestFlagsSuite struct {
	suite.Suite
	cmd *cobra.Command
}

func (p *TestFlagsSuite) SetupTest() {
	p.cmd = &cobra.Command{
		Use: "test", Run: func(cmd *cobra.Command, args []string) {
		},
	}
}

func (p *TestFlagsSuite) flags(prefix, envPrefix string) *config {
	main := new(config)
	err := cobrax.Flags(p.cmd, main, prefix, envPrefix)
	p.Nil(err)
	return main
}

func (p *TestFlagsSuite) TestShowUsage() {
	p.flags("", "")
	p.Nil(p.cmd.Usage())
}

func (p *TestFlagsSuite) TestArgments() {
	os.Args = []string{
		"cobrax",
		"--tls.enable",
		"--data-root", "/var/lib/docker",
		"--daemon-json", "/etc/docker/daemon.json",
		"--envs", "key=value",
	}
	conf := p.flags("", "")
	p.Nil(p.cmd.Execute())
	p.True(conf.TLS.Enable)
	p.Equal("/var/lib/docker", *conf.DataRoot)
	p.Equal("value", (conf.Envs)["key"])
}

func (p TestFlagsSuite) TestEnvs() {
	p.Nil(os.Setenv("VERSION", "v1.0"))
	os.Args = []string{
		"cobrax",
	}
	conf := p.flags("", "")
	p.Nil(p.cmd.Execute())
	p.Equal("v1.0", conf.Version)
}

func TestSetFlags(t *testing.T) {
	suite.Run(t, new(TestFlagsSuite))
}
