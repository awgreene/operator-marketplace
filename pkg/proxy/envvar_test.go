package proxy_test

import (
	"os"
	"testing"

	"github.com/operator-framework/operator-marketplace/pkg/proxy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

const (
	NO_Proxy    = "NO_PROXY"
	HTTP_Proxy  = "HTTP_PROXY"
	HTTPS_Proxy = "HTTPS_PROXY"
)

func TestGetOperatorEnvVars(t *testing.T) {
	defer func() {
		// Remove env vars set during test:
		os.Setenv(NO_Proxy, "")
		os.Setenv(HTTP_Proxy, "")
		os.Setenv(HTTPS_Proxy, "")
	}()

	// Ensure that PROXY variables are not set.
	os.Setenv(NO_Proxy, "")
	os.Setenv(HTTP_Proxy, "")
	os.Setenv(HTTPS_Proxy, "")

	expected := []corev1.EnvVar{
		corev1.EnvVar{Name: NO_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTP_Proxy, Value: ""},
		corev1.EnvVar{Name: HTTPS_Proxy, Value: ""},
	}

	actual := proxy.GetOperatorEnvVars()
	assert.Equal(t, expected, actual)
}

func TestEqualEnvVars(t *testing.T) {
	// Test nil
	assert.True(t, proxy.EqualEnvVars(nil, nil))

	// Test empty array
	assert.True(t, proxy.EqualEnvVars([]corev1.EnvVar{}, []corev1.EnvVar{}))

	// Test same list
	a := []corev1.EnvVar{
		corev1.EnvVar{Name: "Foo"},
	}

	b := []corev1.EnvVar{
		corev1.EnvVar{Name: "Foo"},
	}
	assert.True(t, proxy.EqualEnvVars(a, b))

	// Test order
	a = []corev1.EnvVar{
		corev1.EnvVar{Name: "Foo"},
		corev1.EnvVar{Name: "Bar"},
	}

	b = []corev1.EnvVar{
		corev1.EnvVar{Name: "Bar"},
		corev1.EnvVar{Name: "Foo"},
	}
	assert.False(t, proxy.EqualEnvVars(a, b))
}
