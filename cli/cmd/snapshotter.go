/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/NVIDIA/cloud-native-stack/cli/pkg/snapshotter"
	"github.com/spf13/cobra"
)

// snapshotterCmd represents the snapshotter command
var snapshotterCmd = &cobra.Command{
	Use:   "snapshotter",
	Short: "Snapshot the current environment",
	Long:  `Snapshot the current environment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns := snapshotter.NodeSnapshotter{}
		return ns.Run(nil)
	},
}

func init() {
	rootCmd.AddCommand(snapshotterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapshotterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// snapshotterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
