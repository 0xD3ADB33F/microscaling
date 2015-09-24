/*
Force12.io is a package that monitors demand for resource in a system and then scales and repurposes
containers, based on agreed "quality of service" contracts, to best handle that demand within the constraints of your existing VM or physical infrastructure (for v1).

Force12 is defined to optimize the use of existing physical and VM resources instantly. VMs cannot be scaled in real time (it takes several minutes) and new physical machines take even longer. However, containers can be started or stopped at sub second speeds, allowing your infrastructure to adapt itself in real time to meet system demands.

Force12 is aimed at effectively using the resources you have right now - your existing VMs or physical servers - by using them as optimally as possible.

The Force12 approach is analogous to the way that a router dynamically optimises the use of a physical network. A router is limited by the capacity of the lines physically connected to it. Adding additional capacity is a physical process and takes time. Routers therefore make decisions in real time about which packets will be prioritized on a particular line based on the packet's priority (defined by a "quality of service" contract).

For example, at times of high bandwidth usage a router might prioritize VOIP traffic over web browsing in real time.

Containers allow Force12 to make similar "instant" judgements on service prioritisation within your existing infrastructure. Routers make very simplistic
judgments because they have limited time and cpu and they act at a per packet level. Force12 has the capability of making far more
sophisticated judgements, although even fairly simple ones will still provide a significant new service.

This prototype is a bare bones implementation of Force12.io that recognises only 1 demand types
randomised demand for a priority 1 service (the current value stored in Consul's key value store) aka "client". WHen this minimum demand has been met then a P2 service will
be run (aka "server).

These demand type examples have been chosen purely for simplicity of demonstration. In the future more demand types
will be offered

V1 - Force12.io reacts to increased demand by starting/stopping containers on the slaves already in play.

Note - this version of Force12 starts and stops containers on a Mesos cluser using Marathon as the scheduler
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"bitbucket.org/force12io/force12-scheduler/marathon"
	"bitbucket.org/force12io/force12-scheduler/scheduler"
)

type sendStatePayload struct {
	CreatedAt          int64 `json:"createdAt"`
	Priority1Requested int   `json:"priority1Requested"`
	Priority1Running   int   `json:"priority1Running"`
	Priority2Running   int   `json:"priority2Running"`
}

//CONSTANTS
const const_sleep = 100     //milliseconds
const const_stopsleep = 250 //milliseconds pause between stopping and restarting containers
const const_clientdemandstart int = 5
const const_serverdemandstart int = 4
const const_maxcontainers int = 9

//EXPORTED STRUCTS
type Demand struct {
	sched scheduler.Scheduler

	//  mu  sync.Mutex   // plan ahead for concurrency on this potentially shared resource, this is go after all
	clientdemand     int // indicates number of clients demanded (from dynamo table in prototype)
	serverdemand     int // indicates server demand (deduced in prototype)
	clientsrequested int // indicates how many clients we've tried to kick off.
	serversrequested int // indicates how many servers we've tried to kick off.
}

////////////////////////////////////////////////////////////////////////////////////////
// set
// Setter, returns what was there (client, server)
// if provided value is -1 don't update, demand will always be between 0 and const_maxcontainers
func (d *Demand) set(client, server int) (int, int) {
	//d.mu.Lock()
	serverold := d.serverdemand
	clientold := d.clientdemand
	if server != -1 {
		d.serverdemand = server
	}
	if client != -1 {
		d.clientdemand = client
	}
	//d.mu.Unlock()
	return clientold, serverold
}

////////////////////////////////////////////////////////////////////////////////////////
// get
// Getter, returns client, server AEC - Combine this with the set to reduce code
func (d *Demand) get() (int, int) {
	return d.clientdemand, d.serverdemand
}

////////////////////////////////////////////////////////////////////////////////////////
// handle
//
// Called in response to a change in demand to process it.
//
// output err - true if success (AEC currently assumes success).
//
// Note that handle will make any judgment on what to do with a demand
// change, including potentially nothing.
//
func (d *Demand) handle() bool {
	log.Println("Handle demand change.")

	// AEC NOTE THIS FUNCTION NEEDS TO BE HEAVILY REWRITTEN TO HANDLE ECS
	// WHEN WE PORT THAT OVER TO THE SAME STRUCTURE.
	// THe reason is that all the code we wrote to handle stopping before
	// starting etc.. is handled directly by Marathon so that code
	// from the old scheduler needs to go behind the scheduler interface
	err := d.sched.StopStartNTasks(os.Getenv("CLIENT_TASK"), os.Getenv("CLIENT_FAMILY"), d.clientdemand, d.clientsrequested)
	if err != nil {
		log.Printf("Failed to start client tasks. %v", err)
	}
	d.sched.StopStartNTasks(os.Getenv("SERVER_TASK"), os.Getenv("SERVER_FAMILY"), d.serverdemand, d.serversrequested)
	if err != nil {
		log.Printf("Failed to start server tasks. %v", err)
	}

	return false
}

////////////////////////////////////////////////////////////////////////////////////////
// update
//
// Called periodically to check for changes in demand and update accordingly.
//
// output changed? - true if demand changed
//
// Note that this routine will return any change. It makes no judgement on whether that change is
// significant. handle() will determine that.
//
func (d *Demand) update() bool {
	//log.Println("demand update check.")
	var demandchange bool = false

	container_count, err := d.sched.GetContainerCount("priority1-demand")
	if err != nil {
		log.Printf("Failed to get container count. %v", err)
		return false
	}
	//log.Printf("container count %v\n", container_count)

	//Update our saved client demand
	oldcli, _ := d.set(container_count, const_maxcontainers-container_count)

	//Has the demand changed?
	demandchange = (container_count != oldcli)

	if demandchange {
		log.Println("demandchange from, to ", oldcli, container_count)
	}

	return demandchange
}

////////////////////////////////////////////////////////////////////////////////////////
// sendStateToAPI
//
// Called periodically to check for current state of cluster (or single node) and send that
// state to the f12 API
func sendStateToAPI(currentdemand *Demand) error {
	count1, count2, err := currentdemand.sched.CountAllTasks()
	if err != nil {
		return fmt.Errorf("Failed to get state err %v", err)
	}

	// Submit a PUT request to the API
	// Note the magic hardcoded string is the user ID, we need to pass this in in some way. ENV VAR?
	url := getBaseF12APIUrl() + "/metrics/" + "5k5gk"
	log.Printf("API PUT: %s", url)

	payload := sendStatePayload{
		CreatedAt:          time.Now().Unix(),
		Priority1Requested: currentdemand.p1demand,
		Priority1Running:   count1,
		Priority2Running:   count2,
	}

	w := &bytes.Buffer{}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&payload)
	if err != nil {
		return fmt.Errorf("Failed to encode API json. %v", err)
	}

	req, err := http.NewRequest("PUT", url, w)

	if err != nil {
		return fmt.Errorf("Failed to build API PUT request err %v", err)
	}
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		// handle error
		return fmt.Errorf("API state err %v", err)
	}

	if resp.StatusCode > 204 {
		return fmt.Errorf("error response from API. %s", resp.Status)
	}
	return err
}

func getBaseF12APIUrl() string {
	baseUrl := os.Getenv("API_ADDRESS")
	if baseUrl == "" {
		baseUrl = "https://force12-windtunnel.herokuapp.com"
	}
	return baseUrl
}

////////////////////////////////////////////////////////////////////////////////////////
// MAIN
//
func main() {
	////////////////////////////////////////////////////////////////////////////////////////
	// For the simple prototype, Force12.io sits in a loop checking for demand changes every X milliseconds
	// In phase 2 we'll add a reactive mode where appropriate.
	//
	// Note - we don't route messages from demandcheckers to demandhandlers using channels because we want new values
	// to override old values. Queued history is of no importance here.
	//
	// Also for simplicity this first release is concurrency free (single threaded)
	currentdemand := Demand{
		sched: marathon.NewScheduler(),
	}
	currentdemand.set(const_clientdemandstart, const_serverdemandstart)
	//var errflag bool = false
	var demandchangeflag bool
	//uncomment code below to output logs to file, but there's nothing clever in here to limit file size
	//f, err := os.OpenFile("testlogfile", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	//if err != nil {
	//   panic(err)
	//}
	//defer f.Close()

	//log.SetOutput(f)
	log.Println("This is a test log entry")

	// Initialise container types
	err := currentdemand.sched.InitScheduler(os.Getenv("CLIENT_TASK"))
	if err != nil {
		log.Printf("Failed to start client task. %v", err)
		return
	}
	err = currentdemand.sched.InitScheduler(os.Getenv("SERVER_TASK"))
	if err != nil {
		log.Printf("Failed to start server task. %v", err)
		return
	}

	// Find out how many containers we currently have running and get their details
	// Note have decided to do this periodically as a reset as we are getting mysteriously out of whack on ECS
	currentdemand.clientsrequested, currentdemand.serversrequested, err = currentdemand.sched.CountAllTasks()
	if err != nil {
		log.Printf("Failed to count tasks. %v", err)
	}

	//Now we can talk to the DB to check our client demand
	demandchangeflag = currentdemand.update()
	demandchangeflag = true
	var sleepcount float64 = 0

	for {
		switch {
		case demandchangeflag:
			demandchangeflag = false
			//make any changes dictated by this new demand level
			currentdemand.clientsrequested, currentdemand.serversrequested, err = currentdemand.sched.CountAllTasks()
			if err != nil {
				log.Printf("Failed to count tasks. %v", err)
			}
			//To trace out turn _ = errFlag
			_ = currentdemand.handle()
			//log.Println("demand change. result:", errflag)

		default:
			//log.Println("    .")
			var sleep time.Duration
			sleep = const_sleep * time.Millisecond
			time.Sleep(sleep)
			//Update currentdemand with latest client and server demand, if changed, set flag
			demandchangeflag = currentdemand.update()

			sleepcount++

			//Periodically send state to the API if required
			if os.Getenv("SENDSTATETO_API") == "true" {
				_, frac := math.Modf(math.Mod(sleepcount, 5))
				if frac == 0 {
					err = sendStateToAPI(&currentdemand)
					if err != nil {
						log.Printf("Failed to send state. %v", err)
					}
				}
			}
		}
	}
}
