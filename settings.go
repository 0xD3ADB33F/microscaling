package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/force12io/force12/api"
	"github.com/force12io/force12/demand"
	"github.com/force12io/force12/demandapi"
	"github.com/force12io/force12/docker"
	"github.com/force12io/force12/rng"
	"github.com/force12io/force12/scheduler"
	"github.com/force12io/force12/toy_scheduler"
)

type settings struct {
	demandModelType string
	schedulerType   string
	sendMetrics     bool
	userID          string
	demandInterval  time.Duration
	demandDelta     int
	maxContainers   int
	pullImages      bool
}

func get_settings() settings {
	var st settings
	st.demandModelType = getEnvOrDefault("F12_DEMAND_MODEL", "API")
	st.schedulerType = getEnvOrDefault("F12_SCHEDULER", "DOCKER")
	st.userID = getEnvOrDefault("F12_USER_ID", "5k5gk")
	st.sendMetrics = (getEnvOrDefault("F12_SEND_METRICS_TO_API", "true") == "true")
	st.demandDelta, _ = strconv.Atoi(getEnvOrDefault("F12_DEMAND_DELTA", "3"))
	st.maxContainers, _ = strconv.Atoi(getEnvOrDefault("F12_MAXIMUM_CONTAINERS", "9"))
	demandIntervalMs, _ := strconv.Atoi(getEnvOrDefault("F12_DEMAND_CHANGE_INTERVAL_MS", "3000"))
	st.demandInterval = time.Duration(demandIntervalMs) * time.Millisecond
	st.pullImages = (getEnvOrDefault("F12_PULL_IMAGES", "true") == "true")
	return st
}

func get_demand_input(st settings) (demand.Input, error) {
	var di demand.Input

	switch st.demandModelType {
	case "CONSUL":
		return nil, fmt.Errorf("Demand metric from Consul not yet supported")
	case "RNG":
		log.Println("Local random demand generation")
		di = rng.NewDemandModel(st.demandDelta, st.maxContainers)
		log.Printf("Vary tasks with delta %d up to max %d containers", st.demandDelta, st.maxContainers)
	case "API":
		log.Println("Demand from the API")
		di = demandapi.NewDemandModel(st.userID)
	default:
		return nil, fmt.Errorf("Bad value for F12_DEMAND_MODEL: %s", st.demandModelType)
	}

	if di == nil {
		return nil, fmt.Errorf("No demand input")
	}

	return di, nil
}

func get_scheduler(st settings) (scheduler.Scheduler, error) {
	var s scheduler.Scheduler

	switch st.schedulerType {
	case "DOCKER":
		log.Println("Scheduling with Docker remote API")
		s = docker.NewScheduler(st.pullImages)
	case "ECS":
		return nil, fmt.Errorf("Scheduling with ECS not yet supported. Tweet with hashtag #F12ECS if you'd like us to add this next!")
	case "KUBERNETES":
		return nil, fmt.Errorf("Scheduling with Kubernetes not yet supported. Tweet with hashtag #F12Kubernetes if you'd like us to add this next!")
	case "MESOS":
		return nil, fmt.Errorf("Scheduling with Mesos / Marathon not yet supported. Tweet with hashtag #F12Mesos if you'd like us to add this next!")
	case "NOMAD":
		return nil, fmt.Errorf("Scheduling with Nomad not yet supported. Tweet with hashtag #F12Nomad if you'd like us to add this next!")
	case "TOY":
		log.Println("Scheduling with toy scheduler")
		s = toy_scheduler.NewScheduler()
	default:
		return nil, fmt.Errorf("Bad value for F12_SCHEDULER: %s", st.schedulerType)
	}

	if s == nil {
		return nil, fmt.Errorf("No scheduler")
	}

	return s, nil
}

func get_tasks(st settings) map[string]demand.Task {
	var t map[string]demand.Task

	// Get the tasks that have been configured by this user
	t, err := api.GetApps(st.userID)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
	}

	log.Println(t)
	return t
}

func getEnvOrDefault(name string, defaultValue string) string {
	v := os.Getenv(name)
	if v == "" {
		v = defaultValue
	}

	return v
}
