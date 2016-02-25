package models

import (
	"encoding/json"
	"fmt"
	"strings"

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
		Selector map[string]interface{}
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

func (rc *ReplicaController) GetSelectorAsString() string {
	parts := []string{}
	for k, v := range rc.Spec.Selector {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
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
