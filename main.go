package main

import (
	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Main function
// Initialize connection to Kubernetes API and start watching for services
func main() {
	var kubeConfig *string
	kubeConfig = flag.String("kubeConfig", filepath.Join(".", "config"), "(optional) absolute path to the kube config file")

	flag.Parse()

	// use the current context in kubeConfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	// create the client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	startServiceWatcher(client)

	for {
		time.Sleep(10 * time.Second)
	}
}
