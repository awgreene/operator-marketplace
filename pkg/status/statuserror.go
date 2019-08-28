package status

// StatusErrorReason is a string type that represents known causes for StatusErrors.
// This type is typically used populate the Reason field in the ClusterOperatorStatusCondition
// object and should be in PascalCase.
type statusErrorReason string

const (
	// AppRegistryMetadataEmptyError captures when OperatorSource endpoint returns
	// an empty manifest list while reconciling an enabled default OperatorSource.
	AppRegistryMetadataEmptyError statusErrorReason = "AppRegistryMetadataEmptyError"

	// AppRegistryOptionsError captures when there is an error building an AppRegistry
	// Options object while reconciling an enabled default OperatorSource.
	AppRegistryOptionsError statusErrorReason = "AppRegistryOptionsError"

	// AppRegistryFactoryError captures when there is an error creating a new AppRegistry
	// client using the AppRegistryFactory while reconciling an enabled default OperatorSource.
	AppRegistryFactoryError statusErrorReason = "AppRegistryFactoryError"

	// AppRegistryListPackagesError captures when there is an error returned by the AppRegistry
	// client ListPackages function while reconciling an enabled default OperatorSource.
	AppRegistryListPackagesError statusErrorReason = "AppRegistryListPackagesError"

	// DataStoreWriteError captures when there is an error writing Operator metadata to the
	// datastore while reconciling an enabled default OperatorSource.
	DataStoreWriteError statusErrorReason = "DataStoreWriteError"

	// EnsureResourcesError captures when there is an error ensuring that all GRPC resources
	// exist.
	EnsureResourcesError statusErrorReason = "EnsureResourcesError"

	// UnknownError captures when there is an unknown error. This is a default value returned
	// when the cause for a failing enabled default OperatorSource is unknown.
	UnknownError statusErrorReason = "UnknownError"
)

// StatusError is used to provide the Degraded ClusterOperatorStatusCondition
// with a message and reason as to why the operator is degraded state.
type StatusError interface {
	// Reason returns a PascalCase string that offers a high level explanation for the StatusError.
	Reason() string

	// Error returns a string that explains the cause of the StatusError.
	Error() string
}

// statusError implements StatusError.
type statusError struct {
	// reason is a PascalCase string that offers a high level explanation for the StatusError.
	reason statusErrorReason

	// err is a string that explains the cause of the StatusError.
	err error
}

// NewStatusError returns a StatusError.
func NewStatusError(s statusErrorReason, e error) error {
	return &statusError{s, e}
}

func (s *statusError) Reason() string {
	return string(s.reason)
}

func (s *statusError) Error() string {
	if s.err == nil {
		return ""
	}
	return s.err.Error()
}

// IsStatusError returns whether or not the error is a StatusError.
func IsStatusError(err error) (StatusError, bool) {
	statusErr, ok := err.(StatusError)
	return statusErr, ok
}
