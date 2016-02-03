package models

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
)

// struct representing an item to be listed
type Kube struct {
	Services    []map[string]interface{}
	Controllers []map[string]interface{}
}

func (k *Kube) GetNamespace() string {
	if len(k.Services) > 0 {
		return k.Services[0]["Namespace"].(string)
	} else if len(k.Controllers) > 0 {
		return k.Controllers[0]["Namespace"].(string)
	}
	return ""
}

func (k *Kube) GetVersion() string {
	if len(k.Services) > 0 {
		return k.Services[0]["Version"].(string)
	} else if len(k.Controllers) > 0 {
		return k.Controllers[0]["Version"].(string)
	}
	return ""
}

type ControllerList struct {
	Items []controllerItem
}

type controllerItem struct {
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

// struct for unmarshalling json representations of services
type ServiceList struct {
	Items []serviceItem
}

type serviceItem struct {
	Metadata struct {
		CreationTimestamp string
		Name              string
		Labels            map[string]string
		Namespace         string
	}
	Spec struct {
		ClusterIP string
	}
	Status struct {
		LoadBalancer struct {
			Ingress []struct {
				Hostname string
			}
		}
	}
}

func BuildKubeList(svcList *ServiceList, rcList *ControllerList) map[string]Kube {
	kmap := map[string]Kube{}
	for _, service := range svcList.Items {
		if kubeName, ok := service.Metadata.Labels["kubeware"]; ok {
			// check if kube already in map
			if _, ok := kmap[kubeName]; !ok {
				// if not, create it
				kmap[kubeName] = Kube{}
			}
			// add the service to the kube's list of services
			kube := kmap[kubeName]
			kube.Services = append(kube.Services, buildServiceRecord(service))
			kmap[kubeName] = kube
		}
	}
	for _, controller := range rcList.Items {
		if kubeName, ok := controller.Metadata.Labels["kubeware"]; ok {
			// check if kube already in map
			if _, ok := kmap[kubeName]; !ok {
				// if not, create it
				kmap[kubeName] = Kube{}
			}
			// add the controller to the kube's list of controllers
			kube := kmap[kubeName]
			kube.Controllers = append(kube.Controllers, buildControllerRecord(controller))
			kmap[kubeName] = kube
		}
	}
	return kmap
}

func buildServiceRecord(service serviceItem) map[string]interface{} {
	sr := map[string]interface{}{}
	sr["Name"] = service.Metadata.Name
	sr["Namespace"] = service.Metadata.Namespace
	sr["Version"] = service.Metadata.Labels["kubeware-version"]
	sr["CreationDate"] = service.Metadata.CreationTimestamp
	sr["ClusterIP"] = service.Spec.ClusterIP
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		sr["ExternalFQDN"] = service.Status.LoadBalancer.Ingress[0].Hostname
	}
	return sr
}

func buildControllerRecord(controller controllerItem) map[string]interface{} {
	sr := map[string]interface{}{}
	sr["Name"] = controller.Metadata.Name
	sr["Replicas"] = controller.Status.Replicas
	sr["Namespace"] = controller.Metadata.Namespace
	sr["Version"] = controller.Metadata.Labels["kubeware-version"]
	sr["Up"] = (controller.Status.Replicas / controller.Spec.Replicas) * 100
	return sr
}

func NewControllersJSON(jsonStr string) (*ControllerList, error) {
	log.Debugf("New Controller: %s", jsonStr)
	var rl ControllerList
	err := json.Unmarshal([]byte(jsonStr), &rl)
	if err != nil {
		return nil, err
	}
	return &rl, nil
}

func NewServicesJSON(jsonStr string) (*ServiceList, error) {
	log.Debugf("New Service: %s", jsonStr)
	var sl ServiceList
	err := json.Unmarshal([]byte(jsonStr), &sl)
	if err != nil {
		return nil, err
	}
	return &sl, nil
}
