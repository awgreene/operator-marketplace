package proxy

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	apiconfigv1 "github.com/openshift/api/config/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// HTTPProxy is the name of the environment variable that sets the proxy for HTTP requests.
	HTTPProxy = "HTTP_PROXY"

	// HTTPSProxy is the name of the environment variable that sets the proxy for HTTPS requests.
	HTTPSProxy = "HTTPS_PROXY"

	// NoProxy is the name of the environment variable that has a list of domains for which the proxy should not be used.
	NoProxy = "NO_PROXY"
)

// lock is used to ensure the OS Environment Variables do not experience race conditions
// when getting set or retrieved.
var lock sync.Mutex

// SetOperatorEnvVars accepts a client and updates the Operator environment variables
// based on the cluster proxy.
func SetOperatorEnvVars(client client.Client) error {
	if client == nil {
		return fmt.Errorf("Client cannot be is nil")
	}
	// Ensure that operator environment variables are in sync with those in proxy.
	clusterProxy := &apiconfigv1.Proxy{}
	err := client.Get(context.TODO(), clusterProxyKey, clusterProxy)
	if err != nil {
		return err
	}

	lock.Lock()
	defer lock.Unlock()
	// Store the old proxy values in case an error is encountered when setting
	// the new values.
	oldHTTPProxy := os.Getenv(HTTPProxy)
	oldHTTPSProxy := os.Getenv(HTTPSProxy)
	oldNoProxy := os.Getenv(NoProxy)

	newHTTPProxy := clusterProxy.Status.HTTPProxy
	newHTTPSProxy := clusterProxy.Status.HTTPSProxy
	newNoProxy := clusterProxy.Status.NoProxy

	// If any of the values have changed, update the environment variables.
	if oldHTTPProxy != newHTTPProxy ||
		oldHTTPSProxy != newHTTPSProxy ||
		oldNoProxy != newNoProxy {
		err = setProxyEnvVars(newHTTPProxy, newHTTPSProxy, newNoProxy)
		if err != nil {
			setProxyEnvVars(oldHTTPProxy, oldHTTPSProxy, oldNoProxy)
			return err
		}
		log.Infof("[proxy] %s environment variable updated to %s", HTTPProxy, clusterProxy.Status.HTTPProxy)
		log.Infof("[proxy] %s environment variable updated to %s", HTTPSProxy, clusterProxy.Status.HTTPSProxy)
		log.Infof("[proxy] %s environment variable updated to %s", NoProxy, clusterProxy.Status.NoProxy)
	}

	return nil
}

// setProxyEnvVars sets the proxy env vars to the given values.
func setProxyEnvVars(httpProxyValue, httpsProxyValue, noProxyValue string) error {
	err := os.Setenv(HTTPProxy, httpProxyValue)
	if err != nil {
		return err
	}

	err = os.Setenv(HTTPSProxy, httpsProxyValue)
	if err != nil {
		return err
	}

	err = os.Setenv(NoProxy, noProxyValue)
	if err != nil {
		return err
	}
	return nil
}

// GetOperatorEnvVars will return a list proxy EnvVars set on the operator.
func GetOperatorEnvVars() []corev1.EnvVar {
	lock.Lock()
	defer lock.Unlock()
	return []corev1.EnvVar{
		corev1.EnvVar{Name: NoProxy, Value: os.Getenv(NoProxy)},
		corev1.EnvVar{Name: HTTPProxy, Value: os.Getenv(HTTPProxy)},
		corev1.EnvVar{Name: HTTPSProxy, Value: os.Getenv(HTTPSProxy)},
	}
}

// SortEnvVars will sort a list of EnvVar objects alphabetically by thier name.
func SortEnvVars(envVars []corev1.EnvVar) {
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
}

// EqualEnvVars checks if the two lists of Env Vars are contain the same elements in the same order.
func EqualEnvVars(a, b []corev1.EnvVar) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
