package cobrax

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

//根据paths查找相对应的配置文件，优先级从低到高
func Config(cmd *cobra.Command, flagGet FlagsGet, config interface{}, envName string, unmarshal func(data []byte, v interface{}) error, paths ...string) error {

	flags := flagGet(cmd)

	help := "the global config file path."
	if envName != "" {
		help += fmt.Sprintf("(env: %s)", envName)
	}
	flags.StringSliceVarP(&paths, "conf", "f", paths, help)

	err := cmd.ParseFlags(os.Args[1:])
	if err != nil {
		return err
	}

	//ParseFlags和Find都是只能运行一次，所以必须不能这样
	if help, _ := flags.GetBool("help"); help {
		return nil
	}

	if err := getValue(flags, "conf", envName)(); err != nil {
		return err
	}

	flag := flags.Lookup("conf")
	if paths == nil || len(paths) == 0 {
		return nil
	}
	for i := len(paths) - 1; i >= 0; i-- {
		bs, err := ioutil.ReadFile(paths[i])
		if flag.Changed && err != nil {
			return err
		}
		if err != nil {
			continue
		}
		return unmarshal(bs, config)
	}
	return nil
}

func ConfigJson(cmd *cobra.Command, flagGet FlagsGet, config interface{}, env string, paths ...string) error {
	return Config(cmd, flagGet, config, env, json.Unmarshal, paths...)
}
