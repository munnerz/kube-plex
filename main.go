package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/pkg/api"

	"github.com/munnerz/plex-elastic-transcoder/job"
)

const kubernetesHost = "http://10.20.40.254:8080/"
const kubernetesNamespace = "default"
const dockerImage = "registry.marley.xyz/e720/plex-new-transcoder"
const podBasename = "plex-transcoder"

func main() {
	// Get the arguments passed to Plex New Transcoder
	args := os.Args[1:]
	log.Print("Dispatching job with args: ", args)

	job := job.Job{
		Host: kubernetesHost,
		Pod: 
			&api.Pod{
				TypeMeta: api.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: api.ObjectMeta{
					GenerateName: podBasename,
					Namespace: kubernetesNamespace,
				},
				Spec: api.PodSpec{
					RestartPolicy: api.RestartPolicyNever,
					Containers: []api.Container{
						api.Container{
							Name: podBasename,
							Image: dockerImage,
							Args: args,
						},
					},
				},
			},
	}

	err := job.Start()
	if err != nil {
		log.Fatal("Job start failed with error: ", err)
	}

	log.Print("Waiting for build pod to enter Running state...")

	err = job.WaitForState(api.PodRunning)
	if err != nil {
		log.Fatal("Error waiting for pod to enter running state: ", err)
	}

	log.Print("Job has started running...")
	log.Print("Waiting for job to complete...")

	err = job.WaitForState(api.PodSucceeded)
	if err != nil {
		log.Fatal("Error waiting for job to complete: ", err)
	}

	log.Print("Job completed. Destroying pod...")

	err = job.Stop()
	if err != nil {
		log.Fatal("Error stopping job: ", err)
	}

	log.Print("Pod destroyed. Exiting.")
}
