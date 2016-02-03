package models

// struct representing an item to be listed
type Kube struct {
	Name               string
	Version            string
	Services           []Service
	ReplicaControllers []ReplicaController
}

func BuildKubeList(svcList *[]Service, rcList *[]ReplicaController) map[string]Kube {
	kmap := make(map[string]Kube)
	for _, service := range *svcList {
		if service.IsKubware() {
			index := service.uuid()
			if _, ok := kmap[index]; !ok {
				// if not, create it
				kmap[index] = Kube{}
			}
			// add the service to the kube's list of services
			kube := kmap[index]
			kube.Services = append(kube.Services, service)
			kmap[index] = kube
		}
	}
	for _, replicaController := range *rcList {
		if replicaController.IsKubware() {
			index := replicaController.uuid()
			// check if kube already in map
			if _, ok := kmap[index]; !ok {
				// if not, create it
				kmap[index] = Kube{}
			}
			// add the replicaController to the kube's list of controllers
			kube := kmap[index]
			kube.ReplicaControllers = append(kube.ReplicaControllers, replicaController)
			kmap[index] = kube
		}
	}
	return kmap
}

func (k *Kube) GetNamespace() string {
	if len(k.Services) > 0 {
		return k.Services[0].GetNamespace()
	} else if len(k.ReplicaControllers) > 0 {
		return k.ReplicaControllers[0].GetNamespace()
	}
	return ""
}

func (k *Kube) GetKube() string {
	if len(k.Services) > 0 {
		return k.Services[0].GetKube()
	} else if len(k.ReplicaControllers) > 0 {
		return k.ReplicaControllers[0].GetKube()
	}
	return ""
}

func (k *Kube) GetVersion() string {
	if len(k.Services) > 0 {
		return k.Services[0].GetVersion()
	} else if len(k.ReplicaControllers) > 0 {
		return k.ReplicaControllers[0].GetVersion()
	}
	return ""
}
