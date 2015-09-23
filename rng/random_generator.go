// Generate a random demand metric for high priority containers
package rng

import (
	"bitbucket.org/force12io/force12-scheduler/demand"
	"log"
	"math/rand"
)

const maximum int = 9 // Demand can vary between 0 and maximum
const delta int = 3   // Current value can only go up or down by a maximum of delta

type RandomDemand struct {
	currentDemand int
}

// check that we implement the demand interface
var _ demand.Input = (*RandomDemand)(nil)

func NewDemandModel() *RandomDemand {
	return &RandomDemand{
		currentDemand: 0,
	}
}

// We ignore the container type when we're generating demand randomly
func (rng *RandomDemand) GetDemand(containerType string) (int, error) {

	// Random value between +/- delta is the same as
	// (random value between 0 and 2*delta) - delta
	// noting that if r = rand.Intn(n) then 0 <= r < n

	r := rand.Intn((2 * delta) + 1)
	demand := rng.currentDemand + r - delta
	if demand > maximum {
		demand = maximum
	}

	if demand < 0 {
		demand = 0
	}

	log.Printf("Random demand %d", demand)
	rng.currentDemand = demand
	return demand, nil
}
