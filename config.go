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

// ServiceConfig Logical representation of TCP/UDP service proxy configuration
// FQSN: Fully Qualified Service Name. Format: <namespace>/<name>
// Protocol: tcp or udp
// ServicePort: Target service port
// ProxyPort Port open on proxy (aka. public port)
type ServiceConfig struct {
	FQSN        string
	Protocol    string
	ServicePort string
	ProxyPort   string
}

// RuleTag Returns the config signature to uniquely identifying and comparing configs
// Format: <protocol>:<proxy port>@<FQSN>:<service port>
func (sc ServiceConfig) RuleTag() string {
	return fmt.Sprintf("%s:%s@%s:%s", sc.Protocol, sc.ProxyPort, sc.FQSN, sc.ServicePort)
}

// ParseConfigMap Returns all proxy configs from the config map
func ParseConfigMap(protocol string, client *kubernetes.Clientset) []ServiceConfig {
	configMapClient := client.CoreV1().ConfigMaps("ingress-nginx")
	configMap, err := configMapClient.Get(fmt.Sprintf("%s-services", protocol), metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	configs := []ServiceConfig{}

	for key, value := range configMap.Data {
		config := ServiceConfig{
			strings.Split(value, ":")[0],
			"tcp",
			strings.Split(value, ":")[1],
			key,
		}
		configs = append(configs, config)
	}

	return configs
}

// GetFromService Returns all proxy configs of the given Service
func GetFromService(service *v1.Service) []ServiceConfig {
	configs := []ServiceConfig{}

	for name, annotation := range service.Annotations {
		if !strings.Contains(name, "tunack.dahus.io/") {
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
			if string(port.Protocol) == strings.ToUpper(protocol) && strconv.FormatInt(int64(port.Port), 10) == servicePort {
				servicePortExists = true
				break
			}
		}

		if !servicePortExists {
			fmt.Println(" ! Service port found in annotion but not found in service spec! Ignoring")
			continue
		}

		config := ServiceConfig{
			fmt.Sprintf("%s/%s", service.Namespace, service.Name),
			protocol,
			servicePort,
			proxyPort,
		}

		fmt.Printf(" * Service FQN: %s\n * Protocol: %s\n * Service port: %s\n * Proxy port: %s\n * Rule tag: %s\n", config.FQSN, config.Protocol, config.ServicePort, config.ProxyPort, config.RuleTag())

		configs = append(configs, config)
	}

	return configs
}

//UpdateConfigMap update configMap with service to add and to delete
func UpdateConfigMap(toAdd []ServiceConfig, toDelete []ServiceConfig, client *kubernetes.Clientset) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		configMapClient := client.CoreV1().ConfigMaps("ingress-nginx")
		updatedConfigMap, err := configMapClient.Get("tcp-services", metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		if updatedConfigMap.Data == nil {
			updatedConfigMap.Data = map[string]string{}
		}

		for _, toDeleteConfig := range toDelete {
			delete(updatedConfigMap.Data, toDeleteConfig.ProxyPort)
		}

		for _, toCreateConfig := range toAdd {
			updatedConfigMap.Data[toCreateConfig.ProxyPort] = fmt.Sprintf("%s:%s", toCreateConfig.FQSN, toCreateConfig.ServicePort)
		}

		_, updateErr := configMapClient.Update(updatedConfigMap)

		return updateErr
	})
}
