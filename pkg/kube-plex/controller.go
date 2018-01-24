package kubeplex

import (
	"errors"
	"time"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/runtime"

	clientset "github.com/munnerz/kube-plex/pkg/client/clientset/versioned"
	informers "github.com/munnerz/kube-plex/pkg/client/informers/externalversions"
)

type Controller struct {
	Informer cache.SharedIndexInformer
	KubeplexClient clientset.Interface
	KubeClient kubernetes.Interface
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

func NewController(kubeClient kubernetes.Interface, kubeplexClient clientset.Interface) Controller {
	kubeplexInformerFactory := informers.NewSharedInformerFactory(kubeplexClient, time.Second*30)

	kubeplexInformer := kubeplexInformerFactory.Kubeplex().V1().PlexTranscodeJobs()

	c := Controller{
		KubeplexClient: kubeplexClient,
		KubeClient: kubeClient,
		Informer: kubeplexInformer.Informer(),
		Stop: make(chan struct{}),
	}

	go kubeplexInformerFactory.Start(c.Stop)
	return c
}
