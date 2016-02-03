package delete

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/delete/strategies"
	"github.com/flexiant/kdeploy/models"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDelete implements 'delete' command
func CmdDelete(c *cli.Context) {
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	utils.CheckError(err)

	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)

	namespace := os.Getenv("KDEPLOY_NAMESPACE")
	labelSelector := fmt.Sprintf("kubeware=%s,kubeware-version=%s", md.Name, md.Version)

	// get services which are currently deployed as part of the kube
	serviceList, err := kubernetes.GetServicesForNamespace(namespace, labelSelector)
	utils.CheckError(err)
	log.Debugf("Services: %v", serviceList)

	// get controllers which are currently deployed as part of the kube
	controllerList, err := kubernetes.GetControllersForNamespace(namespace, labelSelector)
	utils.CheckError(err)
	log.Debugf("Controllers: %v", controllerList)

	// delete them
	ds := deletionStrategies.WaitZeroReplicasDeletionStrategy(kubernetes)
	err = ds.Delete(namespace, svcNames(serviceList), rcNames(controllerList))
	utils.CheckError(err)

	fmt.Printf("Kubeware %s from %s has been deleted", md.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}

func rcNames(rcl *models.ControllerList) []string {
	names := []string{}
	for _, rc := range rcl.Items {
		fmt.Printf("rc: %s\n", rc.Metadata.Name)
		names = append(names, rc.Metadata.Name)
	}
	return names
}

func svcNames(sl *models.ServiceList) []string {
	names := []string{}
	for _, s := range sl.Items {
		fmt.Printf("s: %s\n", s.Metadata.Name)
		names = append(names, s.Metadata.Name)
	}
	return names
}
