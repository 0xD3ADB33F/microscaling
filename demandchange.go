package main

import (
	"log"

	"bitbucket.org/force12io/force12-scheduler/demand"
	"bitbucket.org/force12io/force12-scheduler/scheduler"
)

// handleDemandChange checks the new demand
func handleDemandChange(input demand.Input, s scheduler.Scheduler, ts map[string]demand.Task) error {
	var err error = nil
	var demandChanged bool

	demandChanged, err = update(input, ts)
	if err != nil {
		log.Printf("Failed to get new demand. %v", err)
		return err
	}

	if demandChanged {
		// Ask the scheduler to make the changes
		for name, task := range ts {
			// Don't change our own force12 task
			if name != "force12" {
				err = s.StopStartNTasks(name, &task)
				if err != nil {
					log.Printf("Failed to start %s tasks. %v", name, err)
					break
				}
				ts[name] = task
			}
		}
	}

	return err
}

// update checks for changes in demand, returning true if demand changed
// TODO! Make this less tied to the p1 / p2 simple model
func update(input demand.Input, ts map[string]demand.Task) (bool, error) {
	var err error = nil
	var demandchange bool

	var p1 demand.Task = ts[p1TaskName]
	var p2 demand.Task = ts[p2TaskName]

	// Save the old demand
	oldP1Demand := p1.Demand
	oldP2Demand := p2.Demand

	// TODO! In this super-simple RNG model we have to get p1 first so that p2 gets whatever capacity is left over.
	p1.Demand, err = input.GetDemand(p1TaskName)
	if err != nil {
		log.Printf("Failed to get new demand for task %s. %v", p1TaskName, err)
		return false, err
	}
	p2.Demand, err = input.GetDemand(p2TaskName)
	if err != nil {
		log.Printf("Failed to get new demand for task %s. %v", p2TaskName, err)
		return false, err
	}

	//Has the demand changed?
	demandchange = (p1.Demand != oldP1Demand) || (p2.Demand != oldP2Demand)

	// Update tasks map
	ts[p1TaskName] = p1
	ts[p2TaskName] = p2

	// This is where we could decide whether this is a significant enough
	// demand change to do anything

	log.Println(ts)

	return demandchange, err
}
