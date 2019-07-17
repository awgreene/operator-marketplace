package watches

import (
	"context"

	apiconfigv1 "github.com/openshift/api/config/v1"
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
	// Ensure that proxy environment variables are in sync.
	clusterProxy := &apiconfigv1.Proxy{}
	m.client.Get(context.TODO(), proxy.ClusterProxyKey, clusterProxy)
	proxy.SetEnvVars(clusterProxy)

	options := &client.ListOptions{}
	opsrcs := &v1.OperatorSourceList{}

	requests := []reconcile.Request{}
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
