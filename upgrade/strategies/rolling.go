package upgradeStrategies

import (
	"time"

	"github.com/flexiant/kdeploy/webservice"
)

// implements a rolling upgrade strategy
type rolling struct {
	TargetReplicas    uint // Total amount of new replicas desired
	MaxReplicasExcess uint // Total amount of old + new replicas must not exceed new target number of replicas plus this number
}

type replicationController struct {
	CurrentReplicas uint
	TargetReplicas  uint
}

// UpgradeReplicationController using a rolling upgrade strategy
func (s *rolling) UpgradeReplicationController(kube webservice.KubeClient, ns, oldRCid, newRCid string) error {
	for ; ; time.Sleep(1 * time.Second) {
		// build RC objects
		oldRC, err := buildRCObject(kube, ns, oldRCid)
		if err != nil {
			return err
		}
		newRC, err := buildRCObject(kube, ns, newRCid)
		if err != nil {
			return err
		}
		//
		if endCondition(oldRC, newRC, s.TargetReplicas) {
			return nil
		}
		// newRC all pods ready? (spec == status)
		if newRC.CurrentReplicas == newRC.TargetReplicas {
			// newRC reached target number of replicas?
			if newRC.CurrentReplicas < s.TargetReplicas {
				// observe MaxReplicasExcess
				totalCurrentReplicas := newRC.CurrentReplicas + oldRC.CurrentReplicas
				totalAllowedReplicas := s.TargetReplicas + s.MaxReplicasExcess
				if totalCurrentReplicas < totalAllowedReplicas {
					kube.SetSpecReplicas(ns, newRCid, newRC.TargetReplicas+1)
				}
			}
		}

		// status meets spec ?
		if oldRC.CurrentReplicas == oldRC.TargetReplicas {
			if oldRC.CurrentReplicas > 0 {
				// can decrease?
				totalCurrentReplicas := newRC.CurrentReplicas + oldRC.CurrentReplicas
				if totalCurrentReplicas > s.TargetReplicas {
					kube.SetSpecReplicas(ns, oldRCid, oldRC.TargetReplicas-1)
				}
			}
		}
	}
}

func endCondition(oldRC, newRC *replicationController, targetReplicas uint) bool {
	return oldRC.CurrentReplicas == 0 &&
		oldRC.TargetReplicas == 0 &&
		newRC.CurrentReplicas == targetReplicas &&
		newRC.TargetReplicas == targetReplicas
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
		CurrentReplicas: statusReplicas,
		TargetReplicas:  specReplicas,
	}
	return &rc, nil
}
