package cgroups

type ProcessMover interface {
	Move(pid int, cgroup string) error
}
