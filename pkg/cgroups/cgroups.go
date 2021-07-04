package cgroups

import (
	"fmt"
	"syscall"

	"github.com/containerd/cgroups"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

const (
	kill = "KILL" // pseudo group for killing proceses outright
)

type GroupHierarchy struct {
	name      string
	main      cgroups.Cgroup
	subgroups map[string]cgroups.Cgroup
}

func NewGroupHierarchy(config *settings.Settings) *GroupHierarchy {
	main, err := cgroups.New(
		cgroups.V1,
		cgroups.StaticPath(config.Name),
		emptySpec(),
	)
	if err != nil {
		log.Logger.Fatalf("Could not create base cgroup with name %s: %s", config.Name, err)
		return nil
	}
	gh := GroupHierarchy{
		name:      config.Name,
		main:      main,
		subgroups: make(map[string]cgroups.Cgroup),
	}
	for name, g := range config.Groups {
		gh.addSubGroup(name, g)
	}
	return &gh
}

func DeleteGroupHierarchy(config *settings.Settings) error {
	main, err := cgroups.Load(
		cgroups.V1,
		cgroups.StaticPath(config.Name),
	)
	if err != nil {
		log.Logger.Errorf("Could not load base cgroup with name %s: %s", config.Name, err)
		return err
	}
	root, err := cgroups.Load(
		cgroups.V1,
		cgroups.StaticPath(""),
	)
	if err != nil {
		log.Logger.Errorf("Could not load root cgroup: %s", err)
	} else {
		err = main.MoveTo(root)
		if err != nil {
			log.Logger.Errorf("Could not move all processes out of %s: %s", config.Name, err)
		}
	}
	if err := main.Delete(); err != nil {
		log.Logger.Errorf("Could not delete base cgroup with name %s: %s", config.Name, err)
		return err
	}
	return nil
}

func (gh *GroupHierarchy) Move(pid int, cgroup string) error {
	if cgroup == kill {
		log.Logger.Infof("Killing process %d", pid)
		if err := syscall.Kill(pid, 9); err != nil {
			log.Logger.Warnf("Could not kill process: %s", err)
			return err
		}
		return nil
	}
	log.Logger.Infof("Adding process %d to cgroup %s", pid, cgroup)
	if subgroup, ok := gh.subgroups[cgroup]; ok {
		if err := subgroup.Add(cgroups.Process{Pid: pid}); err != nil {
			log.Logger.Warnw("Could not add process to subgroup", "name", cgroup, "pid", pid)
			return err
		}
	} else {
		log.Logger.Warnw("Did not find subgroup", "name", cgroup, "pid", pid)
	}
	return nil
}

func (gh *GroupHierarchy) addSubGroup(name string, g settings.Group) error {
	if name == "" {
		err := fmt.Errorf("could not create subgroup with empty name")
		log.Logger.Errorf("%s", err)
		return err
	}
	subgroup, err := gh.main.New(name, createSpec(name, &g))
	if err != nil {
		log.Logger.Errorf("Could not create subgroup with name %s: %s", name, err)
		return err
	}
	gh.subgroups[name] = subgroup
	if g.Freeze {
		if err := subgroup.Freeze(); err != nil {
			log.Logger.Errorf("Could not freeze %s: %s", name, err)
			return err
		}
	}
	log.Logger.Debugw("Added subgroup", "name", name, "subgroup", g)
	return nil
}
