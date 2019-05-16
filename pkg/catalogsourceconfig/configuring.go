package catalogsourceconfig

import (
	"context"

	marketplace "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/datastore"
	"github.com/operator-framework/operator-marketplace/pkg/phase"
	"github.com/operator-framework/operator-marketplace/pkg/registry"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewConfiguringReconciler returns a Reconciler that reconciles a
// CatalogSourceConfig object in the "Configuring" phase.
func NewConfiguringReconciler(log *logrus.Entry, reader datastore.Reader, client client.Client, cache Cache) Reconciler {
	return &configuringReconciler{
		log:    log,
		reader: reader,
		client: client,
		cache:  cache,
	}
}

// configuringReconciler is an implementation of Reconciler interface that
// reconciles a CatalogSourceConfig object in the "Configuring" phase.
type configuringReconciler struct {
	log    *logrus.Entry
	reader datastore.Reader
	client client.Client
	cache  Cache
}

// Reconcile reconciles a CatalogSourceConfig object that is in the
// "Configuring" phase. It ensures that a corresponding CatalogSource object
// exists.
//
// Upon success, it returns "Succeeded" as the next and final desired phase.
// On error, the function returns "Failed" as the next desired phase
// and Message is set to the appropriate error message.
func (r *configuringReconciler) Reconcile(ctx context.Context, in *marketplace.CatalogSourceConfig) (out *marketplace.CatalogSourceConfig, nextPhase *marketplace.Phase, err error) {
	if in.Status.CurrentPhase.Name != phase.Configuring {
		err = phase.ErrWrongReconcilerInvoked
		return
	}

	out = in

	// Populate the cache before we reconcile to preserve previous data
	// in case of a failure.
	r.cache.Set(out)

	registryDeployer := registry.NewRegistryDeployer(r.log, r.reader, r.client)

	err = registryDeployer.CreateRegistryResources(in)
	if err != nil {
		nextPhase = phase.GetNextWithMessage(phase.Configuring, err.Error())
		return
	}

	registryDeployer.EnsurePackagesInStatus(out)

	nextPhase = phase.GetNext(phase.Succeeded)

	r.log.Info("The object has been successfully reconciled")
	return
}
