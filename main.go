package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/plex-elastic-transcoder/common"
	"github.com/munnerz/plex-elastic-transcoder/executors"

	_ "github.com/munnerz/plex-elastic-transcoder/executors/kubernetes"
)

var executor executors.Executor

func signals() {
	// Signal handling
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		log.Print("Shutting down...")

		if executor != nil {
			log.Print("Terminating running job...")
			err := executor.Stop()
			if err != nil {
				log.Fatal("Error terminating job: ", err)
				os.Exit(1)
			}

			log.Print("Successfully terminated running job.")
		}

		os.Exit(0)
	}()
}
func main() {
	// Setup signals
	signals()

	// Get the arguments passed to Plex New Transcoder
	args := os.Args[1:]
	log.Print("Dispatching job with args: ", args)

	job := executors.Job{
		Command: []string{"/Plex New Transcoder"},
		Args: args,
	}

	executor = common.CreateExecutor(job)

	log.Print("Created executor: ", executor)

	err := executor.Start()
	if err != nil {
		log.Fatal("Job start failed with error: ", err)
	}

	log.Print("Waiting for build pod to enter Running state...")

	err = executor.WaitForState(executors.ExecutorRunning)
	if err != nil {
		log.Fatal("Error waiting for pod to enter running state: ", err)
	}

	log.Print("Job has started running...")
	log.Print("Waiting for job to complete...")

	err = executor.WaitForState(executors.ExecutorSucceeded)
	if err != nil {
		log.Fatal("Error waiting for job to complete: ", err)
	}

	log.Print("Job completed. Destroying pod...")

	err = executor.Stop()
	if err != nil {
		log.Fatal("Error stopping job: ", err)
	}

	log.Print("Pod destroyed. Exiting.")
}
