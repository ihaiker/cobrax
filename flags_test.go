package cobrax_test

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"os"
	"testing"
)

type FlagTest1 struct {
	Name     string   `flag:"name" help:"姓名" short:"n" def:"def"`
	Address  []string `flag:"address" help:"用户地址"`
	Desc     string   `flag:"desc" help:"用户描述"`
	Debug    bool     `flag:"debug" short:"d" help:"否是测试"`
	FromJson string
	UserInfo struct {
		Name    string   `help:"姓名" env:"USER_NAME"`
		Address []string `flag:"address" help:"用户地址"`
	}
	Attrs map[string]string
}

type FlagTest2 struct {
	N1 string `flag:"n1" help:"name" short:"n" def:"t2 name"`
	B1 int    `flag:"b1" help:"test b1" def:"3"`
}

type TestConfig struct {
	T1 *FlagTest1
	T2 *FlagTest2
}

var config = &TestConfig{
	T1: &FlagTest1{
		Desc: "初始化值",
	},
	T2: new(FlagTest2),
}

var cmd = &cobra.Command{
	Use: "test",
	/*RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},*/
}

var t1 = &cobra.Command{
	Use: "t1",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := pretty.Println(config.T1)
		return err
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(" == PersistentPreRunE == ")
		return nil
	},
}

var t2 = &cobra.Command{
	Use: "t2",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := pretty.Println(config.T2)
		return err
	},
}

func init() {
	if err := cobrax.Flags(t1, config.T1, "t1", "TEST"); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	if err := cobrax.Flags(t2, config.T2, "t2", "TEST"); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	cmd.AddCommand(t1, t2)
}

func TestCmd(t *testing.T) {
	if err := cmd.Help(); err != nil {
		t.Fatal(err)
	}
}

func TestT1(t *testing.T) {
	os.Setenv("USER_NAME", "环境变量配置")
	os.Setenv("TEST_T1_NAME", "环境变量")
	os.Setenv("TEST_T1_DEBUG", "true")
	os.Setenv("TEST_T1_ADDRESS", "test,123,1")
	os.Setenv("TEST_T1_ATTRS", "test=123,b=a")
	os.Setenv("TEST_CONF", "/etc/config.json")

	os.Args = []string{
		"test",
		//"-f", "/data/user/",
		"t1",
		//"-h",
		"--t1.name", "命令行",
		//"--t1.user-info.name", "命令行设置",
		//"--t1.attrs", "name=1",
	}

	if err := cobrax.ConfigJson(cmd, config, "TEST_CONF", "./config.json", "/etc/config.json"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.Help(); err != nil {
		t.Fatal(err)
	}

	/*if err := cobrax.Config(cmd, func(cmd *cobra.Command) error {
		if cmd.Name() == t1.Name() {
			config.T1.Name = "--=文件=--"
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}*/

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestT2(t *testing.T) {
	if err := t2.Help(); err != nil {
		t.Fatal(err)
	}
}
