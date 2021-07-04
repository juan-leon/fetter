package scanner

import (
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/juan-leon/fetter/pkg/cgroups"
	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

type ProcessScanner struct {
	ruleMap   map[string]string
	procMover cgroups.ProcessMover
}

func NewProcessScanner(config *settings.Settings, procMover cgroups.ProcessMover) *ProcessScanner {
	ruleMap := make(map[string]string)
	for _, r := range config.Rules {
		if r.Action == "execute" && r.Group != "" {
			for _, path := range r.Paths {
				if _, ok := ruleMap[path]; ok {
					log.Logger.Warnf("Path %s appears in several rules; ignoring", path)
				}
				ruleMap[path] = r.Group
			}
		}
	}
	return &ProcessScanner{
		ruleMap:   ruleMap,
		procMover: procMover,
	}
}

func (ps *ProcessScanner) Scan() {
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
		if group, ok := ps.ruleMap[exe]; ok {
			log.Logger.Debugf("Adding %s (pid %d) to cgroup %s", exe, p.Pid, group)
			ps.procMover.Move(int(p.Pid), group)
		}
	}
}

func (ps *ProcessScanner) Loop() {
	for {
		ps.Scan()
		// Scanning processes uses some CPU in heavily loaded machines, but long
		// delays between scans will increase the likelihood of processes
		// spawning children that are left out the control group (for instance,
		// a rule could be good to catch an IDE, but not its LSP subprocesses).
		// That is a problem better solved with the audit alternative.
		time.Sleep(time.Second)
	}
}
