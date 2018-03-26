package main

import (
	"fmt"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// Synchronise proxy config map with service annotation
// If there is a port conflict config map config is kept
// If service proxy port changes, old line is deleted adn new one is added
func syncConfigWithService(service *v1.Service, client *kubernetes.Clientset) {
	configMapClient := client.CoreV1().ConfigMaps("ingress-nginx")

	proxyConfigs := ParseConfigMap("tcp", client)
	ourConfigs := GetFromService(service)

	if (len(ourConfigs) == 0) {
		return
	}

	toCreateConfigs := []ServiceConfig{}
	toDeleteConfigs := []ServiceConfig{}

	fmt.Println(" Synchronizing configs...")
	for _, configToCheck := range proxyConfigs {
		for _,config := range ourConfigs {
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

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		updatedConfigMap, err := configMapClient.Get("tcp-services", metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		if updatedConfigMap.Data == nil {
			updatedConfigMap.Data = map[string]string{}
		}

		for _, toDeleteConfig := range toDeleteConfigs {
			delete(updatedConfigMap.Data, toDeleteConfig.ProxyPort)
		}

		for _, toCreateConfig := range toCreateConfigs {
			updatedConfigMap.Data[toCreateConfig.ProxyPort] = fmt.Sprintf("%s:%s", toCreateConfig.FQSN, toCreateConfig.ServicePort)
		}

		_, updateErr := configMapClient.Update(updatedConfigMap)

		return updateErr
	})
	if retryErr != nil {
		panic(retryErr.Error())
	}

	fmt.Printf(" Config map updated for %s/%s\n", service.Namespace, service.Name)
}
