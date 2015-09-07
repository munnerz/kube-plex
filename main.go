package main

import (
	"os"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/pkg/api"

	"git.marley.xyz/e720/plex-elastic-transcoder/job"
)

const kubernetesHost = "http://10.20.40.254:8080/"
const kubernetesNamespace = "plex"
const dockerImage = "timhaak/plex"

func generateName() string {
	return "transcode-job"
}

func main() {
	// Get the arguments passed to Plex New Transcoder
	args := os.Args[1:]
	log.Print(fmt.Sprintf("Dispatching job: %s", args))


	job := job.Job{
		Host: kubernetesHost,
		Pod: 
			&api.Pod{
				TypeMeta: api.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: api.ObjectMeta{
					GenerateName: generateName(),
					Namespace: kubernetesNamespace,
				},
				Spec: api.PodSpec{
					RestartPolicy: api.RestartPolicyNever,
					Containers: []api.Container{
						api.Container{
							Name: generateName(),
							Image: dockerImage,
							Command: []string{"/usr/lib/plexmediaserver/Resources/Plex New Transcoder"},
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
}
