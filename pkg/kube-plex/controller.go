package kubeplex

import (
	"errors"
	"time"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/runtime"

	informers "github.com/munnerz/kube-plex/pkg/client/informers/externalversions"
)

type Controller struct {
	Informer cache.SharedIndexInformer
	KubeClient *KubeClient
	Stop chan struct{}
}

func (c *Controller) Run() error {
	defer runtime.HandleCrash()

	go c.Informer.Run(c.Stop)
	
	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(c.Stop, c.Informer.HasSynced) {
		return errors.New("failed to sync caches")
	}

	<-c.Stop
	return nil
}

func (c *Controller) AddEventHandler(handlers cache.ResourceEventHandlerFuncs) {
	c.Informer.AddEventHandler(handlers)
}

func NewController(kubeClient *KubeClient) Controller {
	kubeplexInformerFactory := informers.NewSharedInformerFactory(kubeClient.KubeplexClient, time.Second*30)
	kubeplexInformer := kubeplexInformerFactory.Kubeplex().V1().PlexTranscodeJobs()

	c := Controller{
		KubeClient: kubeClient,
		Informer: kubeplexInformer.Informer(),
		Stop: make(chan struct{}),
	}

	go kubeplexInformerFactory.Start(c.Stop)
	return c
}
