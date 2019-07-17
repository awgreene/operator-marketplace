package operatorsource

import (
	"context"

	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/shared"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/phase"
	"github.com/operator-framework/operator-marketplace/pkg/watches"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewSucceededReconciler returns a Reconciler that reconciles
// an OperatorSource object in "Succeeded" phase.
func NewSucceededReconciler(logger *log.Entry, client client.Client) Reconciler {
	return &succeededReconciler{
		logger: logger,
		client: client,
	}
}

// succeededReconciler is an implementation of Reconciler interface that
// reconciles an OperatorSource object in "Succeeded" phase.
type succeededReconciler struct {
	logger *log.Entry
	client client.Client
}

// Reconcile reconciles an OperatorSource object that is in "Succeeded" phase.
// Since this phase indicates that the object has been successfully reconciled,
// no further action is taken.
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
func (r *succeededReconciler) Reconcile(ctx context.Context, in *v1.OperatorSource) (out *v1.OperatorSource, nextPhase *shared.Phase, err error) {
	if in.GetCurrentPhaseName() != phase.Succeeded {
		err = phase.ErrWrongReconcilerInvoked
		return
	}

	// No change is being made, so return the OperatorSource object that was specified as is.
	out = in

	msg := "No action taken, the object has already been reconciled"
	defer func() {
		r.logger.Info(msg)
	}()

	needsProxyUpdate, err := watches.CheckProxyResource(r.client, in.Name, in.Namespace)
	if err != nil {
		return
	}

	if needsProxyUpdate {
		// The Environment Variables in the deployment created by the CatalogSourceConfig
		// are out of sync with those in the Cluster Proxy.
		// Drop the existing Status field so that reconciliation can start anew.
		out.Status = v1.OperatorSourceStatus{}
		nextPhase = phase.GetNext(phase.Configuring)
		msg = "Proxy environment variables not in sync, scheduling for configuring"
		return
	}

	return out, nil, nil
}
