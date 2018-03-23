package main

import "fmt"

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
