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
	"os"

	"github.com/spf13/cobra"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
)

const flagDiffStrategy = "diff-strategy"

func init() {
	addJsonnetFlagsToCmd(diffCmd)
	addKubectlFlagsToCmd(diffCmd)
	addEnvCmdFlags(diffCmd)
	diffCmd.PersistentFlags().String(flagDiffStrategy, "all", "Diff strategy, all or subset.")
	RootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff [<env>|-f <file-or-dir>]",
	Short: "Display differences between server and local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		var err error

		c := kubecfg.DiffCmd{}

		c.DiffStrategy, err = flags.GetString(flagDiffStrategy)
		if err != nil {
			return err
		}

		c.Environment, c.Files, err = parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		c.Expander, err = newExpander(cmd)
		if err != nil {
			return err
		}

		c.ClientPool, c.Discovery, err = restClientPool(cmd)
		if err != nil {
			return err
		}

		c.DefaultNamespace, _, err = clientConfig.Namespace()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return c.Run(metadata.AbsPath(cwd), cmd.OutOrStdout())
	},
	Long: `Display differences between server and local configuration.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.`,
	Example: `  # Show diff between resources described in a local ksonnet application and
  # the cluster referenced by the 'dev' environment. Can be used in any
  # subdirectory of the application.
  ksonnet diff -e=dev

  # Show diff between resources described in a YAML file and the cluster
  # referenced in '$KUBECONFIG'.
  ksonnet diff -f ./pod.yaml

  # Show diff between resources described in a YAML file and the cluster
  # referred to by './kubeconfig'.
  ksonnet diff --kubeconfig=./kubeconfig -f ./pod.yaml`,
}
