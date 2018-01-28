package fake

import (
	"github.com/munnerz/kube-plex/pkg/kube-plex"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeplex "github.com/munnerz/kube-plex/pkg/client/clientset/versioned/fake"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
)

func NewFakeController(objects ...runtime.Object) kubeplex.Controller {
	kc := kubeplex.KubeClient{
		Clientset: fakekubernetes.NewSimpleClientset(),
		KubeplexClient: fakekubeplex.NewSimpleClientset(objects...),
	}

	controller := kubeplex.NewController(&kc)
	go controller.Run()
	return controller
}
