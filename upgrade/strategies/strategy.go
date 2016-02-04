package upgradeStrategies

// UpgradeStrategy interface that upgrade strategies must implement
type UpgradeStrategy interface {
	Upgrade(namespace string, services map[string]string, controllers map[string]string) error
}
