package main

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSettings(t *testing.T) {
	var s settings
	var testval int = 5000
	var testid string = "hello"
	os.Setenv("F12_DEMAND_CHANGE_INTERVAL_MS", strconv.Itoa(testval))
	os.Setenv("F12_USER_ID", testid)

	s = get_settings()
	if s.demandInterval != time.Duration(testval)*time.Millisecond {
		t.Fatalf("Unexpected demandInterval ")
	}
	if s.userID != testid {
		t.Fatalf("Unexpected userID")
	}

}

func TestTasks(t *testing.T) {
	_ = get_settings()
	// TODO!! This would be async.
	// tasks := get_tasks(s)
	log.Println("Tasks: ", tasks)
	log.Println("Not a test that does much!")
}

func TestInitScheduler(t *testing.T) {
	var err error

	tests := []struct {
		sched string
		pass  bool
	}{
		{sched: "COMPOSE", pass: false},
		{sched: "DOCKER", pass: true},
		{sched: "ECS", pass: false},
		{sched: "KUBERNETES", pass: false},
		{sched: "MESOS", pass: false},
		{sched: "NOMAD", pass: false},
		{sched: "TOY", pass: true},
		{sched: "BLAH", pass: false},
	}

	for _, test := range tests {
		os.Setenv("F12_SCHEDULER", test.sched)
		st := get_settings()
		_, err = get_scheduler(st)
		if err != nil && test.pass {
			t.Fatalf("Should have been able to create %s", test.sched)
		}
		if err == nil && !test.pass {
			t.Fatalf("Should not have been able to create %s", test.sched)
		}
	}
}

func TestInitDemand(t *testing.T) {
	var err error

	tests := []struct {
		input string
		pass  bool
	}{
		{input: "CONSUL", pass: false},
		{input: "RNG", pass: true},
		{input: "BLAH", pass: false},
	}

	for _, test := range tests {
		os.Setenv("F12_DEMAND_MODEL", test.input)
		st := get_settings()
		_, err = get_demand_input(st)
		if err != nil && test.pass {
			t.Fatalf("Should have been able to create %s", test.input)
		}
		if err == nil && !test.pass {
			t.Fatalf("Should not have been able to create %s", test.input)
		}
	}
}
