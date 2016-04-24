package executors

import (
	"fmt"
	
	"github.com/munnerz/plex-elastic-transcoder/common"
)

type AbstractExecutor struct {
	Config 	   common.Config
	Job        common.Job
}

func (e *AbstractExecutor) Start() error {
	return nil
}

func (e *AbstractExecutor) Stop() error {
	return nil
}

func (e *AbstractExecutor) WaitForState(p common.ExecutorPhase) error {
	return nil
}

func (e *AbstractExecutor) String() string {
	return fmt.Sprintf("%s", e.Job.Args)
}