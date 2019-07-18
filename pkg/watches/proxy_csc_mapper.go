package watches

import (
	"context"

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
	requests := []reconcile.Request{}

	// Ensure that operator environment variables are in sync with those in proxy.
	err := proxy.SetOperatorEnvVars(m.client)
	if err != nil {
		return requests
	}

	options := &client.ListOptions{}
	cscs := &v2.CatalogSourceConfigList{}

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
