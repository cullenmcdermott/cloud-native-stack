/*
Copyright © 2025 NVIDIA Corporation
SPDX-License-Identifier: Apache-2.0
*/
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information including commit hash and build date.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf(`%s (commit: %s date: %s)`, version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
