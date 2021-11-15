package servicediscovery

import (
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

// Client is the external client
type Client interface {
	servicediscoveryiface.ServiceDiscoveryAPI
}
