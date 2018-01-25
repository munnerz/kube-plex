package worker

import (
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex"

	"k8s.io/client-go/tools/cache"
)

const myPodName = "helloworld"

func Run(controller kubeplex.Controller) error {
	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			if updated.Status.Transcoder != myPodName {
				return
			}

			if updated.Status.State != ptjv1.PlexTranscodeStateAssigned {
				return
			}

			state := kubeplex.RunPlexTranscodeJob(updated)
			kubeplex.UpdatePlexTranscodeJobState(updated, state, controller.KubeClient)
		},
	})

	return nil
}
