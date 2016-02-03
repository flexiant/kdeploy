package deletionStrategies

// DeletionStrategy is an abstract interface to support several possible
// deletion strategies; e.g. preserving pods or not
type DeletionStrategy interface {
	Delete(namespace string, services []string, replicationControllers []string) error
}
