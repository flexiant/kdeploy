package models

import (
	"encoding/json"
	"fmt"

	"github.com/flexiant/kdeploy/utils"
)

type ReplicaController struct {
	Metadata struct {
		Name      string
		Labels    map[string]string
		Namespace string
	}
	Spec struct {
		Replicas int
	}
	Status struct {
		Replicas int
	}
}

func (rc *ReplicaController) IsKubware() bool {
	if rc.GetKube() != "" && rc.GetVersion() != "" {
		return true
	}
	return false
}

func (rc *ReplicaController) uuid() string {
	return utils.GetMD5Hash(fmt.Sprintf("%s%s%s", rc.Metadata.Namespace, rc.Metadata.Labels["kubeware"], rc.Metadata.Labels["kubeware-version"]))
}

func (rc *ReplicaController) GetNamespace() string {
	return rc.Metadata.Namespace
}

func (rc *ReplicaController) GetVersion() string {
	return rc.Metadata.Labels["kubeware-version"]
}

func (rc *ReplicaController) GetKube() string {
	return rc.Metadata.Labels["kubeware"]
}

func (rc *ReplicaController) GetName() string {
	return rc.Metadata.Name
}

func (rc *ReplicaController) GetReplicas() int {
	return rc.Status.Replicas
}

func (rc *ReplicaController) GetUpStats() int {
	return (rc.Status.Replicas / rc.Spec.Replicas) * 100
}

func NewControllersJSON(jsonStr string) (*[]ReplicaController, error) {
	// log.Debugf("New Controller: %s", jsonStr)
	type ReplicaControllerList struct {
		Items []ReplicaController
	}
	var rl ReplicaControllerList
	err := json.Unmarshal([]byte(jsonStr), &rl)
	if err != nil {
		return nil, err
	}
	return &rl.Items, nil
}
