package status

import (
	"sync"

	"github.com/operator-framework/operator-marketplace/pkg/operatorhub"
)

// OpSrcErrorTracker is an interface that provides functions for tracking errors associated
// with OperatorSources.
type OpSrcErrorTracker interface {
	// GetKeysAndMap returns a map of errors and the list of keys present in the map.
	GetKeysAndMap() ([]string, map[string]error)

	// Add adds the error to the error map with the given key.
	Add(key string, err error)

	// Remove removes the given key from the err map.
	Remove(key string)

	// Sync removes entries from the err map that are not
	// present and enabled in the provided OperatorHub object.
	Sync(operatorHub operatorhub.OperatorHub)
}

type opSrcErrorTracker struct {
	lock        sync.Mutex
	opsrcErrors map[string]error
}

// NewOpSrcErrorTracker returns a new OpSrcErrorTracker.
func NewOpSrcErrorTracker() OpSrcErrorTracker {
	return &opSrcErrorTracker{
		lock:        sync.Mutex{},
		opsrcErrors: make(map[string]error),
	}
}

func (o *opSrcErrorTracker) GetKeysAndMap() ([]string, map[string]error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	// Create a map to return.
	clone := make(map[string]error)

	// Clone the original map.
	for key, value := range o.opsrcErrors {
		clone[key] = value
	}

	return o.getKeys(), clone
}

func (o *opSrcErrorTracker) Add(key string, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.opsrcErrors[key] = err
}

func (o *opSrcErrorTracker) Remove(key string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.remove(key)
}

func (o *opSrcErrorTracker) Sync(operatorHub operatorhub.OperatorHub) {
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, key := range o.getKeys() {
		if !operatorHub.IsPresentAndEnabled(key) {
			o.remove(key)
		}
	}
}

// remove is an internal non-blocking function that removes the given key
// and its value from the opsrcErrors map.
func (o *opSrcErrorTracker) remove(key string) {
	delete(o.opsrcErrors, key)
}

// getKeys is an internal non-blocking function that returns the keys present
// in the opsrcErrors map.
func (o *opSrcErrorTracker) getKeys() []string {
	keys := make([]string, 0, len(o.opsrcErrors))
	for k := range o.opsrcErrors {
		keys = append(keys, k)
	}
	return keys
}
