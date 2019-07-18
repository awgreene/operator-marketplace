package watches

import (
	"context"

	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/proxy"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type proxyToOperatorSources struct {
	client client.Client
}

// Map will update the proxy environment variables and create a reconcile request for
// all OperatorSources.
func (m *proxyToOperatorSources) Map(obj handler.MapObject) []reconcile.Request {
	requests := []reconcile.Request{}

	// Ensure that operator environment variables are in sync with those in proxy.
	err := proxy.SetOperatorEnvVars(m.client)
	if err != nil {
		return requests
	}

	options := &client.ListOptions{}
	opsrcs := &v1.OperatorSourceList{}

	if err := m.client.List(context.TODO(), options, opsrcs); err != nil {
		return requests
	}

	for _, opsrc := range opsrcs.Items {
		requests = append(requests, reconcile.Request{types.NamespacedName{Name: opsrc.GetName(), Namespace: opsrc.GetNamespace()}})
	}
	return requests
}

// ProxyToOperatorSources returns a mapper that maps the proxy to the OperatorSources.
func ProxyToOperatorSources(client client.Client) handler.Mapper {
	return &proxyToOperatorSources{client: client}
}
