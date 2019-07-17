package watches

import (
	"context"
	"fmt"

	apiconfigv1 "github.com/openshift/api/config/v1"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v2"
	"github.com/operator-framework/operator-marketplace/pkg/builders"
	"github.com/operator-framework/operator-marketplace/pkg/proxy"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// CheckChildResources returns true if any of the child resources are missing
func CheckChildResources(client cl.Client, name, namespace, targetNamespace string, secretIsPresent bool) bool {
	var err error

	// CatalogSource lives in the target Namespace
	key := cl.ObjectKey{
		Name:      name,
		Namespace: targetNamespace,
	}

	// Check if the CatalogSource exists
	if err = client.Get(context.TODO(), key, new(builders.CatalogSourceBuilder).WithTypeMeta().CatalogSource()); err != nil {
		return true
	}

	// Other child resources lives in the object's Namespace
	key = cl.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	// Check if the Deployment exists
	if err = client.Get(context.TODO(), key, new(builders.DeploymentBuilder).WithTypeMeta().Deployment()); err != nil {
		return true
	}

	// Check if the Service exists
	if err = client.Get(context.TODO(), key, new(builders.ServiceBuilder).WithTypeMeta().Service()); err != nil {
		return true
	}

	if !secretIsPresent {
		return false
	}

	// The OperatorSource has an authorization token which implies that a ServiceAccount, Role and RoleBinding are
	// associated with this resource.

	// Check if the ServiceAccount exists
	if err = client.Get(context.TODO(), key, new(builders.ServiceAccountBuilder).WithTypeMeta().ServiceAccount()); err != nil {
		return true
	}

	// Check if the Role exists
	if err = client.Get(context.TODO(), key, new(builders.RoleBuilder).WithTypeMeta().Role()); err != nil {
		return true
	}

	// Check if the RoleBinding exists
	if err = client.Get(context.TODO(), key, new(builders.RoleBindingBuilder).WithTypeMeta().RoleBinding()); err != nil {
		return true
	}

	return false
}

// WatchChildResourcesDeletionEvents adds watches for CatalogSource, Deployment, Service,
// ServiceAccount, Roles and RoleBinding deletion events.
func WatchChildResourcesDeletionEvents(c controller.Controller, client cl.Client, owner string) error {
	// We only care if the resource was deleted, so add a predicate that returns
	// false for all events except for delete. The DeleteFunc will change depending
	// on the owner.
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	var enqueueRequestsFromMapFunc handler.EnqueueRequestsFromMapFunc
	switch owner {
	case v2.CatalogSourceConfigKind:
		pred.DeleteFunc = cscDeleteFunc
		enqueueRequestsFromMapFunc = handler.EnqueueRequestsFromMapFunc{ToRequests: ChildResourceToCatalogSourceConfig(client)}
	default:
		return fmt.Errorf("Unknown owner %s", owner)
	}

	err := c.Watch(&source.Kind{Type: &olm.CatalogSource{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &apps.Deployment{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &core.Service{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &core.ServiceAccount{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &core.ServiceAccount{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.Role{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.RoleBinding{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	return nil
}

// cscDeleteFunc is the predicate function for checking delete events. It returns
// true only if the object's owner is a CatalogSourceConfig.
func cscDeleteFunc(e event.DeleteEvent) bool {
	// If DeleteStateUnknown is true it implies that the Delete event was missed
	// and we can ignore it.
	if e.DeleteStateUnknown {
		return false
	}

	if getCscOwnerKey(e.Meta.GetLabels()) == nil {
		return false
	}
	return true
}

// getCscOwnerKey checks for the CatalogSourceConfig owner labels within the
// labels of the child resource  and computes the key if present. It returns nil
// if they are not present.
func getCscOwnerKey(labels map[string]string) *client.ObjectKey {
	ownerNamespace, present := labels[builders.CscOwnerNamespaceLabel]
	if !present {
		return nil
	}

	ownerName, present := labels[builders.CscOwnerNameLabel]
	if !present {
		return nil
	}

	return &client.ObjectKey{Namespace: ownerNamespace, Name: ownerName}
}

// WatchProxyEvents adds watches for proxy events.
func WatchProxyEvents(c controller.Controller, client cl.Client, owner string) error {
	// We only care if the event came from the cluster proxy.
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == proxy.ClusterProxyName {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// If the cluster operator is ever changed we should kick off an event.
			if e.MetaOld.GetName() == proxy.ClusterProxyName || e.MetaNew.GetName() == proxy.ClusterProxyName {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetName() == proxy.ClusterProxyName {
				return true
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			if e.Meta.GetName() == proxy.ClusterProxyName {
				return true
			}
			return false
		},
	}

	var enqueueRequestsFromMapFunc handler.EnqueueRequestsFromMapFunc
	switch owner {
	case v2.CatalogSourceConfigKind:
		enqueueRequestsFromMapFunc = handler.EnqueueRequestsFromMapFunc{ToRequests: ProxyToCatalogSourceConfigs(client)}
	case v1.OperatorSourceKind:
		enqueueRequestsFromMapFunc = handler.EnqueueRequestsFromMapFunc{ToRequests: ProxyToOperatorSources(client)}
	default:
		return fmt.Errorf("Unknown owner %s", owner)
	}

	err := c.Watch(&source.Kind{Type: &apiconfigv1.Proxy{}}, &enqueueRequestsFromMapFunc, pred)
	if err != nil {
		return err
	}

	return nil
}

// CheckProxyResource returns true if the deployment with the given namespacedname
// needs to have its environment variables update due to a change in the cluster proxy.
func CheckProxyResource(client cl.Client, name, namespace string) (bool, error) {
	// Check if the Proxy API exists.
	if !proxy.IsAPIAvailable() {
		return false, nil
	}

	// Get the Deployment with the given namespacedname.
	deployment := &apps.Deployment{}
	key := types.NamespacedName{Namespace: namespace, Name: name}
	if err := client.Get(context.TODO(), key, deployment); err != nil {
		return false, err
	}

	// Check if the deployment environment variables are in sync with the proxy.
	if proxy.NeedsUpdate(*deployment) {
		return true, nil
	}

	return false, nil
}
