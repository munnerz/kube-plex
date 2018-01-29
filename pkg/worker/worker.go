package worker

import (
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/kube-plex"

	"k8s.io/client-go/tools/cache"
	"log"
)

const myPodName = "helloworld"

func isMyJob(ptj *ptjv1.PlexTranscodeJob) bool {
	return ptj.Status.Transcoder == myPodName
}

func Run(controller kubeplex.Controller) error {
	controller.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			if isMyJob(updated) != true {
				return
			}

			if updated.Status.State != ptjv1.PlexTranscodeStateAssigned {
				return
			}

			log.Println("Running job: ", updated.ObjectMeta.Name)
			updated.Status.State, updated.Status.Error = kubeplex.RunPlexTranscodeJob(updated)
			kubeplex.UpdatePlexTranscodeJob(updated, controller.KubeClient)
			log.Println("Updated job status for ", updated.ObjectMeta.Name, ": ", updated.Status.State)
		},
	})

	return nil
}
