package main

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Start a watcher on all services
// Call syncConfigWithService when a service is added/modified
func startServiceWatcher(client *kubernetes.Clientset) {
	serviceWatchlist := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "services", v1.NamespaceAll, fields.Everything())
	_, serviceController := cache.NewInformer(serviceWatchlist, &v1.Service{}, time.Second*0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := obj.(*v1.Service)
			fmt.Printf("> Discovered new Service: %s\n", service.Name)
			syncConfigWithService(service, client)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Printf("Service deleted: %s\n", obj.(*v1.Service).Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			service := newObj.(*v1.Service)
			fmt.Printf("> Service changed %s", service.Name)
			syncConfigWithService(service, client)
		},
	})

	stop := make(chan struct{})
	go serviceController.Run(stop)
}
