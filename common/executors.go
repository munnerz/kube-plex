package common

import (
	"errors"
	"fmt"

	"github.com/munnerz/plex-elastic-transcoder/executors"

	log "github.com/Sirupsen/logrus"
)

type ExecutorFactory struct {
	Create        func(executors.Job) executors.Executor
}

var executorFactories map[string]ExecutorFactory

func CreateExecutor(j executors.Job) executors.Executor {
	// TODO: Some sort of executor selection algorithm
	for _, e := range executorFactories {
		// TODO: Here, run the create command defined in the factory
		return e.Create(j)
		// Hacky way to dispatch to the first executor
	}
	panic("No executors registered!")
}

func RegisterExecutor(name string, e ExecutorFactory) error {
	if executorFactories == nil {
		executorFactories = make(map[string]ExecutorFactory)
	}

	if _, ok := executorFactories[name]; ok {
		return errors.New(fmt.Sprintf("Executor already registered: %s", name))
	}

	log.Print("Registered executor: ", name)
	executorFactories[name] = e
	return nil
}