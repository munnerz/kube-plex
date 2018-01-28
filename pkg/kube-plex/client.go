package kubeplex

import (
	clientset "github.com/munnerz/kube-plex/pkg/client/clientset/versioned"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"os"
)

type KubeClient struct {
	Cfg *rest.Config
	Clientset kubernetes.Interface
	KubeplexClient clientset.Interface
}

func NewKubeClient() (kc *KubeClient, err error) {
	kc = new(KubeClient)

	kc.Cfg, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return
	}

	kc.Clientset, err = kubernetes.NewForConfig(kc.Cfg)
	if err != nil {
		return
	}

	kc.KubeplexClient, err = clientset.NewForConfig(kc.Cfg)
	return
}
