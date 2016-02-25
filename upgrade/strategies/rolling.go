package upgradeStrategies

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/webservice"
)

// implements a rolling upgrade strategy
type rollingStrategy struct {
	maxReplicasExcess uint // Total amount of old + new replicas must not exceed new target number of replicas plus this number
	kubeClient        webservice.KubeClient
}

type replicationController struct {
	readyReplicas uint
	specReplicas  uint
}

// RollingUpgradeStrategy builds rolling upgrade strategy objects
func RollingUpgradeStrategy(k webservice.KubeClient, maxReplicasExcess uint) UpgradeStrategy {
	return &rollingStrategy{
		maxReplicasExcess: maxReplicasExcess,
		kubeClient:        k,
	}
}

func (s *rollingStrategy) Upgrade(namespace string, services, controllers map[string]string) error {
	log.Debugf("Using rolling upgrade strategy")

	// for each rc
	for rcName, rcJSON := range controllers {
		// create new rc with new name (e.g. name-next) and 0 target replicas (why not rename old? -> repeatability)
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

func extractSpecReplicas(rcJSON string) (uint, error) {
	d, err := digger.NewJSONDigger([]byte(rcJSON))
	if err != nil {
		return 0, err
	}
	n, err := d.GetNumber("spec/replicas")
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// here we should create the new RC considering:
// - overwriting its name with the custom (temporal) one
// - setting replicas to zero so they dont get created all at once
// - setting a specific label in the pod template and the selector so that the pods
//   dont get mixed with the previous version ones
func (s *rollingStrategy) createRCAsRollingTarget(namespace, name, rcJSON string) (string, error) {
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

func renameRC(rc map[string]interface{}, name string) {
	m := rc["metadata"].(map[string]interface{})
	m["name"] = name
	rc["metadata"] = m
}

func setZeroReplicasOnRC(rc map[string]interface{}) error {
	rcspec := rc["spec"].(map[string]interface{})
	rcspec["replicas"] = 0
	return nil
}

func (s *rollingStrategy) upgradeService(namespace, svcName, svcJSON string) error {
	// replace if exists, create if not
	deployed, err := s.kubeClient.IsServiceDeployed(namespace, svcName)
	if err != nil {
		return err
	}
	if deployed {
		log.Debugf("Replacing service '%s'", svcName)
		err = s.kubeClient.ReplaceService(namespace, svcName, svcJSON)
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

func (s *rollingStrategy) rollReplicationController(ns, oldRCid, newRCid string, targetReplicas uint) error {
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

func endCondition(oldRC, newRC *replicationController, specReplicas uint) bool {
	return oldRC.readyReplicas == 0 &&
		oldRC.specReplicas == 0 &&
		newRC.readyReplicas == specReplicas &&
		newRC.specReplicas == specReplicas
}

func buildRCObject(kube webservice.KubeClient, ns, rcname string) (*replicationController, error) {
	readyReplicas, err := countReadyReplicas(kube, ns, rcname)
	if err != nil {
		return nil, err
	}
	specReplicas, err := kube.GetSpecReplicas(ns, rcname)
	if err != nil {
		return nil, err
	}
	rc := replicationController{
		readyReplicas: readyReplicas,
		specReplicas:  specReplicas,
	}
	return &rc, nil
}

func countReadyReplicas(kube webservice.KubeClient, ns, rcname string) (uint, error) {
	pods, err := kube.GetPodsForController(ns, rcname)
	if err != nil {
		return 0, err
	}
	var count uint
	for _, p := range *pods {
		for _, c := range p.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				count++
				break
			}
		}
	}
	return count, nil
}
