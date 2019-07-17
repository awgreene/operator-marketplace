package testsuites

import (
	"context"
	"testing"

	apiconfigv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/operator-marketplace/test/helpers"
	"github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ProxyTests is a test suite that ensure that the watches for proxy resources
// are firing correctly and the deployment resources are updated as a result.
func ProxyTests(t *testing.T) {
	t.Run("no-proxy", testSetNoProxyVariable)
}

// testSetNoProxyVariable creates an OperatorSource and then set the cluster proxy.Spec.NoProxy
// to ensure that the registry deployment is updated to include the proxy variable.
func testSetNoProxyVariable(t *testing.T) {
	// Create a ctx that is used to delete the OperatorSource and CatalogSouceConfig at the
	// completion of this function.
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	// Get test namespace.
	namespace, err := ctx.GetNamespace()
	require.NoError(t, err, "Could not get namespace")

	// Get global framework variables
	client := test.Global.Client

	// Get the cluster proxy
	clusterProxyKey := types.NamespacedName{Name: "cluster", Namespace: ""}
	clusterProxy := &apiconfigv1.Proxy{}
	err = client.Get(context.TODO(), clusterProxyKey, clusterProxy)
	require.NoError(t, err, "Could not retrieve the clusterProxy")

	// Set the NoProxy value in the cluster proxy spec and update the clutser proxy.
	clusterProxy.Spec.NoProxy = "some-value"
	err = client.Update(context.TODO(), clusterProxy)
	require.NoError(t, err, "Could not update the clusterProxy")

	// Get the deployment that should have been updated by the watch on the proxy.
	resultDeployment := &apps.Deployment{}
	err = helpers.WaitForResult(client, resultDeployment, namespace, helpers.TestOperatorSourceName)
	require.NoError(t, err, "Could not get the deployment")

	// Check if the correct env vars are set.
	actualEnvVars := resultDeployment.Spec.Template.Spec.Containers[0].Env
	for _, v := range actualEnvVars {
		if v.Name == "NO_PROXY" {
			assert.Equal(t, v.Value, "some-value")
		} else {
			assert.Equal(t, v.Value, "")
		}
	}

	// Remove the NoProxy value in the proxy spec.
	clusterProxy.Spec.NoProxy = ""
	err = client.Update(context.TODO(), clusterProxy)
	require.NoError(t, err, "Could not update the clusterProxy")
}
