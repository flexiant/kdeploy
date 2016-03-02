package upgradeStrategies

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// implements a rolling upgrade strategy
type rollPatchStrategy struct {
	maxReplicasExcess uint // Total amount of old + new replicas must not exceed new target number of replicas plus this number
	kubeClient        webservice.KubeClient
}

// RollRcPatchSvcStrategy will roll-update replication controllers and patch services without recreating them
func RollRcPatchSvcStrategy(k webservice.KubeClient, maxReplicasExcess uint) UpgradeStrategy {
	return &rollPatchStrategy{
		maxReplicasExcess: maxReplicasExcess,
		kubeClient:        k,
	}
}

func (s *rollPatchStrategy) Upgrade(namespace string, services, controllers map[string]string) error {
	log.Debugf("Using rolling upgrade strategy")

	// for each rc
	for rcName, rcJSON := range controllers {

		tempName := fmt.Sprintf("%s-next", rcName)
		log.Debugf("Creating temporal RC '%s'", tempName)
		_, err := s.createRCAsRollingTarget(namespace, tempName, rcJSON)
		if err != nil {
			log.Debugf("error creating rolling target: %v", err)
			return err
		}
		// read desired replicas
		targetReplicas, err := extractSpecReplicas(rcJSON)
		if err != nil {
			return err
		}
		log.Debugf("Want '%v' replicas", targetReplicas)
		// roll them
		err = s.rollReplicationController(namespace, rcName, tempName, targetReplicas)
		if err != nil {
			return fmt.Errorf("error rolling out %s : %v", rcName, err)
		}
		// replace "name" with new rc
		s.kubeClient.ReplaceReplicationController(namespace, rcName, rcJSON)
		// delete "name-next"
		s.kubeClient.DeleteReplicationController(namespace, tempName)
	}
	// for each service
	for svcName, svcJSON := range services {
		err := s.upgradeService(namespace, svcName, svcJSON)
		if err != nil {
			log.Debugf("Error upgrading service: %v", err)
			return err
		}
	}
	return nil
}

// here we should create the new RC considering:
// - overwriting its name with the custom (temporal) one
// - setting replicas to zero so they dont get created all at once
// - setting a specific label in the pod template and the selector so that the pods
//   dont get mixed with the previous version ones
func (s *rollPatchStrategy) createRCAsRollingTarget(namespace, name, rcJSON string) (string, error) {
	// parse json object
	log.Debugf("parsing json for new RC '%s'", name)
	var rc map[string]interface{}
	err := json.Unmarshal([]byte(rcJSON), &rc)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal rc: %v", err)
	}
	// set zero replicas
	log.Debugf("setting zero replicas for new RC '%s'", name)
	err = setZeroReplicasOnRC(rc)
	if err != nil {
		return "", fmt.Errorf("could not zero replicas: %v", err)
	}
	// rename it
	log.Debugf("renaming new RC '%s'", name)
	renameRC(rc, name)
	// create it
	log.Debugf("creating new RC '%s'", name)
	newJSON, err := json.Marshal(rc)
	if err != nil {
		log.Debugf("could not marshal modified rc: %v", err)
		return "", fmt.Errorf("could not marshal modified rc: %v", err)
	}
	return s.kubeClient.CreateReplicaController(namespace, newJSON)
}

func (s *rollPatchStrategy) upgradeService(namespace, svcName, svcJSON string) error {
	// patch if exists, create if not
	deployed, err := s.kubeClient.IsServiceDeployed(namespace, svcName)
	if err != nil {
		return err
	}
	if deployed {
		log.Debugf("Patch service '%s'", svcName)
		err = s.kubeClient.PatchService(namespace, svcName, onlyMetadata(svcJSON))
		if err != nil {
			return err
		}
	} else {
		log.Debugf("Creating service '%s' since it wasnt deployed previously", svcName)
		_, err = s.kubeClient.CreateService(namespace, []byte(svcJSON))
		if err != nil {
			return err
		}
	}
	return nil
}

func onlyMetadata(svcJSON string) string {
	var svc map[string]interface{}
	err := json.Unmarshal([]byte(svcJSON), &svc)
	utils.CheckError(err)

	filtered := map[string]interface{}{
		"apiVersion": svc["apiVersion"],
		"kind":       svc["kind"],
		"metadata":   svc["metadata"],
	}

	filteredJSON, err := json.Marshal(filtered)
	utils.CheckError(err)
	return string(filteredJSON)
}

func (s *rollPatchStrategy) rollReplicationController(ns, oldRCid, newRCid string, targetReplicas uint) error {
	log.Debugf("Rolling out RC '%s' to '%s'", oldRCid, newRCid)
	for ; ; time.Sleep(1 * time.Second) {
		// build RC objects
		oldRC, err := buildRCObject(s.kubeClient, ns, oldRCid)
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
		newRC, err := buildRCObject(s.kubeClient, ns, newRCid)
		if err != nil {
			log.Debugf("error: %v", err)
			return err
		}
		//
		if endCondition(oldRC, newRC, targetReplicas) {
			return nil
		}
		// newRC all pods ready? (spec == status)
		if newRC.readyReplicas == newRC.specReplicas {
			// newRC reached target number of replicas?
			if newRC.readyReplicas < targetReplicas {
				// observe MaxReplicasExcess
				totalCurrentReplicas := newRC.readyReplicas + oldRC.readyReplicas
				totalAllowedReplicas := targetReplicas + s.maxReplicasExcess
				if totalCurrentReplicas < totalAllowedReplicas {
					s.kubeClient.SetSpecReplicas(ns, newRCid, newRC.specReplicas+1)
					log.Debugf("Set '%s' to '%v' replicas", newRCid, newRC.specReplicas+1)
				}
			}
		}
		// status meets spec ?
		if oldRC.readyReplicas == oldRC.specReplicas {
			if oldRC.readyReplicas > 0 {
				// can decrease?
				totalCurrentReplicas := newRC.readyReplicas + oldRC.readyReplicas
				if totalCurrentReplicas > targetReplicas {
					s.kubeClient.SetSpecReplicas(ns, oldRCid, oldRC.specReplicas-1)
					log.Debugf("Set '%s' to '%v' replicas", oldRCid, oldRC.specReplicas-1)
				}
			}
		}
	}
}
