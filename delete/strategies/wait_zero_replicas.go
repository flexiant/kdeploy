package deletionStrategies

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/kdeploy/webservice"
)

// WaitZeroReplicas is a DeletionStrategy that deletes all the pods associated with the replication
// controllers, by setting the number of replicas to zero and waiting for them to be gone
type waitZeroReplicas struct {
	kubeClient webservice.KubeClient
}

// WaitZeroReplicasDeletionStrategy builds a new instance of 'waitZeroReplicas' deletion strategy
func WaitZeroReplicasDeletionStrategy(k webservice.KubeClient) DeletionStrategy {
	return &waitZeroReplicas{k}
}

func (zr *waitZeroReplicas) DeleteService(namespace string, name string) error {
	return zr.kubeClient.DeleteService(namespace, name)
}

func (zr *waitZeroReplicas) DeleteReplicationController(namespace string, name string) error {
	// set replicas number to zero
	err := zr.kubeClient.SetSpecReplicas(namespace, name, 0)
	if err != nil {
		return fmt.Errorf("could not set replicas to zero: %v", err)
	}
	// wait for pods to actually be deleted
	var n uint
	for n = 1; n > 0; time.Sleep(1 * time.Second) {
		log.Debugf("Waiting for all pods to be gone (%s.%s %v remaining)", namespace, name, n)
		n, err = zr.kubeClient.GetStatusReplicas(namespace, name)
		if err != nil {
			return err
		}
	}
	// then delete the RC
	return zr.kubeClient.DeleteReplicationController(namespace, name)
}
