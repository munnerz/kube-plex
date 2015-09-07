package job

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"

	log "github.com/Sirupsen/logrus"
)

type Job struct {
	Host       string
	Pod        *api.Pod

	client     *client.Client
}

func (t Job) String() string {
	return t.Pod.ObjectMeta.Name
}

func (t Job) Start() error {
	log.Print("Executing job: ", t)

	client, err := client.New(
		&client.Config{
			Host: t.Host,
		})

	if err != nil {
		return err
	}

	t.client = client

	log.Print("Creating pod...")
	pod, err := t.client.Pods(t.Pod.ObjectMeta.Namespace).Create(t.Pod)

	if err != nil {
		return err
	}

	t.Pod = pod

	log.Print("Successfully scheduled job on cluster with name: ", pod.ObjectMeta.Name)

	return nil
}

func (t Job) Stop() error {
	return t.client.Pods(t.Pod.ObjectMeta.Namespace).Delete(t.Pod.ObjectMeta.Name, nil)
}

func (t Job) WaitForState(targetState api.PodPhase) error {
	Loop:
	for {
		pod, err := t.client.Pods(t.Pod.ObjectMeta.Namespace).Get(t.Pod.ObjectMeta.Name)
		if err != nil {
			return err
		}

		Switch:
		switch pod.Status.Phase {
		case targetState:
			break Loop
		default:
			break Switch
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}