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
	"github.com/spf13/cobra"

	"github.com/bitnami/kubecfg/pkg/kubecfg"
)

const (
	flagDiffStrategy  = "diff-strategy"
	flagNoErrorOnDiff = "no-error-on-diff"
	flagOmitSecrets   = "omit-secrets"
)

func init() {
	diffCmd.PersistentFlags().String(flagDiffStrategy, "all", "Diff strategy, all or subset.")
	diffCmd.PersistentFlags().Bool(flagNoErrorOnDiff, false, "don't exit with error if there are differences")
	diffCmd.PersistentFlags().Bool(flagOmitSecrets, false, "hide secret details when showing diff")
	RootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Display differences between server and local config",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		var err error

		c := kubecfg.DiffCmd{}

		c.DiffStrategy, err = flags.GetString(flagDiffStrategy)
		if err != nil {
			return err
		}

		c.NoErrorOnDiff, err = flags.GetBool(flagNoErrorOnDiff)
		if err != nil {
			return err
		}

		c.OmitSecrets, err = flags.GetBool(flagOmitSecrets)
		if err != nil {
			return err
		}

		c.Client, c.Mapper, _, err = getDynamicClients(cmd)
		if err != nil {
			return err
		}

		c.DefaultNamespace, err = defaultNamespace(clientConfig)
		if err != nil {
			return err
		}

		objs, err := readObjs(cmd, args)
		if err != nil {
			return err
		}

		return c.Run(cmd.Context(), objs, cmd.OutOrStdout())
	},
}
