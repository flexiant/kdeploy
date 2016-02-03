package deletionStrategies

// DeletionStrategy is an abstract interface to support several possible
// deletion strategies; e.g. preserving pods or not
type DeletionStrategy interface {
	DeleteService(namespace string, name string) error
	DeleteReplicationController(namespace string, name string) error
}
