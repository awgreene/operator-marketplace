package proxy

import (
	"context"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ClusterProxyName is the name of the global config proxy.
	ClusterProxyName = "cluster"
)

var (
	// clusterProxyKey is the key for the cluster proxy.
	clusterProxyKey = types.NamespacedName{Name: ClusterProxyName}
)

// CheckDeploymentEnvVars returns true if the deployment with the given name and namespace
// needs to have its environment variables update to match those defined in the marketplace operator.
func CheckDeploymentEnvVars(client cli.Client, name, namespace string) (bool, error) {
	// Check if the Proxy API exists.
	if !IsAPIAvailable() {
		return false, nil
	}

	// Get the Deployment with the given namespacedname.
	deployment := &apps.Deployment{}
	key := types.NamespacedName{Namespace: namespace, Name: name}
	if err := client.Get(context.TODO(), key, deployment); err != nil {
		return false, err
	}

	// Get the array of EnvVar objects from the deployment and the operator.
	deploymentEnv := deployment.Spec.Template.Spec.Containers[0].Env
	operatorEnv := GetOperatorEnvVars()

	// Sort the two arrays.
	SortEnvVars(deploymentEnv)
	SortEnvVars(operatorEnv)

	// Check if the deployment environment variables are in sync with the proxy.
	if EqualEnvVars(deploymentEnv, operatorEnv) {
		return false, nil
	}

	return true, nil
}
