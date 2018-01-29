package executor

import (
	"log"
	"testing"
	"os"
	"k8s.io/client-go/tools/cache"
	"github.com/stretchr/testify/assert"
	"github.com/munnerz/kube-plex/pkg/kube-plex"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex/fake"
)

func TestExecutor(t *testing.T) {
	controller := fake.NewFakeController()

	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			created := new.(*ptjv1.PlexTranscodeJob)
			dir, _ := os.Getwd()
			assert.Equal(t, created.Spec.Cwd, dir, "directory should be set")

			log.Println("status", created.Status.State)

			created.Status.State = ptjv1.PlexTranscodeStateCompleted
			created.Status.Transcoder = "helloworld"
			kubeplex.UpdatePlexTranscodeJob(created, controller.KubeClient)
		},
	})

	Run(controller)

	<-controller.Stop
}
