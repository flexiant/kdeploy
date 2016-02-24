package models

import (
	"encoding/json"
	"fmt"

	"github.com/flexiant/kdeploy/utils"
)

type Service struct {
	Metadata struct {
		CreationTimestamp string
		Name              string
		Labels            map[string]string
		Namespace         string
		ResourceVersion   string
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

func (svc *Service) GetFQDN() string {
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		return svc.Status.LoadBalancer.Ingress[0].Hostname
	}
	return ""
}
func (svc *Service) IsKubware() bool {
	if svc.GetKube() != "" && svc.GetVersion() != "" {
		return true
	}
	return false
}

func (svc *Service) uuid() string {
	return utils.GetMD5Hash(fmt.Sprintf("%s%s%s", svc.Metadata.Namespace, svc.Metadata.Labels["kubeware"], svc.Metadata.Labels["kubeware-version"]))
}

func (svc *Service) GetNamespace() string {
	return svc.Metadata.Namespace
}

func (svc *Service) GetVersion() string {
	return svc.Metadata.Labels["kubeware-version"]
}

func (svc *Service) GetKube() string {
	return svc.Metadata.Labels["kubeware"]
}

func (svc *Service) GetName() string {
	return svc.Metadata.Name
}

func (svc *Service) GetInternalIp() string {
	return svc.Spec.ClusterIP
}

func NewServicesJSON(jsonStr string) (*[]Service, error) {
	type ServiceList struct {
		Items []Service
	}
	var sl ServiceList
	err := json.Unmarshal([]byte(jsonStr), &sl)
	if err != nil {
		return nil, err
	}
	return &sl.Items, nil
}
