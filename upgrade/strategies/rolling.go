package upgradeStrategies

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flexiant/digger"
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
	for rcName, rcJSON := range controllers {
		// create new rc with new name (e.g. name-next) and 0 target replicas (why not rename old? -> repeatability)
		tempName := fmt.Sprintf("%s-next", rcName)
		_, err := s.createRCAsRollingTarget(namespace, tempName, rcJSON)
		if err != nil {
			return err
		}
		// read desired replicas
		targetReplicas, err := extractSpecReplicas(rcJSON)
		if err != nil {
			return err
		}
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
	var rc map[string]interface{}
	err := json.Unmarshal([]byte(rcJSON), &rc)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal rc: %v", err)
	}
	// set kube version as label in pod template, and in selector
	err = setKubeVersionOnRC(rc)
	if err != nil {
		return "", fmt.Errorf("could not set version label: %v", err)
	}
	// set zero replicas
	err = setZeroReplicasOnRC(rc)
	if err != nil {
		return "", fmt.Errorf("could not zero replicas: %v", err)
	}
	// rename it
	err = renameRC(rc)
	if err != nil {
		return "", fmt.Errorf("could not rename rc: %v", err)
	}
	// create it
	newJSON, err := json.Marshal(rc)
	if err != nil {
		return "", fmt.Errorf("could not marshal modified rc: %v", err)
	}
	return s.kubeClient.CreateReplicaController(namespace, newJSON)
}

func renameRC(rc map[string]interface{}) error {
	return fmt.Errorf("Not Implemented Yet")
}

func setKubeVersionOnRC(rc map[string]interface{}) error {
	kv, err := extractKubeVersion(rc)
	if err != nil {
		return err
	}
	// set at pod template
	path := []string{"spec", "template", "spec", "metadata", "labels"}
	m := rc
	for _, s := range path {
		if m[s] == nil {
			m[s] = map[string]interface{}{}
		}
		m = m[s].(map[string]interface{})
	}
	m["kubeware"] = kv
	// set at label selector
	rcspec := rc["spec"].(map[string]interface{})
	ls, found := rcspec["selector"].(string)
	if !found || len(ls) == 0 {
		rcspec["selector"] = fmt.Sprintf("kubeware=%s", kv)
	} else {
		rcspec["selector"] = fmt.Sprintf("%s,kubeware=%s", ls, kv)
	}

	return nil
}

func setZeroReplicasOnRC(rc map[string]interface{}) error {
	rcspec := rc["spec"].(map[string]interface{})
	rcspec["replicas"] = 0
	return nil
}

func extractKubeVersion(rc map[string]interface{}) (string, error) {
	d, err := digger.NewMapDigger(rc)
	if err != nil {
		return "", err
	}
	k, err := d.GetString("metadata/labels/kubeware")
	if err != nil {
		return "", err
	}
	v, err := d.GetString("metadata/labels/kubeware-version")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", k, v), nil
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
