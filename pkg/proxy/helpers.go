package proxy

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// ClusterProxyName is the name of the global config proxy.
	ClusterProxyName = "cluster"
)

var (
	// ClusterProxyKey is the key for the cluster proxy.
	ClusterProxyKey = types.NamespacedName{Name: ClusterProxyName}
)

// NeedsUpdate return true if the deployment is out of sync with the proxy environment variables.
func NeedsUpdate(registryDeployment apps.Deployment) bool {
	// Get the list of environment variables defined in the registry deployment
	actualEnvVars := registryDeployment.Spec.Template.Spec.Containers[0].Env

	desiredEnvVars := Manage(actualEnvVars)

	if len(desiredEnvVars) != len(actualEnvVars) {
		return true
	}

	for _, envVar := range desiredEnvVars {
		position, found := GetPosition(actualEnvVars, envVar.Name)
		if !found {
			return true
		}
		if actualEnvVars[position].Value != envVar.Value {
			return true
		}
	}

	return false
}
