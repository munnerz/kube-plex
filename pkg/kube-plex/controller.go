package kubeplex

import (
	"errors"
	"log"
	"sync"
	"time"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/runtime"

	informers "github.com/munnerz/kube-plex/pkg/client/informers/externalversions"
)

type Controller struct {
	Informer cache.SharedIndexInformer
	InformerFactory informers.SharedInformerFactory
	KubeClient *KubeClient
	Stop chan struct{}
  stopped bool
	stopMutex sync.Mutex
}

func (c *Controller) Shutdown() {
	c.stopMutex.Lock()
	defer c.stopMutex.Unlock()

	if c.stopped == true {
		log.Println("already stopped")
		return
	}

	log.Println("stopping!")
	c.stopped = true
	close(c.Stop)
}

func (c *Controller) Run() error {
	defer runtime.HandleCrash()

	go c.InformerFactory.Start(c.Stop)

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
	kubeplexInformerFactory := informers.NewFilteredSharedInformerFactory(kubeClient.KubeplexClient, time.Second*1, "kube-plex", nil)
	kubeplexInformer := kubeplexInformerFactory.Kubeplex().V1().PlexTranscodeJobs()

	c := Controller{
		KubeClient: kubeClient,
		Informer: kubeplexInformer.Informer(),
		InformerFactory: kubeplexInformerFactory,
		Stop: make(chan struct{}),
	}

	return c
}
