package upgradeStrategies

import (
	"github.com/flexiant/kdeploy/delete/strategies"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

type recreateAll struct {
	kubeClient webservice.KubeClient
}

// RecreateAllStrategy builds a new instance of 'waitZeroReplicas' deletion strategy
func RecreateAllStrategy(k webservice.KubeClient) UpgradeStrategy {
	return &recreateAll{k}
}

func (rst *recreateAll) Upgrade(namespace string, services map[string]string, controllers map[string]string) error {
	svcNames := utils.Keys(services)
	rcNames := utils.Keys(controllers)
	// delete all
	delStrategy := deletionStrategies.WaitZeroReplicasDeletionStrategy(rst.kubeClient)
	err := delStrategy.Delete(namespace, svcNames, rcNames)
	if err != nil {
		return err
	}
	// recreate all
	err = rst.kubeClient.CreateServices(utils.Values(services))
	if err != nil {
		return err
	}
	err = rst.kubeClient.CreateReplicaControllers(utils.Values(controllers))
	if err != nil {
		return err
	}
	// done !
	return nil
}
