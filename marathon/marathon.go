package marathon

import (
	"os"

	"bitbucket.org/force12io/force12-scheduler/consul"
	"bitbucket.org/force12io/force12-scheduler/scheduler"
)

type MarathonScheduler struct {
	baseMarathonUrl  string
	demandFromConsul *consul.DemandFromConsul
}

// compile-time assert that we implement the right interface
var _ scheduler.Scheduler = (*MarathonScheduler)(nil)

func NewScheduler() *MarathonScheduler {
	return &MarathonScheduler{
		baseMarathonUrl:  getBaseMarathonUrl(),
		demandFromConsul: consul.NewDemandModel(),
	}
}

func getBaseMarathonUrl() string {
	baseUrl := os.Getenv("MARATHON_ADDRESS")
	if baseUrl == "" {
		baseUrl = "http://marathon.force12.io:8080"
	}
	return baseUrl + "/v2/apps"
}

func (m *MarathonScheduler) GetContainerCount(key string) (int, error) {
	// TODO! This is just reflecting back what we had demanded. We need to separate out the demand input from
	// finding out what's actually running
	return m.demandFromConsul.GetDemand(key)
}
