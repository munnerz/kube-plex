package kubernetes

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	stableApi "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/munnerz/plex-elastic-transcoder/common"
	"github.com/munnerz/plex-elastic-transcoder/executors"
)

type KubernetesExecutor struct {
	executors.AbstractExecutor

	pod    *api.Pod
	client *client.Client
}

func (e *KubernetesExecutor) createPod() *api.Pod {
	return &api.Pod{
		TypeMeta: stableApi.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: api.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", e.Config.Kubernetes.PodBasename),
			Namespace:    e.Config.Kubernetes.Namespace,
		},
		Spec: api.PodSpec{
			Volumes: []api.Volume{
				api.Volume{
					Name:         "source-dir",
					VolumeSource: e.Config.Kubernetes.MediaVolumeSource,
				},
				api.Volume{
					Name:         "transcode-dir",
					VolumeSource: e.Config.Kubernetes.TranscodeVolumeSource,
				},
			},
			RestartPolicy: api.RestartPolicyNever,
			Containers: []api.Container{
				api.Container{
					Name:  "plex-new-transcoder",
					Image: e.Config.Kubernetes.Image,
					Args:  e.Job.Args,
					VolumeMounts: []api.VolumeMount{
						api.VolumeMount{
							Name:      "source-dir",
							MountPath: e.Config.Plex.MediaDir,
						},
						api.VolumeMount{
							Name:      "transcode-dir",
							MountPath: e.Config.Plex.TranscodeDir,
						},
					},
					ImagePullPolicy: api.PullAlways,
				},
			},
		},
	}
}

func (e *KubernetesExecutor) Start() error {
	log.Debugf("executing job: %s", e.Job)

	if len(e.Config.Kubernetes.ProxyURL) > 0 {
		e.client = client.NewOrDie(&restclient.Config{
			Host: e.Config.Kubernetes.ProxyURL,
		})
	} else {
		var err error
		e.client, err = client.NewInCluster()
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %s", err.Error())
		}
	}

	pod := e.createPod()

	var err error
	e.pod, err = e.client.Pods(pod.ObjectMeta.Namespace).Create(pod)

	if err != nil {
		return err
	}

	log.Debugf("successfully scheduled job on cluster with name: %s", e.pod.ObjectMeta.Name)

	return nil
}

func (e *KubernetesExecutor) Stop() error {
	return e.client.Pods(e.pod.ObjectMeta.Namespace).Delete(e.pod.ObjectMeta.Name, nil)
}

func podPhaseToExecutorPhase(in api.PodPhase) common.ExecutorPhase {
	switch in {
	case api.PodRunning:
		return common.ExecutorPhaseRunning
	case api.PodPending:
		return common.ExecutorPhasePreparing
	case api.PodSucceeded:
		return common.ExecutorPhaseSucceeded
	case api.PodFailed:
		return common.ExecutorPhaseFailed
	default:
		return common.ExecutorPhaseUnknown
	}
}

func (e *KubernetesExecutor) WaitForState(targetState common.ExecutorPhase) error {
Loop:
	for {
		pod, err := e.client.Pods(e.pod.ObjectMeta.Namespace).Get(e.pod.ObjectMeta.Name)
		if err != nil {
			return err
		}

		log.Debugf("pod state: %s", pod.Status.Phase)

	Switch:
		switch podPhaseToExecutorPhase(pod.Status.Phase) {
		case targetState:
			break Loop
		case common.ExecutorPhaseFailed:
			return fmt.Errorf("pod failed whilst waiting for state '%s': %s", targetState, pod.Status.Reason)
		default:
			break Switch
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func init() {
	common.RegisterExecutor("kubernetes", common.ExecutorFactory{
		Create: func(config common.Config, job common.Job) common.Executor {
			return &KubernetesExecutor{
				AbstractExecutor: executors.AbstractExecutor{
					Config: config,
					Job:    job,
				},
			}
		},
	})
}
