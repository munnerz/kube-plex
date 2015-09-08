package executors

import (
	"fmt"
)

type Job struct {
	Command       []string
	Args          []string
}

type Executor interface {
	Start() error
	Stop() error
	WaitForState(ExecutorPhase) error
	String() string
}

type AbstractExecutor struct {
	Job        Job
}

type ExecutorPhase string

const (
	ExecutorPreparing    ExecutorPhase = "Preparing"
	ExecutorRunning      ExecutorPhase = "Running"
	ExecutorSucceeded    ExecutorPhase = "Succeeded"
	ExecutorFailed       ExecutorPhase = "Failed"
	ExecutorUnknown      ExecutorPhase = "Unknown"
)

func (e *AbstractExecutor) Start() error {
	return nil
}

func (e *AbstractExecutor) Stop() error {
	return nil
}

func (e *AbstractExecutor) WaitForState(p ExecutorPhase) error {
	return nil
}

func (e *AbstractExecutor) String() string {
	return fmt.Sprintf("%s %s", e.Job.Command, e.Job.Args)
}