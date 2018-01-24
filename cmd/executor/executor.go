package main

import (
	"fmt"
	"log"
	"os"
	"github.com/munnerz/kube-plex/pkg/kube-plex"
	"k8s.io/client-go/tools/cache"

	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"

	clientset "github.com/munnerz/kube-plex/pkg/client/clientset/versioned"
)

func main() {
	cfg, kubeClient, err := kubeplex.KubeClient()
	if err != nil {
		log.Fatal(err)
	}

	kubeplexClient, err := clientset.NewForConfig(cfg)

	args := os.Args
	kubeplex.RewriteArgs(args)

	controller := kubeplex.NewController(kubeClient, kubeplexClient)

	ptj := kubeplex.GeneratePlexTranscodeJob(args)
	new_ptj, err := controller.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex").Create(&ptj)
	if err != nil {
		log.Fatal(err)
	}

	controller.Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			updated := new.(*ptjv1.PlexTranscodeJob)

			if updated.ObjectMeta.Name != new_ptj.ObjectMeta.Name {
				return
			}

			fmt.Println(updated.ObjectMeta.Name, updated.Status.State)
			if updated.Status.State == ptjv1.PlexTranscodeStateCompleted {
				// why doesn't this shut it down?
				controller.Stop <- struct{}{}
			}
		},
	})

	go controller.Run()

	// Set assigned
	_, err = controller.UpdatePlexTranscodeJobState(new_ptj.ObjectMeta.Name, ptjv1.PlexTranscodeStateAssigned)
	if err != nil {
		log.Fatal(err)
	}

	// Set completed, which should make us exit
	_, err = controller.UpdatePlexTranscodeJobState(new_ptj.ObjectMeta.Name, ptjv1.PlexTranscodeStateCompleted)
	if err != nil {
		log.Fatal(err)
	}

	<-controller.Stop
}
