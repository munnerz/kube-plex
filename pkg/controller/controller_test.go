package controller

import (
	"log"
	"testing"
	"k8s.io/client-go/tools/cache"
	"github.com/stretchr/testify/assert"
	"github.com/munnerz/kube-plex/pkg/kube-plex"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex/fake"
)

func TestControllerAssignsJobs(t *testing.T) {
	ptj := kubeplex.GeneratePlexTranscodeJob([]string{"/bin/touch", "/tmp/test"}, []string{})
	controller := fake.NewFakeController(&ptj)

	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			log.Println("status", updated.Status.State)

			if updated.Status.State == ptjv1.PlexTranscodeStateAssigned {
				assert.Equal(t, updated.Status.Transcoder, "helloworld", "invalid transcoder")
				controller.Shutdown()
			}
		},
	})

	Run(controller)

	kubeplex.CreatePlexTranscodeJob(&ptj, controller.KubeClient)

	<-controller.Stop
}
