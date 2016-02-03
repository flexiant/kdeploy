package webservice

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flexiant/kdeploy/models"
	"github.com/flexiant/kdeploy/utils"
)

// KubeClient interface for a custom Kubernetes API client
type KubeClient interface {
	GetControllers(...string) (*models.ControllerList, error) // GetControllers gets deployed replication controllers that match the labels specified
	GetServices(...string) (*models.ServiceList, error)       // GetServices gets deployed services that match the labels specified
	CreateReplicaController(string, []byte) (string, error)
	CreateReplicaControllers([]string) error
	CreateService(string, []byte) (string, error)
	CreateServices([]string) error
	DeleteReplicationController(string, string) error
	DeleteService(string, string) error
	SetSpecReplicas(string, string, uint) error
	GetSpecReplicas(string, string) (uint, error)
	GetStatusReplicas(string, string) (uint, error)
	DeleteServices(namespace string, serviceList *models.ServiceList) error
	DeleteControllers(namespace string, controllerList *models.ControllerList) error
}

// kubeClient implements KubeClient interface
type kubeClient struct {
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
	return &kubeClient{service: rs}, nil
}

// GetServices retrieves a json representation of existing services
func (k *kubeClient) GetServices(labelSelector ...string) (*models.ServiceList, error) {
	if len(labelSelector) > 1 {
		return nil, fmt.Errorf("too many parameters")
	}
	params := map[string]string{"pretty": "true"}
	if len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector[0]
	}
	json, _, err := k.service.Get("/api/v1/services", params)
	if err != nil {
		return nil, fmt.Errorf("error getting services: %s", err)
	}

	return models.NewServicesJSON(string(json))
}

// GetServices retrieves a json representation of existing controllers
func (k *kubeClient) GetControllers(labelSelector ...string) (*models.ControllerList, error) {
	if len(labelSelector) > 1 {
		return nil, fmt.Errorf("too many parameters")
	}
	params := map[string]string{"pretty": "true"}
	if len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector[0]
	}
	json, _, err := k.service.Get("/api/v1/replicationcontrollers", params)
	if err != nil {
		return nil, fmt.Errorf("error getting replication controllers: %s", err)
	}

	return models.NewControllersJSON(string(json))
}

// CreateReplicaController creates a replication controller as specified in the json doc received as argument
func (k *kubeClient) CreateReplicaController(namespace string, rcjson []byte) (string, error) {
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
func (k *kubeClient) CreateService(namespace string, svcjson []byte) (string, error) {
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
func (k *kubeClient) DeleteService(namespace, service string) error {
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

func (k *kubeClient) DeleteReplicationController(namespace, controller string) error {
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

// GetSpecReplicas gets a replication controller's target number of replicas
func (k *kubeClient) GetSpecReplicas(namespace, controller string) (uint, error) {
	rcJSON, err := k.getController(namespace, controller)
	if err != nil {
		return 0, err
	}
	var rc struct{ Spec struct{ Replicas uint } }
	err = json.Unmarshal([]byte(rcJSON), &rc)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling '%s' : %v", rcJSON, err)
	}
	return rc.Spec.Replicas, nil
}

// GetStatusReplicas gets a replication controller's current number of replicas
func (k *kubeClient) GetStatusReplicas(namespace, controller string) (uint, error) {
	rcJSON, err := k.getController(namespace, controller)
	if err != nil {
		return 0, err
	}
	var rc struct{ Status struct{ Replicas uint } }
	err = json.Unmarshal([]byte(rcJSON), &rc)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling '%s' : %v", rcJSON, err)
	}
	return rc.Status.Replicas, nil
}

func (k *kubeClient) getController(namespace, controller string) (string, error) {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers/%s", namespace, controller)
	rcJSON, status, err := k.service.Get(path, nil)
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("wrong http status code: %v", status)
	}
	return string(rcJSON), nil
}

func (k *kubeClient) SetSpecReplicas(namespace, controller string, replicas uint) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers/%s", namespace, controller)
	body := jsonPatchSpecReplicas(replicas)
	_, status, err := k.service.Patch(path, []byte(body))
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("error creating service: wrong http status code: %v", status)
	}
	return nil
}

// Deletes a list of Services
func (k *kubeClient) DeleteServices(namespace string, serviceList *models.ServiceList) error {
	for _, service := range serviceList.Items {
		err := k.DeleteService(namespace, service.Metadata.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deletes a list of Controllers
func (k *kubeClient) DeleteControllers(namespace string, controllerList *models.ControllerList) error {
	for _, controller := range controllerList.Items {
		err := k.DeleteService(namespace, controller.Metadata.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func jsonPatchSpecReplicas(nr uint) []byte {
	var patch struct {
		Spec struct {
			Replicas uint
		}
	}
	patch.Spec.Replicas = nr
	data, _ := json.Marshal(patch)
	return data
}

func (k *kubeClient) CreateServices(svcSpecs []string) error {
	for _, spec := range svcSpecs {
		_, err := k.CreateService(os.Getenv("KDEPLOY_NAMESPACE"), []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating services: %v", err)
		}
	}
	return nil
}

func (k *kubeClient) CreateReplicaControllers(rcSpecs []string) error {
	for _, spec := range rcSpecs {
		_, err := k.CreateReplicaController(os.Getenv("KDEPLOY_NAMESPACE"), []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating replication controllers: %v", err)
		}
	}
	return nil
}
