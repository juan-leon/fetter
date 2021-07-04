package cgroups

import (
	"runtime"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

func emptySpec() *specs.LinuxResources {
	return &specs.LinuxResources{}
}

var period = uint64(1000000)
var NumCPUs = runtime.NumCPU()

func createSpec(name string, g *settings.Group) (spec *specs.LinuxResources) {
	spec = emptySpec()
	if g.CPU > 0 {
		spec.CPU = specCPU(g.CPU)
	}
	if g.RAM > 0 {
		spec.Memory = specRAM(g.RAM)
	}
	if g.Pids > 0 {
		spec.Pids = specPids(g.Pids)
	}
	log.Logger.Debugw("CGroup spec created", "cgroup", name, "spec", spec)
	return
}

func specCPU(cpu int) (spec *specs.LinuxCPU) {
	quota := int64(uint64(cpu) * period * uint64(NumCPUs) / 100)
	spec = &specs.LinuxCPU{
		Quota:  &quota,
		Period: &period,
	}
	return
}

func specRAM(ram int64) (spec *specs.LinuxMemory) {
	bytes := ram * 1024 * 1024
	spec = &specs.LinuxMemory{
		Limit: &bytes,
	}
	return
}

func specPids(pids int64) (spec *specs.LinuxPids) {
	spec = &specs.LinuxPids{Limit: pids}
	return
}
