package main

import (
	"log"
	"os"

	"github.com/munnerz/kube-plex/pkg/kube-plex"
	"github.com/munnerz/kube-plex/pkg/executor"
	"github.com/munnerz/kube-plex/pkg/worker"
)

func main() {
	log.Println("Initializing kubernetes API client.")
	kubeClient, err := kubeplex.NewKubeClient()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating controller.")
	controller := kubeplex.NewController(kubeClient)

	switch os.Getenv("KUBEPLEX_ENV") {
	case "executor":
		log.Println("Running executor.")
		err = executor.Run(controller)
	case "worker":
		log.Println("Running worker.")
		err = worker.Run(controller)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Running controller.")
	go controller.Run()


	log.Println("Waiting for stop signal.")
	<-controller.Stop
}
