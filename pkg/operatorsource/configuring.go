package operatorsource

import (
	"context"

	marketplace "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/datastore"
	"github.com/operator-framework/operator-marketplace/pkg/phase"
	"github.com/operator-framework/operator-marketplace/pkg/registry"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewConfiguringReconciler returns a Reconciler that reconciles
// an OperatorSource object in "Configuring" phase.
func NewConfiguringReconciler(logger *log.Entry, datastore datastore.Writer, reader datastore.Reader, client client.Client) Reconciler {
	return &configuringReconciler{
		logger:    logger,
		datastore: datastore,
		client:    client,
		reader:    reader,
	}
}

// configuringReconciler is an implementation of Reconciler interface that
// reconciles an OperatorSource object in "Configuring" phase.
type configuringReconciler struct {
	logger    *log.Entry
	datastore datastore.Writer
	client    client.Client
	reader    datastore.Reader
}

// Reconcile reconciles an OperatorSource object that is in "Configuring" phase.
// It ensures that a corresponding CatalogSourceConfig object exists.
//
// in represents the original OperatorSource object received from the sdk
// and before reconciliation has started.
//
// out represents the OperatorSource object after reconciliation has completed
// and could be different from the original. The OperatorSource object received
// (in) should be deep copied into (out) before changes are made.
//
// nextPhase represents the next desired phase for the given OperatorSource
// object. If nil is returned, it implies that no phase transition is expected.
//
// Upon success, it returns "Succeeded" as the next and final desired phase.
// On error, the function returns "Failed" as the next desied phase
// and Message is set to appropriate error message.
//
// If the corresponding CatalogSourceConfig object already exists
// then no further action is taken.
func (r *configuringReconciler) Reconcile(ctx context.Context, in *marketplace.OperatorSource) (out *marketplace.OperatorSource, nextPhase *marketplace.Phase, err error) {
	if in.GetCurrentPhaseName() != phase.Configuring {
		err = phase.ErrWrongReconcilerInvoked
		return
	}

	out = in

	manifests := r.datastore.GetPackageIDsByOperatorSource(in.GetUID())

	cscCreate := new(CatalogSourceConfigBuilder).WithTypeMeta().
		WithNamespacedName(in.Namespace, in.Name).
		WithLabels(in.GetLabels()).
		WithSpec(in.Namespace, manifests, in.Spec.DisplayName, in.Spec.Publisher).
		WithOwnerLabel(in).
		CatalogSourceConfig()

	registryDeployer := registry.NewRegistryDeployer(r.logger, r.reader, r.client)

	err = registryDeployer.CreateRegistryResources(cscCreate)
	if err != nil {
		nextPhase = phase.GetNextWithMessage(phase.Configuring, err.Error())
		return
	}

	registryDeployer.EnsurePackagesInStatus(cscCreate)

	nextPhase = phase.GetNext(phase.Succeeded)
	return
}
