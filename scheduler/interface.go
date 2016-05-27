// Package scheduler defines the interface with schedulers & orchestration systems
package scheduler

import (
	"github.com/microscaling/microscaling/demand"
)

// Scheduler is the interface to schedulers / orchestration systems
type Scheduler interface {
	// InitScheduler creates and starts the app identified by appId
	InitScheduler(task *demand.Task) error

	// StopStartTasks changes the count of containers to match task.Demand
	StopStartTasks(tasks *demand.Tasks) error

	// CountAllTasks updates task.Running to tell us how many instances of each task are currently running
	CountAllTasks(tasks *demand.Tasks) error

	// Cleanup is called to give the scheduler a chance to clean up
	Cleanup() error
}
