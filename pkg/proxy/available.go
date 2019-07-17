package proxy

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	apidiscovery "k8s.io/client-go/discovery"
)

const (
	// This is the error message thrown by ServerSupportsVersion function
	// when an API version is not supported by the server.
	notSupportedErrorMessage = "server does not support API version"
)

// isAPIAvailable tracks if the proxy API is available.
var isAPIAvailable = false

// SetProxyAvailability will set isAPIAvailable to the correct value if
// no unexpected errors are encountered.
func SetProxyAvailability(discovery apidiscovery.DiscoveryInterface) error {
	if discovery == nil {
		return errors.New("discovery interface can not be <nil>")
	}

	opStatusGV := schema.GroupVersion{
		Group:   "config.openshift.io",
		Version: "v1",
	}

	if discoveryErr := apidiscovery.ServerSupportsVersion(discovery, opStatusGV); discoveryErr != nil {
		if strings.Contains(discoveryErr.Error(), notSupportedErrorMessage) {
			return nil
		}

		return discoveryErr
	}

	isAPIAvailable = true
	return nil
}

// IsAPIAvailable returns whether or not the proxy API is available.
func IsAPIAvailable() bool {
	return isAPIAvailable
}
