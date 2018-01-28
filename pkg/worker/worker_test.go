package worker

import (
	"log"
	"os"
	"testing"
	"time"
	"k8s.io/client-go/tools/cache"
	"github.com/stretchr/testify/assert"
	"github.com/munnerz/kube-plex/pkg/kube-plex"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex/fake"
)

func TestWorkerJobSuccess(t *testing.T) {
	ptj := kubeplex.GeneratePlexTranscodeJob([]string{"/bin/touch", "/tmp/test"}, []string{})
	controller := fake.NewFakeController(&ptj)

	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			log.Println("status", updated.Status.State)

			if updated.Status.State != ptjv1.PlexTranscodeStateCompleted {
				return
			}

			_, err := os.Stat("/tmp/test")
			assert.Equal(t, err, nil, "/tmp/test should exist!")
			controller.Shutdown()
		},
	})

	Run(controller)

	new_ptj, _ := kubeplex.CreatePlexTranscodeJob(&ptj, controller.KubeClient)

	new_ptj.Status.State = ptjv1.PlexTranscodeStateAssigned
	new_ptj.Status.Transcoder = "helloworld"
	kubeplex.UpdatePlexTranscodeJob(new_ptj, controller.KubeClient)

	<-controller.Stop
}

func TestWorkerJobFailure(t *testing.T) {
	ptj := kubeplex.GeneratePlexTranscodeJob([]string{"/bin/cat", "/does/not/exist"}, []string{})
	controller := fake.NewFakeController(&ptj)

	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			log.Println("(failure test) status", updated.Status.State)
			if updated.Status.State == ptjv1.PlexTranscodeStateFailed {
				controller.Shutdown()
			} else if updated.Status.State == ptjv1.PlexTranscodeStateAssigned {
				return
			} else {
				t.Error("State is not ASSIGNED or FAILED: ", updated.Status.State)
			}
		},
	})

	Run(controller)

	new_ptj, _ := kubeplex.CreatePlexTranscodeJob(&ptj, controller.KubeClient)

	new_ptj.Status.State = ptjv1.PlexTranscodeStateAssigned
	new_ptj.Status.Transcoder = "helloworld"
	kubeplex.UpdatePlexTranscodeJob(new_ptj, controller.KubeClient)

	<-controller.Stop
}

func TestWorkerJobAssignedToOtherWorker(t *testing.T) {
	ptj := kubeplex.GeneratePlexTranscodeJob([]string{"/bin/touch", "/tmp/mytest"}, []string{})
	controller := fake.NewFakeController(&ptj)

	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			log.Println("(worker name test) status", updated.Status.State)
			if updated.Status.State == ptjv1.PlexTranscodeStateAssigned {
				time.Sleep(time.Second*1)
				controller.Shutdown()
			} else if updated.Status.State == ptjv1.PlexTranscodeStateCreated {
				return
			} else {
				t.Error("Worker should not have processed the job.")
			}
		},
	})

	Run(controller)

	new_ptj, _ := kubeplex.CreatePlexTranscodeJob(&ptj, controller.KubeClient)

	new_ptj.Status.State = ptjv1.PlexTranscodeStateAssigned
	new_ptj.Status.Transcoder = "hellomoon"
	kubeplex.UpdatePlexTranscodeJob(new_ptj, controller.KubeClient)

	<-controller.Stop
}
