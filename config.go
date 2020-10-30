package cobrax

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

func Config(cmd *cobra.Command, fn func(*cobra.Command) error) error {
	r, _, err := cmd.Find(os.Args[1:])
	if err != nil {
		return err
	} else if err := r.ParseFlags(os.Args[1:]); err != nil {
		if strings.Contains(err.Error(), "help requested") {
			return nil
		}
		return err
	}
	if err := fn(r); err != nil {
		return err
	}
	return r.ParseFlags(os.Args[1:])
}

//根据paths查找相对应的配置文件，优先级从低到高
func ConfigFrom(cmd *cobra.Command, config interface{}, envName string, unmarshal func(data []byte, v interface{}) error, paths ...string) error {
	help := "the global config file path."
	if envName != "" {
		help += fmt.Sprintf("(env: %s)", envName)
	}
	cmd.PersistentFlags().StringSliceVarP(&paths, "conf", "f", paths, help)

	return Config(cmd, func(c *cobra.Command) error {
		if err := envget(cmd, "conf", envName)(); err != nil {
			return err
		}
		flag := cmd.PersistentFlags().Lookup("conf")
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
	})
}

func ConfigJson(cmd *cobra.Command, config interface{}, env string, paths ...string) error {
	return ConfigFrom(cmd, config, env, json.Unmarshal, paths...)
}
