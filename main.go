package main

import (
	"os"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
)

const kubernetesHost = "http://10.20.40.254:8080/"
const kubernetesNamespace = "plex"
const dockerImage = "registry.marley.xyz/e720/plex-new-transcoder"

type TranscodeJob struct {
	pod        *api.Pod
}

func kubernetesClient() (*client.Client, error) {
	client, err := client.New(
		&client.Config{
			Host: kubernetesHost,
			Version: "v1",
		})

	if err != nil {
		return client, err
	}

	return client, err
}

func (t TranscodeJob) String() string {
	return t.pod.ObjectMeta.Name
}

func (t TranscodeJob) start() error {
	log.Print("Executing job: ", t)
	client, err := kubernetesClient()

	if err != nil {
		return err
	}

	pod, err := client.Pods(t.pod.ObjectMeta.Namespace).Create(t.pod)

	if err != nil {
		return err
	}

	log.Print("Successfully scheduled transcode on pod with name: ", pod.ObjectMeta.Name)

	return nil
}

func generateName() string {
	return "transcode-job"
}

func main() {
	// Get the arguments passed to Plex New Transcoder
	cmd := "/Plex New Transcoder"
	args := os.Args[1:]
	log.Print(fmt.Sprintf("Dispatching job: %s %s", cmd, args))


	job := TranscodeJob{
		pod: 
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
							Args: args,
						},
					},
				},
			},
	}

	err := job.start()
	if err != nil {
		log.Fatal("Job start failed with error: ", err)
	}
}
