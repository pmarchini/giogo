package core

type CgroupManager interface {
	AddProcess(pid int) error // AddProcess adds a process to the cgroup
	Delete() error            // Delete deletes the cgroup
}
