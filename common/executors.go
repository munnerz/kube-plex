package common

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type ExecutorPhase string

const (
	ExecutorPhasePreparing ExecutorPhase = "Preparing"
	ExecutorPhaseRunning   ExecutorPhase = "Running"
	ExecutorPhaseSucceeded ExecutorPhase = "Succeeded"
	ExecutorPhaseFailed    ExecutorPhase = "Failed"
	ExecutorPhaseUnknown   ExecutorPhase = "Unknown"
)

type Executor interface {
	Start() error
	Stop() error
	WaitForState(ExecutorPhase) error
	String() string
}

type Job struct {
	Args []string
}

type ExecutorFactory struct {
	Create func(Config, Job) Executor
}

var executorFactories = make(map[string]ExecutorFactory)

func CreateExecutor(config Config, j Job) (Executor, error) {
	// TODO: Select which executor to use based on config
	for _, e := range executorFactories {
		return e.Create(config, j), nil
	}
	return nil, fmt.Errorf("no configured executor found")
}

func RegisterExecutor(name string, e ExecutorFactory) error {
	if _, ok := executorFactories[name]; ok {
		return fmt.Errorf(fmt.Sprintf("executor already registered: %s", name))
	}

	log.Infof("Registered executor: %s", name)
	executorFactories[name] = e
	return nil
}
