package cgroups

// ProcessMover objects implement the ability to move processes into process
// control groups.
type ProcessMover interface {
	// Move a process, identified byt its pid, to a control group, identified by
	// its name
	Move(pid int, cgroup string) error
}
