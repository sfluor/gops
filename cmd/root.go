package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sfluor/gops/watcher"
	"github.com/spf13/cobra"
)

var output string
var json, noChildren bool
var interval, duration time.Duration

func init() {
	rootCmd.Flags().StringVarP(&output, "output", "o", "record", "Output file name")
	rootCmd.Flags().BoolVarP(&json, "json", "j", false, "JSON output format")
	rootCmd.Flags().BoolVarP(&noChildren, "no-children", "n", false, "Don't retrieve process' children metrics")
	rootCmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "Interval between two polls")
	rootCmd.Flags().DurationVarP(&duration, "duration", "d", 24*time.Hour, "Duration of the watch")
}

var rootCmd = &cobra.Command{
	Use:   "gops [pid]",
	Short: "Gops is a tool to monitor a specific process ressources",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid PID: %v\n", args[0])
			os.Exit(1)
		}
		watcher.Watch(pid, interval, duration, noChildren, output, json)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
