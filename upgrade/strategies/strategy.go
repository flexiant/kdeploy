package upgradeStrategies

import "github.com/flexiant/kdeploy/webservice"

// UpgradeStrategy interface that upgrade strategies must implement
type UpgradeStrategy interface {
	Upgrade(namespace string, services map[string]string, controllers map[string]string) error
}

// BuildUpgradeStrategy is a factory of upgrade strategies
func BuildUpgradeStrategy(strategy string, k webservice.KubeClient) UpgradeStrategy {
	switch strategy {
	case "recreateAll":
		return RecreateAllStrategy(k)
	case "rollPreserveServices":
		return RollRcPatchSvcStrategy(k, 1)
	case "rollReplaceServices":
		return RollRcReplaceSvcStrategy(k, 1)
	default:
		return nil
	}
}
