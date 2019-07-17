package watches

import (
	"context"

	apiconfigv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v2"
	"github.com/operator-framework/operator-marketplace/pkg/proxy"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type proxyToCatalogSourceConfigs struct {
	client client.Client
}

// Map will update the proxy environment variables and create a reconcile request for
// all CatalogSourceConfig.
func (m *proxyToCatalogSourceConfigs) Map(obj handler.MapObject) []reconcile.Request {
	// Ensure that proxy environment variables are in sync.
	clusterProxy := &apiconfigv1.Proxy{}
	m.client.Get(context.TODO(), proxy.ClusterProxyKey, clusterProxy)
	proxy.SetEnvVars(clusterProxy)

	options := &client.ListOptions{}
	cscs := &v2.CatalogSourceConfigList{}

	requests := []reconcile.Request{}
	if err := m.client.List(context.TODO(), options, cscs); err != nil {
		return requests
	}

	for _, csc := range cscs.Items {
		requests = append(requests, reconcile.Request{types.NamespacedName{Name: csc.GetName(), Namespace: csc.GetNamespace()}})
	}
	return requests
}

// ProxyToCatalogSourceConfigs returns a mapper that maps the proxy to the CatalogSourceConfigs.
func ProxyToCatalogSourceConfigs(client client.Client) handler.Mapper {
	return &proxyToCatalogSourceConfigs{client: client}
}
