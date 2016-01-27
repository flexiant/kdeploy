package webservice

import (
	"fmt"
	"net/url"

	"github.com/flexiant/kdeploy/utils"
)

// KubeClient interface for a custom Kubernetes API client
type KubeClient interface {
	GetControllers(map[string]string) (string, error) // GetControllers gets deployed replication controllers that match the labels specified
	GetServices(map[string]string) (string, error)    // GetServices gets deployed services that match the labels specified
	CreateReplicaController(string, []byte) (string, error)
	CreateService(string, []byte) (string, error)
	DeleteReplicationController(string, string) error
	DeleteService(string, string) error
}

// kubeClientImpl implements KubeClient interface
type kubeClientImpl struct {
	service *RestService
}

// NewKubeClient builds a KubeClient object
func NewKubeClient() (KubeClient, error) {
	cfg, err := utils.CachedConfig()
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
func (k *kubeClientImpl) GetServices(labelSelector map[string]string) (string, error) {
	filter := url.Values{}
	for k, v := range labelSelector {
		filter.Add(k, v)
	}
	params := map[string]string{
		"pretty":        "true",
		"labelSelector": filter.Encode(),
	}
	json, _, err := k.service.Get("/api/v1/services", params)
	if err != nil {
		return "", fmt.Errorf("error getting services: %s", err)
	}
	return string(json), nil
}

// GetServices retrieves a json representation of existing controllers
func (k *kubeClientImpl) GetControllers(labelSelector map[string]string) (string, error) {
	filter := url.Values{}
	for k, v := range labelSelector {
		filter.Add(k, v)
	}
	params := map[string]string{
		"pretty":        "true",
		"labelSelector": filter.Encode(),
	}
	json, _, err := k.service.Get("/api/v1/replicationcontrollers", params)
	if err != nil {
		return "", fmt.Errorf("error getting replication controllers: %s", err)
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

// DeleteService deletes a service
func (k *kubeClientImpl) DeleteService(namespace, service string) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/services/%s", namespace, service)
	_, status, err := k.service.Delete(path)
	if err != nil {
		return fmt.Errorf("error deleting service: %s", err)
	}
	if status != 200 {
		return fmt.Errorf("error deleting service: wrong http status code: %v", status)
	}
	return nil
}

// DeleteService deletes a service
func (k *kubeClientImpl) DeleteReplicationController(namespace, controller string) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers/%s", namespace, controller)
	_, status, err := k.service.Delete(path)
	if err != nil {
		return fmt.Errorf("error deleting controller: %s", err)
	}
	if status != 200 {
		return fmt.Errorf("error deleting controller: wrong http status code: %v", status)
	}
	return nil
}
