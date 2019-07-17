package proxy_test

import (
	"os"
	"testing"

	//apiconfigv1 "github.com/openshift/api/config/v1"
	apiconfigv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/operator-marketplace/pkg/proxy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

const (
	NO_Proxy    = "NO_PROXY"
	HTTP_Proxy  = "HTTP_PROXY"
	HTTPS_Proxy = "HTTPS_PROXY"
)

func resetProxyVars() {
	os.Setenv(NO_Proxy, "")
	os.Setenv(HTTP_Proxy, "")
	os.Setenv(HTTPS_Proxy, "")
}

func TestSetEnvVars(t *testing.T) {
	defer resetProxyVars()
	testProxy := &apiconfigv1.Proxy{}
	testProxy.Spec.NoProxy = "bar1"
	testProxy.Spec.HTTPProxy = "bar2"
	testProxy.Spec.HTTPSProxy = "bar3"
	proxy.SetEnvVars(testProxy)

	assert.Equal(t, "bar1", os.Getenv(NO_Proxy))
	assert.Equal(t, "bar2", os.Getenv(HTTP_Proxy))
	assert.Equal(t, "bar3", os.Getenv(HTTPS_Proxy))
}

func TestGetIndex(t *testing.T) {
	list := []corev1.EnvVar{
		corev1.EnvVar{Name: "Foo1", Value: "Bar1"},
		corev1.EnvVar{Name: "Foo2", Value: "Bar2"},
	}

	position, found := proxy.GetPosition(list, "Foo1")
	assert.True(t, found)
	assert.Equal(t, 0, position)

	position, found = proxy.GetPosition(list, "Foo2")
	assert.True(t, found)
	assert.Equal(t, 1, position)

	position, found = proxy.GetPosition(list, "Foo3")
	assert.False(t, found)
	assert.Equal(t, -1, position)
}

func TestRemove(t *testing.T) {
	list := []corev1.EnvVar{
		corev1.EnvVar{Name: "Foo1", Value: "Bar1"},
		corev1.EnvVar{Name: "Foo2", Value: "Bar2"},
		corev1.EnvVar{Name: "Foo3", Value: "Bar3"},
	}

	// Remove the second element
	list = proxy.Remove(list, 1)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, corev1.EnvVar{Name: "Foo1", Value: "Bar1"}, list[0])
	assert.Equal(t, corev1.EnvVar{Name: "Foo3", Value: "Bar3"}, list[1])

	// Remove the first element
	list = proxy.Remove(list, 0)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, corev1.EnvVar{Name: "Foo3", Value: "Bar3"}, list[0])

	// Remove the remaining element
	list = proxy.Remove(list, 0)
	assert.Equal(t, 0, len(list))
}

func TestManageNilList(t *testing.T) {
	// If the provided list is nil and No Proxy Env Vars are set, nil should be returned.
	list := proxy.Manage(nil)
	assert.Nil(t, list)
}

func TestManage(t *testing.T) {
	defer resetProxyVars()
	list := []corev1.EnvVar{
		corev1.EnvVar{Name: NO_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTP_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTPS_Proxy, Value: ""},
	}

	// No Proxy Env Vars are set, should remove all items
	list = proxy.Manage(list)
	assert.Equal(t, 0, len(list))

	list = []corev1.EnvVar{
		corev1.EnvVar{Name: NO_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTP_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTPS_Proxy, Value: ""},
		corev1.EnvVar{Name: "Foo", Value: "Bar"},
	}

	// No Proxy Env Vars are set, should leave EnvVar Foo
	list = proxy.Manage(list)
	assert.Equal(t, 1, len(list))

	// Set an Proxy Env Var
	os.Setenv(NO_Proxy, "test1")
	list = proxy.Manage(list)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, "Foo", list[0].Name)
	assert.Equal(t, "Bar", list[0].Value)
	assert.Equal(t, NO_Proxy, list[1].Name)
	assert.Equal(t, "test1", list[1].Value)

	// Set an additional Proxy Env Var
	os.Setenv(HTTP_Proxy, "test2")
	list = proxy.Manage(list)
	assert.Equal(t, 3, len(list))
	assert.Equal(t, "Foo", list[0].Name)
	assert.Equal(t, "Bar", list[0].Value)
	assert.Equal(t, NO_Proxy, list[1].Name)
	assert.Equal(t, "test1", list[1].Value)
	assert.Equal(t, HTTP_Proxy, list[2].Name)
	assert.Equal(t, "test2", list[2].Value)

	// Update the HTTP_PROXY Env Var
	os.Setenv(HTTP_Proxy, "updated")
	list = proxy.Manage(list)
	assert.Equal(t, 3, len(list))
	assert.Equal(t, "Foo", list[0].Name)
	assert.Equal(t, "Bar", list[0].Value)
	assert.Equal(t, NO_Proxy, list[1].Name)
	assert.Equal(t, "test1", list[1].Value)
	assert.Equal(t, HTTP_Proxy, list[2].Name)
	assert.Equal(t, "updated", list[2].Value)

	// Remove the HTTP_PROXY Env Var
	os.Setenv(HTTP_Proxy, "")
	list = proxy.Manage(list)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, "Foo", list[0].Name)
	assert.Equal(t, "Bar", list[0].Value)
	assert.Equal(t, NO_Proxy, list[1].Name)
	assert.Equal(t, "test1", list[1].Value)
}
