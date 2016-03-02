package upgradeStrategies

import (
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/webservice"
)

type replicationController struct {
	readyReplicas uint
	specReplicas  uint
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

func endCondition(oldRC, newRC *replicationController, specReplicas uint) bool {
	return oldRC.readyReplicas == 0 &&
		oldRC.specReplicas == 0 &&
		newRC.readyReplicas == specReplicas &&
		newRC.specReplicas == specReplicas
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
