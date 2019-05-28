package catalogsourceconfig_test

import (
	"testing"

	"github.com/operator-framework/operator-marketplace/pkg/catalogsourceconfig"

	"github.com/stretchr/testify/assert"
)

func TestRemoveNamespaces(t *testing.T) {
	packages := "community-operators/jager,certified-operators/orca"
	expected := "jager,orca"

	actual := catalogsourceconfig.RemoveNamespaces(packages)
	assert.Equal(t, expected, actual)
}

func TestRemoveNamespaces2(t *testing.T) {
	packages := "jager,orca"
	expected := "jager,orca"

	actual := catalogsourceconfig.RemoveNamespaces(packages)
	assert.Equal(t, expected, actual)
}

func TestRemoveNamespaces3(t *testing.T) {
	packages := "some/jager,silly/orca,test"
	expected := "jager,orca,test"

	actual := catalogsourceconfig.RemoveNamespaces(packages)
	assert.Equal(t, expected, actual)
}
