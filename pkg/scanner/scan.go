package scanner

import (
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/juan-leon/fetter/pkg/cgroups"
	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

type ProcessScanner struct {
	config  *settings.Settings
	ruleMap map[string]string
	cgroups *cgroups.GroupHierarchy
}

func NewProcessScanner(config *settings.Settings, cgroups *cgroups.GroupHierarchy) *ProcessScanner {
	ruleMap := make(map[string]string)
	for _, r := range config.Rules {
		if r.Action == "execute" {
			ruleMap[r.Path] = r.Group
		}
	}
	return &ProcessScanner{
		config:  config,
		ruleMap: ruleMap,
		cgroups: cgroups,
	}
}

func (pc *ProcessScanner) Scan() {
	processes, err := process.Processes()
	if err != nil {
		log.Logger.Fatalf("Cannot scan processes %s", err)
	}
	for _, p := range processes {
		exe, err := p.Exe()
		if err != nil {
			// Typically, condition races related to short lived processes
			continue
		}
		if group, ok := pc.ruleMap[exe]; ok {
			log.Logger.Debugf("Adding %s (pid %d) to cgroup %s", exe, p.Pid, group)
			pc.cgroups.Add(int(p.Pid), group)
		}
	}
}

func (pc *ProcessScanner) Loop() {
	for {
		pc.Scan()
		// Scanning processes use some CPU in heavily loaded machines, but long
		// delays between scans will increase the likelihood of processes
		// spawning children that are left out the control group (for instance,
		// a rule could be good to catch an IDE, but not its LSP subprocesses).
		// That is a problem better solved with the audit alternative.
		time.Sleep(time.Second)
	}
}
