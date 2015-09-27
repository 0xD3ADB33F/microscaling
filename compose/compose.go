// Schedule using docker compose
package compose

import (
	"bitbucket.org/force12io/force12-scheduler/demand"
	"bitbucket.org/force12io/force12-scheduler/scheduler"
)

type ComposeScheduler struct {
}

func NewScheduler() *ComposeScheduler {
	return &ComposeScheduler{}
}

// compile-time assert that we implement the right interface
var _ scheduler.Scheduler = (*ComposeScheduler)(nil)

func (c *ComposeScheduler) InitScheduler(appId string) error {
	return nil
}

func (c *ComposeScheduler) StopStartNTasks(appId string, family string, demandcount int, currentcount int) error {
	return nil
}

func (c *ComposeScheduler) CountTaskInstances(taskName string, task demand.Task) (int, int, error) {
	return 0, 0, nil
}
