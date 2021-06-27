package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/juan-leon/fetter/internal"
	"github.com/juan-leon/fetter/pkg/log"
)

const (
	Name  = "fetter"
	Short = "Move processes into control groups based on configurable actions"
)

var (
	configFile  string
	daemonize   bool
	scan        bool
	scanAndExit bool
	clean       bool

	BuildDate string // injected from linker
	Commit    string // injected from linker
	Version   string // injected from linker
)

func run(cmd *cobra.Command, args []string) {
	if clean {
		internal.Clean(configFile)
		return
	}
	if scanAndExit {
		internal.Scan(configFile)
		return
	}
	internal.Loop(configFile, daemonize, scan)
}

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
	cmd := &cobra.Command{
		Use:     Name,
		Short:   Short,
		Run:     run,
		PreRunE: assertUsage,
		Version: fmt.Sprintf("%s, built on %s from %s\n", Version, BuildDate, Commit),
	}
	f := cmd.Flags()
	f.StringVarP(&configFile, "config", "c", "/etc/fetter/config.yaml", "Path to configuration file")
	f.BoolVarP(&daemonize, "daemon", "d", false, "Fork to a daemonized process in background")
	f.BoolVarP(&scan, "scan", "s", false, "Scan already active processes according to rules")
	f.BoolVarP(&clean, "clean-up", "D", false, "Delete cgroups and exit")
	f.BoolVarP(&scanAndExit, "scan-and-exit", "S", false, "Scan processes according to rules and exit")
	if err := cmd.Execute(); err != nil {
		os.Exit(2)
	}
}
