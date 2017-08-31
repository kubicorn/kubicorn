// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"io"

	"github.com/kris-nova/kubicorn/cutil/local"
	"github.com/kris-nova/kubicorn/cutil/logger"
	lol "github.com/kris-nova/lolgopher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	bashCompletionFunc = `
__kubicorn_parse_list()
{
    local kubicorn_out
    if kubicorn_out=$(kubicorn list --no-headers 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kubicorn_out[*]}" -- "$cur" ) )
    fi
}
__kubicorn_parse_profiles()
{
    local kubicorn_out
    if kubicorn_out=(amazon aws digitalocean do); then
        COMPREPLY=( $( compgen -W "${kubicorn_out[*]}" -- "$cur" ) )
    fi
}
__custom_func() {
    case ${last_command} in
        kubicorn_apply | kubicorn_create | kubicorn_delete | kubicorn_getconfig)
            __kubicorn_parse_list
            return
            ;;
        *)
            ;;
    esac
}
`
)

var cfgFile = local.Expand("~/.kubicorn")

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubicorn",
	Short: "Kubernetes cluster management, without any magic",
	Long: fmt.Sprintf(`
%s
`, Unicorn),
	Run: func(cmd *cobra.Command, args []string) {
		if logger.Fabulous {
			cmd.SetOutput(logger.FabulousWriter)
		}
		if os.Getenv("KUBICORN_TRUECOLOR") != "" {
			cmd.SetOutput(&lol.Writer{Output: os.Stdout, ColorMode: lol.ColorModeTrueColor})
		}
		cmd.Help()
	},
	BashCompletionFunction: bashCompletionFunc,
}

type Options struct {
	StateStore     string
	StateStorePath string
	Name           string
	CloudId        string
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initViper)

	//flags here
	addPersistentFlagInt(RootCmd, &logger.Level, "verbose", "v", 3, "Log level")
	addPersistentFlagBool(RootCmd, &logger.Color, "color", "C", true, "Toggle colorized logs")
	addPersistentFlagBool(RootCmd, &logger.Fabulous, "fab", "f", false, "Toggle colorized logs")

	registerEnvironmentalVariables()
}

// initViper initializes viper to handle configuration.
func initViper() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("KUBICORN")
	viper.AutomaticEnv()

	if _, err := os.Stat(cfgFile); err != nil {
		logger.Debug("unable to find kubicorn configuration")
		err := writeConfig()
		if err != nil {
			logger.Critical("unable to create kubicorn configuration")
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		logger.Critical("unable to read kubicorn configuration")
		os.Exit(1)
	}
}

// configFileWriter creates the configuration file and returns writer.
func configFileWriter() (io.WriteCloser, error) {
	f, err := os.Create(cfgFile)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(cfgFile, 0644); err != nil {
		return nil, err
	}

	return f, nil
}

// writeConfig writes defaults to the configuration file.
func writeConfig() error {
	f, err := configFileWriter()
	if err != nil {
		return err
	}
	defer f.Close()

	c, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("unable to export configuration")
	}

	_, err = f.Write(c)
	if err != nil {
		return fmt.Errorf("unable to write configuration")
	}

	return nil
}

// registerEnvironmentalVariables bind environmental variables to appropriate flags.
func registerEnvironmentalVariables() {
	viper.BindEnv("state-store", "STATE_STORE")
	viper.BindEnv("state-store-path", "STATE_STORE_PATH")
	viper.BindEnv("profile", "PROFILE")
}
