package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/juan-leon/fetter/internal"
	"github.com/juan-leon/fetter/pkg/log"
)

var (
	configFile string
	daemonize  bool
	scan       bool

	// BuildDate is the date project was build.  Injected from linker
	BuildDate string
	// Commit is the commit the project was built from.  Injected from linker
	Commit string
	// Version is the tag (if any).  Injected from linker
	Version string
)

func assertUsage(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("too many args: %s", args)
	}
	if os.Geteuid() != 0 {
		// We could check for process capabilities (CAP_AUDIT_*, etc), but for
		// the time being let's keep it simple.
		return fmt.Errorf("this program needs root privileges")
	}
	return nil
}

func main() {
	log.InitConsoleLogger()
	root := &cobra.Command{
		Use:               "fetter",
		Short:             "Move processes into control groups based on configurable actions",
		PersistentPreRunE: assertUsage,
		Version:           fmt.Sprintf("%s, built on %s from %s\n", Version, BuildDate, Commit),
	}
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/fetter/config.yaml", "Path to configuration file")
	clean := &cobra.Command{
		Use:        "clean",
		Short:      "Delete fetter cgroups",
		Long:       "Delete fetter cgroups, moving any remaining process in them to root cgroup",
		Run:        func(cmd *cobra.Command, args []string) { internal.Clean(configFile) },
		SuggestFor: []string{"delete"},
	}
	run := &cobra.Command{
		Use:        "run",
		Short:      "Listen for rules defined in configuration and act accordlingly",
		Run:        func(cmd *cobra.Command, args []string) { internal.Loop(configFile, daemonize, scan) },
		SuggestFor: []string{"daemon"},
	}
	run.Flags().BoolVarP(&daemonize, "daemon", "d", false, "Fork to a daemonized process in background")
	run.Flags().BoolVarP(&scan, "scan", "s", false, "Scan already active processes according to rules")
	quickRun := &cobra.Command{
		Use:   "quick-run",
		Short: "Scan currently running processes according to rules and exit",
		Run:   func(cmd *cobra.Command, args []string) { internal.Scan(configFile) },
	}
	root.AddCommand(clean, run, quickRun)
	if err := root.Execute(); err != nil {
		os.Exit(2)
	}
}
