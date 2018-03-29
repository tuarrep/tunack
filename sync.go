package main

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Synchronise proxy config map with service annotation
// If there is a port conflict config map config is kept
// If service proxy port changes, old line is deleted adn new one is added
func syncConfigWithService(service *v1.Service, client *kubernetes.Clientset) {
	proxyConfigs := ParseConfigMap("tcp", client)
	ourConfigs := GetFromService(service)

	if len(ourConfigs) == 0 {
		return
	}

	toCreateConfigs := []ServiceConfig{}
	toDeleteConfigs := []ServiceConfig{}

	fmt.Println(" Synchronizing configs...")
	for _, configToCheck := range proxyConfigs {
		for _, config := range ourConfigs {
			if configToCheck.ProxyPort == config.ProxyPort {
				if configToCheck.RuleTag() == config.RuleTag() {
					fmt.Printf(" * %s : Already in sync\n", config.RuleTag())
				} else {
					fmt.Printf(" ! Service %s already bound on proxy port %s. Ignoring.\n", configToCheck.FQSN, configToCheck.ProxyPort)
				}
			} else {
				if configToCheck.FQSN == config.FQSN && configToCheck.ServicePort == config.ServicePort {
					fmt.Printf(" * %s Was bound to another port. Updating it.\n", config.RuleTag())
					toDeleteConfigs = append(toDeleteConfigs, configToCheck)
				}

				toCreateConfigs = append(toCreateConfigs, config)
			}
		}
	}

	fmt.Println(" Updating config map")

	updateErr := UpdateConfigMap(toCreateConfigs, toDeleteConfigs, client)

	if updateErr != nil {
		panic(updateErr.Error())
	}

	fmt.Printf(" Config map updated for %s/%s\n", service.Namespace, service.Name)
}
