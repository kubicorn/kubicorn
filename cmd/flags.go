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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// flagApplyAnnotations applies Bash completion annotations.
func flagApplyAnnotations(cmd *cobra.Command, flag, completion string) {
	if cmd.Flag(flag) != nil {
		if cmd.Flag(flag).Annotations == nil {
			cmd.Flag(flag).Annotations = map[string][]string{}
		}
		cmd.Flag(flag).Annotations[cobra.BashCompCustom] = append(
			cmd.Flag(flag).Annotations[cobra.BashCompCustom],
			completion,
		)
	}
}

// addPersistentFlagInt is used to add an integer persistent flag and register it with viper.
func addPersistentFlagInt(cmd *cobra.Command, field *int, name, shorthand string, value int, usage string) {
	cmd.PersistentFlags().IntVarP(field, name, shorthand, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

// addPersistentFlagBool is used to add a bool persistent flag and register it with viper.
func addPersistentFlagBool(cmd *cobra.Command, field *bool, name, shorthand string, value bool, usage string) {
	cmd.PersistentFlags().BoolVarP(field, name, shorthand, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

// addFlagString is used to add a string flag and register it with viper.
func addFlagString(cmd *cobra.Command, field *string, name, shorthand, value string, usage string) {
	cmd.PersistentFlags().StringVarP(field, name, shorthand, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

// addFlagString is used to add an integer flag and register it with viper.
func addFlagBool(cmd *cobra.Command, field *bool, name, shorthand string, value bool, usage string) {
	cmd.PersistentFlags().BoolVarP(field, name, shorthand, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}
