package upgradeStrategies

import (
	"time"

	"github.com/flexiant/kdeploy/webservice"
)

// implements a rolling upgrade strategy
type rollingStrategy struct {
	maxReplicasExcess uint // Total amount of old + new replicas must not exceed new target number of replicas plus this number
	kubeClient        webservice.KubeClient
}

type replicationController struct {
	statusReplicas uint
	specReplicas   uint
}

// RollingUpgradeStrategy builds rolling upgrade strategy objects
func RollingUpgradeStrategy(k webservice.KubeClient, maxReplicasExcess uint) UpgradeStrategy {
	return &rollingStrategy{
		maxReplicasExcess: maxReplicasExcess,
		kubeClient:        k,
	}
}

func (s *rollingStrategy) Upgrade(namespace string, services, controllers map[string]string) error {
	// for each service
	for svcName, svcJSON := range services {
		err := s.upgradeService(namespace, svcName, svcJSON)
		if err != nil {
			return err
		}
	}
	// for each rc
	//    create new rc with new name (e.g. name-next) and 0 target replicas (why not rename old? -> repeatability)
	//    roll "name" and "name-next"
	//    replace "name" with new rc
	// 		delete "name-next"
	return nil
}

func (s *rollingStrategy) upgradeService(namespace, svcName, svcJSON string) error {
	// replace if exists, create if not
	deployed, err := s.kubeClient.IsServiceDeployed(namespace, svcName)
	if err != nil {
		return err
	}
	if deployed {
		err = s.kubeClient.ReplaceService(namespace, svcName, svcJSON)
		if err != nil {
			return err
		}
	} else {
		_, err = s.kubeClient.CreateService(namespace, []byte(svcJSON))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *rollingStrategy) rollReplicationController(ns, oldRCid, newRCid string, targetReplicas uint) error {
	for ; ; time.Sleep(1 * time.Second) {
		// build RC objects
		oldRC, err := buildRCObject(s.kubeClient, ns, oldRCid)
		if err != nil {
			return err
		}
		newRC, err := buildRCObject(s.kubeClient, ns, newRCid)
		if err != nil {
			return err
		}
		//
		if endCondition(oldRC, newRC, targetReplicas) {
			return nil
		}
		// newRC all pods ready? (spec == status)
		if newRC.statusReplicas == newRC.specReplicas {
			// newRC reached target number of replicas?
			if newRC.statusReplicas < targetReplicas {
				// observe MaxReplicasExcess
				totalCurrentReplicas := newRC.statusReplicas + oldRC.statusReplicas
				totalAllowedReplicas := targetReplicas + s.maxReplicasExcess
				if totalCurrentReplicas < totalAllowedReplicas {
					s.kubeClient.SetSpecReplicas(ns, newRCid, newRC.specReplicas+1)
				}
			}
		}
		// status meets spec ?
		if oldRC.statusReplicas == oldRC.specReplicas {
			if oldRC.statusReplicas > 0 {
				// can decrease?
				totalCurrentReplicas := newRC.statusReplicas + oldRC.statusReplicas
				if totalCurrentReplicas > targetReplicas {
					s.kubeClient.SetSpecReplicas(ns, oldRCid, oldRC.specReplicas-1)
				}
			}
		}
	}
}

func endCondition(oldRC, newRC *replicationController, specReplicas uint) bool {
	return oldRC.statusReplicas == 0 &&
		oldRC.specReplicas == 0 &&
		newRC.statusReplicas == specReplicas &&
		newRC.specReplicas == specReplicas
}

func buildRCObject(kube webservice.KubeClient, ns, rcname string) (*replicationController, error) {
	statusReplicas, err := kube.GetStatusReplicas(ns, rcname)
	if err != nil {
		return nil, err
	}
	specReplicas, err := kube.GetSpecReplicas(ns, rcname)
	if err != nil {
		return nil, err
	}
	rc := replicationController{
		statusReplicas: statusReplicas,
		specReplicas:   specReplicas,
	}
	return &rc, nil
}
