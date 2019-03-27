package e2e

import (
	"fmt"
	"testing"

	operator "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-sdk/pkg/test"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	marketplace "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
)

const (
	// nonExistingTargetNamespaceCSCName is the name of the catalogsourceconfig that points
	// to a non-existing targetNamespace
	cscName string = "other-namespace-operators"

	// targetNamespace is the non-existing target namespace
	targetNamespace string = "other-namespace"
)

func runCSCOtherNamespace(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	// Get global framework variables
	f := test.Global
	// Run tests
	if err := cscOtherNamespace(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func cscOtherNamespace(t *testing.T, f *test.Framework, ctx *test.TestCtx) error {
	// Get test namespace
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	// Create the operatorsource to download the manifests
	testOperatorSource := &operator.OperatorSource{
		TypeMeta: metav1.TypeMeta{
			Kind: operator.OperatorSourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operators",
			Namespace: namespace,
		},
		Spec: operator.OperatorSourceSpec{
			Type:              "appregistry",
			Endpoint:          "https://quay.io/cnr",
			RegistryNamespace: "marketplace_e2e",
		},
	}
	err = createRuntimeObject(f, ctx, testOperatorSource)
	if err != nil {
		return err
	}

	// Create a namespace based on the targetNamespace string
	targetNamespaceRuntimeObject := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: targetNamespace}}
	err = createRuntimeObject(f, ctx, targetNamespaceRuntimeObject)
	if err != nil {
		return err
	}

	// Create a new catalogsourceconfig with a non-existing targetNamespace
	otherTargetNamespaceCSC := &operator.CatalogSourceConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: operator.OperatorSourceKind,
		}, ObjectMeta: metav1.ObjectMeta{
			Name:      cscName,
			Namespace: namespace,
		},
		Spec: operator.CatalogSourceConfigSpec{
			TargetNamespace: targetNamespace,
			Packages:        "descheduler",
		}}
	err = createRuntimeObject(f, nil, otherTargetNamespaceCSC)
	if err != nil {
		return err
	}

	// Check that we created the catalogsourceconfig.
	resultCatalogSourceConfig := &marketplace.CatalogSourceConfig{}
	err = WaitForResult(t, f, resultCatalogSourceConfig, namespace, cscName)
	if err != nil {
		return err
	}

	// Then check for the catalog source.
	resultCatalogSource := &olm.CatalogSource{}
	err = WaitForResult(t, f, resultCatalogSource, targetNamespace, cscName)
	if err != nil {
		return err
	}

	// Then check that the service was created.
	resultService := &corev1.Service{}
	err = WaitForResult(t, f, resultService, namespace, cscName)
	if err != nil {
		return err
	}

	// Then check that the deployment was created.
	resultDeployment := &apps.Deployment{}
	err = WaitForResult(t, f, resultDeployment, namespace, cscName)
	if err != nil {
		return err
	}

	// Now check that the deployment is ready.
	err = WaitForSuccessfulDeployment(t, f, *resultDeployment)
	if err != nil {
		return err
	}

	// delete the csc
	err = deleteRuntimeObject(f, ctx, otherTargetNamespaceCSC)
	if err != nil {
		return err
	}

	// check that all objects are deleted
	// Check that we deleted the catalogsourceconfig.
	err = WaitForNotFound(t, f, resultCatalogSourceConfig, namespace, cscName)
	if err != nil {
		return err
	}

	// Then check that the catalog source was deleted.
	err = WaitForNotFound(t, f, resultCatalogSource, targetNamespace, cscName)
	if err != nil {
		return err
	}

	// Then check that the service was deleted.
	err = WaitForNotFound(t, f, resultService, namespace, cscName)
	if err != nil {
		return err
	}

	// Then check that the deployment was deleted.
	err = WaitForNotFound(t, f, resultDeployment, namespace, cscName)
	if err != nil {
		return err
	}

	return nil
}
