package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	
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
	configMap, err := configMapClient.Get("tcp-services", metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	proxyConfigs := []ServiceConfig{}
	ourConfigs := []ServiceConfig{}
	toDeleteConfigs := []ServiceConfig{}

	for key, value := range configMap.Data {
		config := ServiceConfig{
			strings.Split(value, ":")[0],
			"tcp",
			strings.Split(value, ":")[1],
			key,
		}
		proxyConfigs = append(proxyConfigs, config)
	}

	for name, annotation := range service.Annotations {
		if (!strings.Contains(name, "tunack.dahus.io/")) {
			continue
		}
		fmt.Printf(" Found annotation %s: %s\n", name, annotation)
		annotationRegex := regexp.MustCompile(`tunack\.dahus\.io/(tcp|udp)-service-([0-9]{1,5})`)

		matches := annotationRegex.FindAllStringSubmatch(name, -1)[0]
		protocol := matches[1]
		proxyPort := matches[2]
		servicePort := annotation

		servicePortExists := false

		for _, port := range service.Spec.Ports {
			if (string(port.Protocol) == strings.ToUpper(protocol) && strconv.FormatInt(int64(port.Port), 10) == servicePort) {
				servicePortExists = true
				break
			}
		}

		if (!servicePortExists) {
			fmt.Println(" ! Serevice port found in annotion but not found in service spec! Ignoring")
			continue
		}

		config := ServiceConfig{
			fmt.Sprintf("%s/%s", service.Namespace, service.Name),
			protocol,
			servicePort,
			proxyPort,
		}

		fmt.Printf(" * Service FQN: %s\n * Protocol: %s\n * Service port: %s\n * Proxy port: %s\n * Rule tag: %s\n", config.FQSN, config.Protocol, config.ServicePort, config.ProxyPort, config.RuleTag())

		appendConfig := true

		for _, configToCheck := range proxyConfigs {
			if configToCheck.ProxyPort == config.ProxyPort {
				appendConfig = false

				if configToCheck.RuleTag() == config.RuleTag() {
					fmt.Println(" * Already in sync")
				} else {
					fmt.Printf(" ! Service %s already bound on proxy port %s. Ignoring.\n", configToCheck.FQSN, configToCheck.ProxyPort)
				}

				break
			} else {
				if configToCheck.FQSN == config.FQSN && configToCheck.ServicePort == config.ServicePort {
					fmt.Println(" * This service was bound to another port. Updating it.")
					toDeleteConfigs = append(toDeleteConfigs, configToCheck)
				}
			}
		}

		if appendConfig {
			fmt.Println(" * Need to be added")
			ourConfigs = append(ourConfigs, config)
		}
	}

	if (len(ourConfigs) == 0) {
		return
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

		for _, toCreateConfig := range ourConfigs {
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
