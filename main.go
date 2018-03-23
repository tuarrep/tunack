package main

import (
	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Main function
// Initialize connection to Kubernetes API and start watching for services
func main() {
	var kubeConfig *string
	var inCluster *bool
	kubeConfig = flag.String("kubeConfig", filepath.Join(".", "config"), "(optional) absolute path to the kube config file")
	inCluster = flag.Bool( "inCluster", false, "Use InCluster config")

	flag.Parse()

	var config *rest.Config
	var err error

	// use the current context in kubeConfig
	if !*inCluster {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}
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
