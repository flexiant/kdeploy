package webservice

import (
	"fmt"

	"github.com/flexiant/kdeploy/config"
)

// KubeClient interface for a custom Kubernetes API client
type KubeClient interface {
	GetServices() (string, error)
	CreateReplicaController(string, []byte) (string, error)
	CreateService(string, []byte) (string, error)
}

// kubeClientImpl implements KubeClient interface
type kubeClientImpl struct {
	service *RestService
}

// NewKubeClient builds a KubeClient object
func NewKubeClient() (KubeClient, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}
	rs, err := NewRestService(*cfg)
	if err != nil {
		return nil, err
	}
	return &kubeClientImpl{service: rs}, nil
}

// GetServices retrieves a json representation of existing services
func (k *kubeClientImpl) GetServices() (string, error) {
	json, _, err := k.service.Get("/api/v1/services?pretty=true")
	if err != nil {
		return "", fmt.Errorf("error getting services: %s", err)
	}
	return string(json), nil
}

// CreateReplicaController creates a replication controller as specified in the json doc received as argument
func (k *kubeClientImpl) CreateReplicaController(namespace string, rcjson []byte) (string, error) {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers", namespace)
	json, status, err := k.service.Post(path, []byte(rcjson))
	if err != nil {
		return "", fmt.Errorf("error creating replication controller: %s", err)
	}
	if status != 200 && status != 201 {
		return "", fmt.Errorf("error creating service: wrong http status code: %v", status)
	}
	return string(json), nil
}

// CreateService creates a service as specified in the json doc received as argument
func (k *kubeClientImpl) CreateService(namespace string, svcjson []byte) (string, error) {
	path := fmt.Sprintf("api/v1/namespaces/%s/services", namespace)
	json, status, err := k.service.Post(path, []byte(svcjson))
	if err != nil {
		return "", fmt.Errorf("error creating service: %s", err)
	}
	if status != 200 && status != 201 {
		return "", fmt.Errorf("error creating service: wrong http status code: %v", status)
	}
	return string(json), nil
}
