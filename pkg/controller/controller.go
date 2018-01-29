package controller

import (
	"log"

	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex"

	"k8s.io/client-go/tools/cache"
)

var IdlePods = []string{"helloworld"}

func Run(controller kubeplex.Controller) error {
	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			if updated.Status.State == ptjv1.PlexTranscodeStateFailed {
				log.Println("Job failed: " + updated.Status.Error)
				return
			}

			if updated.Status.State != ptjv1.PlexTranscodeStateCreated {
				return
			}

			log.Println("Assigned job to worker.")
			updated.Status.State = ptjv1.PlexTranscodeStateAssigned
			updated.Status.Transcoder = IdlePods[0]
			kubeplex.UpdatePlexTranscodeJob(updated, controller.KubeClient)
		},
	})

	return nil
}
