// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/spf13/cobra"
)

const (
	flagFormat = "format"
)

func init() {
	RootCmd.AddCommand(showCmd)
	addJsonnetFlagsToCmd(showCmd)
	addKubectlFlagsToCmd(showCmd)
	addEnvCmdFlags(showCmd)
	showCmd.PersistentFlags().StringP(flagFormat, "o", "yaml", "Output format.  Supported values are: json, yaml")
}

var showCmd = &cobra.Command{
	Use:   "show [<env>|-f <file-or-dir>]",
	Short: "Show expanded resource definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		var err error

		c := kubecfg.ShowCmd{}

		files, err := getFiles(cmd, args)
		if err != nil {
			return err
		}

		vm, err := newExpander(cmd)
		if err != nil {
			return err
		}

		c.Objs, err = vm.Expand(files)
		if err != nil {
			return err
		}

		c.Format, err = flags.GetString(flagFormat)
		if err != nil {
			return err
		}

		return c.Run(cmd.OutOrStdout())
	},
}
