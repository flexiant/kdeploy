package webservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/flexiant/kdeploy/models"
	"github.com/flexiant/kdeploy/utils"
)

// ErrNotFound indicates the object that we are looking for was not found
var ErrNotFound = errors.New("Object not found")

// KubeClient interface for a custom Kubernetes API client
type KubeClient interface {
	FindDeployedKubewareVersion(namespace, kubeName string) (string, error)
	GetControllers(labelSelector ...string) (*[]models.ReplicaController, error)                               // GetControllers gets deployed replication controllers that match the labels specified
	GetControllersForNamespace(namespace string, labelSelector ...string) (*[]models.ReplicaController, error) // GetControllers gets deployed replication controllers that match the labels specified
	GetServices(labelSelector ...string) (*[]models.Service, error)                                            // GetServices gets deployed services that match the labels specified
	GetServicesForNamespace(namespace string, labelSelector ...string) (*[]models.Service, error)              // GetServices gets deployed services that match the labels specified
	CreateReplicaController(namespace string, spec []byte) (string, error)
	CreateReplicaControllers(rcSpecs []string) error
	CreateService(namespace string, spec []byte) (string, error)
	CreateServices(svcSpecs []string) error
	DeleteReplicationController(namespace, rcName string) error
	DeleteService(namespace, svcName string) error
	SetSpecReplicas(namespace, rcName string, nreplicas uint) error
	GetSpecReplicas(namespace, rcName string) (uint, error)
	GetStatusReplicas(namespace, rcName string) (uint, error)
	IsServiceDeployed(namespace, svcName string) (bool, error)
	ReplaceReplicationController(namespace, rcName, rcJSON string) error
	ReplaceService(namespace, svcName, svcJSON string) error
	GetPodsForNamespace(namespace, labelSelector string) (*[]models.Pod, error)
	GetPodsForController(namespace, rcName string) (*[]models.Pod, error)
	PatchService(namespace, svcName, svcJSON string) error
}

// kubeClient implements KubeClient interface
type kubeClient struct {
	service *RestService
}

// NewKubeClient builds a KubeClient object
func NewKubeClient() (KubeClient, error) {
	//cfg, err := utils.CachedConfig()
	cfg, err := utils.GetConfig()
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
func (k *kubeClient) GetServices(labelSelector ...string) (*[]models.Service, error) {
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

	srv, err := models.NewServicesJSON(string(json))
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %s %s", err, json)
	}
	return srv, nil
}

func (k *kubeClient) GetServicesForNamespace(namespace string, labelSelector ...string) (*[]models.Service, error) {
	if len(labelSelector) > 1 {
		return nil, fmt.Errorf("too many parameters")
	}
	params := map[string]string{"pretty": "true"}
	if len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector[0]
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/services", namespace)
	json, _, err := k.service.Get(path, params)
	if err != nil {
		return nil, fmt.Errorf("error getting services: %s", err)
	}

	return models.NewServicesJSON(string(json))
}

// GetServices retrieves a json representation of existing controllers
func (k *kubeClient) GetControllers(labelSelector ...string) (*[]models.ReplicaController, error) {
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

func (k *kubeClient) GetControllersForNamespace(namespace string, labelSelector ...string) (*[]models.ReplicaController, error) {
	if len(labelSelector) > 1 {
		return nil, fmt.Errorf("too many parameters")
	}
	params := map[string]string{"pretty": "true"}
	if len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector[0]
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/replicationcontrollers", namespace)
	json, _, err := k.service.Get(path, params)
	if err != nil {
		return nil, fmt.Errorf("error getting replication controllers: %s", err)
	}

	return models.NewControllersJSON(string(json))
}

func (k *kubeClient) GetPodsForNamespace(namespace string, labelSelector string) (*[]models.Pod, error) {
	params := map[string]string{
		"labelSelector": labelSelector,
		"pretty":        "true",
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace)
	json, _, err := k.service.Get(path, params)
	if err != nil {
		return nil, fmt.Errorf("error getting pods: %s", err)
	}
	return models.NewPodsJSON(string(json))
}

func (k *kubeClient) GetPodsForController(namespace string, rcName string) (*[]models.Pod, error) {
	rcJSON, err := k.getController(namespace, rcName)
	if err != nil {
		return nil, err
	}
	var rc models.ReplicaController
	err = json.Unmarshal([]byte(rcJSON), &rc)
	if err != nil {
		return nil, err
	}
	return k.GetPodsForNamespace(rc.GetNamespace(), rc.GetSelectorAsString())
}

// CreateReplicaController creates a replication controller as specified in the json doc received as argument
func (k *kubeClient) CreateReplicaController(namespace string, rcjson []byte) (string, error) {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers", namespace)
	json, status, err := k.service.Post(path, []byte(rcjson))
	if err != nil {
		return "", fmt.Errorf("error creating replication controller: %s", err)
	}
	if status != 200 && status != 201 {
		return "", fmt.Errorf("error creating service: wrong http status code: %v (body: %s)", status, json)
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
		return fmt.Errorf("wrong http status code: %v", status)
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
	if status == 404 {
		return "", ErrNotFound
	}
	if status != 200 {
		return "", fmt.Errorf("wrong http status code: %v", status)
	}
	return string(rcJSON), nil
}

func (k *kubeClient) getService(namespace, svcName string) (string, error) {
	path := fmt.Sprintf("api/v1/namespaces/%s/services/%s", namespace, svcName)
	rcJSON, status, err := k.service.Get(path, nil)
	if err != nil {
		return "", err
	}
	if status == 404 {
		return "", ErrNotFound
	}
	if status != 200 {
		return "", fmt.Errorf("wrong http status code: %v (body: %s)", status, rcJSON)
	}
	return string(rcJSON), nil
}

func (k *kubeClient) SetSpecReplicas(namespace, controller string, replicas uint) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers/%s", namespace, controller)
	body := jsonPatchSpecReplicas(replicas)
	resp, status, err := k.service.Patch(path, []byte(body))
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("wrong http status code: %v (%s)", status, resp)
	}
	return nil
}

func jsonPatchSpecReplicas(nr uint) []byte {
	var patch struct {
		Spec struct {
			Replicas uint `json:"replicas"`
		} `json:"spec"`
	}
	patch.Spec.Replicas = nr
	data, err := json.Marshal(patch)
	utils.CheckError(err)
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

func (k *kubeClient) FindDeployedKubewareVersion(namespace, name string) (string, error) {
	kubename, err := utils.NormalizeName(name)
	if err != nil {
		return "", err
	}
	versions := map[string]string{}
	services, err := k.GetServicesForNamespace(namespace)
	if err != nil {
		return "", err
	}
	controllers, err := k.GetControllersForNamespace(namespace)
	if err != nil {
		return "", err
	}
	// Iterate over services and collect versions
	for _, s := range *services {
		n := s.GetKube()
		v := s.GetVersion()
		prev, found := versions[n]
		// Check if version already found
		if !found {
			versions[n] = v
		}
		// Check if a different version was already found
		if found && prev != v {
			return "", fmt.Errorf("found more than one version of the same Kubeware (%s.%s %s/%s)", namespace, kubename, prev, v)
		}
	}
	// Iterate over controllers and collect versions
	for _, c := range *controllers {
		n := c.GetKube()
		v := c.GetVersion()
		prev, found := versions[n]
		// Check if version already found
		if !found {
			versions[n] = v
		}
		// Check if a different version was already found
		if found && prev != v {
			return "", fmt.Errorf("found more than one version of the same Kubeware (%s.%s %s/%s)", namespace, kubename, prev, v)
		}
	}
	v, found := versions[kubename]
	if !found {
		return "", nil
	}
	return v, nil
}

func (k *kubeClient) IsServiceDeployed(namespace, svcName string) (bool, error) {
	_, err := k.getService(namespace, svcName)
	if err == ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (k *kubeClient) ReplaceService(namespace, svcName, svcJSON string) error {
	// Get currently deployed service
	glog.V(4).Infof("Getting currently deployed service")
	deployedSvcJSON, err := k.getService(namespace, svcName)
	if err != nil {
		return err
	}
	var deployedSvc models.Service
	err = json.Unmarshal([]byte(deployedSvcJSON), &deployedSvc)
	if err != nil {
		return err
	}
	// Unmarshal service to be modified
	glog.V(4).Infof("Unmarshaling service to be modified")
	var modifiedSvc map[string]interface{}
	err = json.Unmarshal([]byte(svcJSON), &modifiedSvc)
	if err != nil {
		return err
	}
	// Some fields must match or the update will be denied
	glog.V(4).Infof("Merging inmutables")
	err = mergeServiceInmutableFields(modifiedSvc, deployedSvc)
	if err != nil {
		return err
	}
	// For concerto-LB, we need to preserve the current nodePort, since this wouldnt result in a
	// LB update and the haproxy rules would still refer to the old one, raising a 503 http error on access
	glog.V(4).Infof("Preserving nodeports")
	err = preserveNodePorts(modifiedSvc, deployedSvc)
	if err != nil {
		return err
	}
	// marshal again
	modifiedJSON, err := json.Marshal(modifiedSvc)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("api/v1/namespaces/%s/services/%s", namespace, svcName)
	body, status, err := k.service.Put(path, []byte(modifiedJSON))
	if err != nil {
		return err
	}
	if status == 404 {
		return ErrNotFound
	}
	if status != 200 && status != 201 {
		return fmt.Errorf("wrong http status code: %v (body: %s)", status, body)
	}
	return nil
}

func (k *kubeClient) PatchService(namespace, svcName, svcJSON string) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/services/%s", namespace, svcName)
	resp, status, err := k.service.Patch(path, []byte(svcJSON))
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("wrong http status code: %v (%s)", status, resp)
	}
	return nil
}

func mergeServiceInmutableFields(modifiedSvc map[string]interface{}, deployedService models.Service) error {
	modifiedSvc["metadata"].(map[string]interface{})["resourceVersion"] = deployedService.Metadata.ResourceVersion
	modifiedSvc["spec"].(map[string]interface{})["clusterIP"] = deployedService.Spec.ClusterIP
	return nil
}

func preserveNodePorts(modifiedSvc map[string]interface{}, deployedService models.Service) error {
	newports := modifiedSvc["spec"].(map[string]interface{})["ports"].([]interface{})
	for _, oldport := range deployedService.Spec.Ports {
		for _, newport := range newports {
			mport := newport.(map[string]interface{})
			if int(mport["port"].(float64)) == oldport.Port {
				glog.V(4).Infof("Found port %v already associated with nodePort %v for service '%s'", oldport.Port, oldport.NodePort, deployedService.GetName())
				mport["nodePort"] = oldport.NodePort
			}
		}
	}
	return nil
}

func (k *kubeClient) ReplaceReplicationController(namespace, rcName, rcJSON string) error {
	path := fmt.Sprintf("api/v1/namespaces/%s/replicationcontrollers/%s", namespace, rcName)
	_, status, err := k.service.Put(path, []byte(rcJSON))
	if err != nil {
		return err
	}
	if status == 404 {
		return ErrNotFound
	}
	if status != 200 && status != 201 {
		return fmt.Errorf("wrong http status code: %v", status)
	}
	return nil
}
