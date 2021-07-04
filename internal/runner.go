package internal

import (
	"time"

	"github.com/sevlyar/go-daemon"

	"github.com/juan-leon/fetter/pkg/audit"
	"github.com/juan-leon/fetter/pkg/cgroups"
	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/scanner"
	"github.com/juan-leon/fetter/pkg/settings"
	"github.com/juan-leon/fetter/pkg/triggers"
)

// Loop implements the run subcommand.  This command never returns, unless
// daemonize is true (the parent process will return, but the child will enter
// in same loop)
func Loop(
	configFile string,
	daemonize bool,
	scan bool,
) {
	config := loadConfig(configFile)
	if daemonize {
		cntxt := &daemon.Context{
			PidFileName: "/run/fetter.pid",
		}
		child, err := cntxt.Reborn()
		if err != nil {
			log.Console.Fatalf("Unable to daemonize: %s", err)
		}
		if child != nil {
			log.Console.Infof("Dettaching")
			return
		}
		defer cntxt.Release()
	}
	log.InitFileLogger(config.Logging)
	log.Logger.Infof("Initializing Control Groups...")
	groups := cgroups.NewGroupHierarchy(config)
	if config.Mode == settings.RunModeScanner {
		log.Logger.Infof("Scanning active processes...")
		s := scanner.NewProcessScanner(config, groups)
		s.Loop()
	} else {
		log.Logger.Infof("Auditing system calls according to rules...")
		s := audit.NewSysCallListener(config, groups, triggers.NewTriggerRunner(config))
		if s == nil {
			log.Logger.Fatalf("Could not setup a kernel syscall listener")
		}
		if scan {
			go func() {
				// The sleep here if to avoid (unlikely) race conditions between
				// receiving audit events and process spawning
				time.Sleep(time.Second)
				log.Logger.Infof("Scanning already active processes...")
				scanner.NewProcessScanner(config, groups).Scan()
			}()
		}
		s.Loop()
	}
}

// Clean implements the clean subcommand
func Clean(configFile string) {
	config := loadConfig(configFile)
	log.InitFileLogger(config.Logging)
	cgroups.DeleteGroupHierarchy(config)
}

// Scan implements the quick-run subcommand
func Scan(configFile string) {
	config := loadConfig(configFile)
	log.InitFileLogger(config.Logging)
	log.Logger.Infof("Initializing Control Groups...")
	groups := cgroups.NewGroupHierarchy(config)
	log.Logger.Infof("Scanning active processes...")
	scanner.NewProcessScanner(config, groups).Scan()
}

func loadConfig(configFile string) (config *settings.Settings) {
	config, err := settings.Load(configFile)
	if err != nil {
		log.Console.Fatalf("Could not read config: %s", err)
	}
	return
}
