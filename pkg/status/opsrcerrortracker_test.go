package status_test

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/operator-marketplace/pkg/operatorhub"
	"github.com/operator-framework/operator-marketplace/pkg/status"
	"github.com/stretchr/testify/assert"
)

func TestOpSrcErrorTrackerAddAndRemove(t *testing.T) {
	// Create the OpSrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()
	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))

	// Add an entry to the OpSrcErrorTracker.
	expectedKey := "test"
	expectedValue := status.NewStatusError("test-err", nil)
	opsrcErrorTracker.Add(expectedKey, expectedValue)

	// Confirm that the expectes list of keys and values are returned.
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 1, expectedKey, expectedValue))

	// Add a second entry to the OpSrcErrorTracker.
	secondExpectedKey := "test2"
	secondExpectedValue := status.NewStatusError("test-err2", nil)
	opsrcErrorTracker.Add(secondExpectedKey, secondExpectedValue)

	// Confirm that the expectes list of keys and values are returned.
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 2, expectedKey, expectedValue))
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 2, secondExpectedKey, secondExpectedValue))

	// Remove the second entry from the OpSrcErrorTracker.
	opsrcErrorTracker.Remove(secondExpectedKey)
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 1, expectedKey, expectedValue))

	// Remove the remaining entry from the OpSrcErrorTracker.
	opsrcErrorTracker.Remove(expectedKey)
	actualKeys, actualMap = opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))
}

func TestOpSrcErrorTrackerRemoveMissingKey(t *testing.T) {
	// Create the OpSrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()
	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))

	// Attempt to remove a key that is not present in the OpSrcErrorTracker.
	// Ensure no errors are thrown.
	opsrcErrorTracker.Remove("Missing key")
}

func TestOpsrcErrorTrackerClone(t *testing.T) {
	// Create the OpSrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()
	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))

	// Add an entry to the OpSrcErrorTracker.
	expectedKey := "test"
	expectedValue := status.NewStatusError("test-err", nil)
	opsrcErrorTracker.Add(expectedKey, expectedValue)

	// Confirm that the expectes list of keys and values are returned.
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 1, expectedKey, expectedValue))

	// Remove the existing key from the clone.
	keys, clone := opsrcErrorTracker.GetKeysAndMap()
	delete(clone, keys[0])
	assert.Equal(t, 0, len(clone))

	// Confirm that the expectes list of keys and values are still present in the opsrcErrorTracker.
	assert.True(t, checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker, 1, expectedKey, expectedValue))
}

func TestOpSrcErrorTrackerSyncEmptyHubSources(t *testing.T) {
	// Create the OpSrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()

	// Add an entry to the OpSrcErrorTracker.
	expectedKey := "test"
	expectedValue := status.NewStatusError("test-err", nil)
	opsrcErrorTracker.Add(expectedKey, expectedValue)
	operatorhub.GetSingleton().Set(configv1.OperatorHubSpec{})
	opsrcErrorTracker.Sync(operatorhub.GetSingleton())

	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))
}

func TestOpSrcErrorTrackerSyncLeaveEnabled(t *testing.T) {
	// Create the OpSrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()

	// Add an entry to the OpSrcErrorTracker.
	expectedKey := "test"
	expectedValue := status.NewStatusError("test-err", nil)
	opsrcErrorTracker.Add(expectedKey, expectedValue)
	operatorhub.GetSingleton().Set(configv1.OperatorHubSpec{
		Sources: []configv1.HubSource{
			configv1.HubSource{
				Name:     expectedKey,
				Disabled: false},
		},
	})
	defer resetOperatorHubSingleton()
	opsrcErrorTracker.Sync(operatorhub.GetSingleton())

	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{expectedKey}, actualKeys)
	assert.Equal(t, 1, len(actualMap))
}

func TestOpSrcErrorTrackerSyncRemoveDisabled(t *testing.T) {
	// Create the opsrcErrorTracker and assert that there are no entries.
	opsrcErrorTracker := status.NewOpSrcErrorTracker()

	// Add an entry to the opsrcErrorTracker.
	expectedKey := "test"
	expectedValue := status.NewStatusError("test-err", nil)
	opsrcErrorTracker.Add(expectedKey, expectedValue)

	operatorhub.GetSingleton().Set(configv1.OperatorHubSpec{
		Sources: []configv1.HubSource{
			configv1.HubSource{
				Name:     expectedKey,
				Disabled: true},
		},
	})
	defer resetOperatorHubSingleton()
	opsrcErrorTracker.Sync(operatorhub.GetSingleton())

	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	assert.Equal(t, []string{}, actualKeys)
	assert.Equal(t, 0, len(actualMap))
}

func checkOpSrcErrorTrackerForLengthAndKeyValue(opsrcErrorTracker status.OpSrcErrorTracker, expectedLength int, expectedKey string, expectedValue error) bool {
	actualKeys, actualMap := opsrcErrorTracker.GetKeysAndMap()
	if expectedLength != len(actualMap) || !contains(actualKeys, expectedKey) || expectedValue != actualMap[expectedKey] {
		return false
	}

	return true
}

func contains(slice []string, target string) bool {
	for i := range slice {
		if slice[i] == target {
			return true
		}
	}
	return false
}

func resetOperatorHubSingleton() {
	operatorhub.GetSingleton().Set(configv1.OperatorHubSpec{})
}
