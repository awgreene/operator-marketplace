package proxy

import (
	"os"

	apiconfigv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// HTTP_PROXY is the URL of the proxy for HTTP requests.
	envHTTPProxyName = "HTTP_PROXY"

	// HTTPS_PROXY is the URL of the proxy for HTTPS requests.
	envHTTPSProxyName = "HTTPS_PROXY"

	// NO_PROXY is the list of domains for which the proxy should not be used.
	envNoProxyName = "NO_PROXY"
)

var (
	allProxyEnvVarNames = []string{
		envHTTPProxyName,
		envHTTPSProxyName,
		envNoProxyName,
	}
)

// SetEnvVars accepts a Proxy object and sets the proxy variables
// as environment variables.
func SetEnvVars(proxy *apiconfigv1.Proxy) {
	if proxy == nil {
		return
	}
	os.Setenv(envHTTPProxyName, proxy.Spec.HTTPProxy)
	os.Setenv(envHTTPSProxyName, proxy.Spec.HTTPSProxy)
	os.Setenv(envNoProxyName, proxy.Spec.NoProxy)
}

// GetPosition returns the index of the EnvVar with the given name.
// If the list does not contain an EnvVar with the given name, -1 is returned.
func GetPosition(list []corev1.EnvVar, name string) (position int, found bool) {
	for i := range list {
		if name == list[i].Name {
			return i, true
		}
	}
	return -1, false
}

// Remove removes the given index from the list
func Remove(list []corev1.EnvVar, index int) []corev1.EnvVar {
	return append(list[:index], list[index+1:]...)
}

// Manage will ensure that the list of EnvVars are in sync with the proxy variables.
// If the proxy variables are not set but are present in the list, they will be removed from the list.
// If the proxy variables are set but missing from the list, they will be added to the list.
// If the proxy variables are set and present in the list, their values will be updated in the list.
func Manage(list []corev1.EnvVar) []corev1.EnvVar {
	// For each proxy variable
	for _, proxyVarName := range allProxyEnvVarNames {
		// Get the value associated with the proxy variable.
		proxyVarValue := os.Getenv(proxyVarName)

		// Check if the list of EnvVars contains an EnvVar with the given Name
		position, found := GetPosition(list, proxyVarName)

		if found {
			// If the proxy is not set, remove it from the list
			if proxyVarValue == "" {
				list = Remove(list, position)
				continue
			} else {
				// If the proxy is  set, update the existing value
				list[position] = corev1.EnvVar{Name: proxyVarName, Value: proxyVarValue}
				continue
			}
		} else {
			// If the EnvVar was not found and the proxy variable is set, add it to the list.
			if proxyVarValue != "" {
				list = append(list, corev1.EnvVar{Name: proxyVarName, Value: proxyVarValue})
				continue
			}
		}
	}
	return list
}
