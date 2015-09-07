package job

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"

	log "github.com/Sirupsen/logrus"
)

type Job struct {
	Host       string
	Pod        *api.Pod
}

func (t Job) kubernetesClient() (*client.Client, error) {
	client, err := client.New(
		&client.Config{
			Host: t.Host,
			Version: "v1",
		})

	if err != nil {
		return client, err
	}

	return client, err
}

func (t Job) String() string {
	return t.Pod.ObjectMeta.Name
}

func (t Job) Start() error {
	log.Print("Executing job: ", t)
	client, err := t.kubernetesClient()

	if err != nil {
		return err
	}

	log.Print("Creating pod...")
	pod, err := client.Pods(t.Pod.ObjectMeta.Namespace).Create(t.Pod)

	if err != nil {
		return err
	}

	log.Print("Successfully scheduled job on cluster with name: ", pod.ObjectMeta.Name)

	return nil
}