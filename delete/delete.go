package delete

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/delete/strategies"
	"github.com/flexiant/kdeploy/models"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDelete implements 'delete' command
func CmdDelete(c *cli.Context) {
	log.Debugf("deleting : %s", os.Getenv("KDEPLOY_KUBEWARE"))
	var kubewareName string
	var kubewareVersion string
	var kubewareURL string
	var labelSelector string
	var err error

	namespace := os.Getenv("KDEPLOY_NAMESPACE")

	if govalidator.IsURL(os.Getenv("KDEPLOY_KUBEWARE")) {
		kubewareURL = os.Getenv("KDEPLOY_KUBEWARE")
		labelSelector, kubewareName, kubewareVersion = labelSelectorFromURL(kubewareURL)
	} else {
		// not URL so we will interpret it as a name
		kubewareName, err = utils.NormalizeName(os.Getenv("KDEPLOY_KUBEWARE"))
		utils.CheckError(err)
		labelSelector = labelSelectorFromName(kubewareName)
	}

	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)

	// get services which are currently deployed as part of the kube
	serviceList, err := kubernetes.GetServicesForNamespace(namespace, labelSelector)
	utils.CheckError(err)
	log.Debugf("Services: %v", serviceList)

	// get controllers which are currently deployed as part of the kube
	controllerList, err := kubernetes.GetControllersForNamespace(namespace, labelSelector)
	utils.CheckError(err)
	log.Debugf("Controllers: %v", controllerList)

	// If no resources found that means it's not deployed
	if len(*serviceList) == 0 || len(*controllerList) == 0 {
		var version string
		if kubewareVersion != "" {
			version = fmt.Sprintf(" (%s)", kubewareVersion)
		}
		log.Warnf("Could not delete kubeware '%s'%s since it is not currently deployed", kubewareName, version)
		return
	}

	// delete them
	ds := deletionStrategies.WaitZeroReplicasDeletionStrategy(kubernetes)
	err = ds.Delete(namespace, svcNames(serviceList), rcNames(controllerList))
	utils.CheckError(err)

	log.Infof("Kubeware '%s.%s' has been deleted", namespace, kubewareName)
}

func rcNames(rcl *[]models.ReplicaController) []string {
	names := []string{}
	for _, rc := range *rcl {
		names = append(names, rc.Metadata.Name)
	}
	return names
}

func svcNames(sl *[]models.Service) []string {
	names := []string{}
	for _, s := range *sl {
		names = append(names, s.Metadata.Name)
	}
	return names
}

func labelSelectorFromName(name string) string {
	return fmt.Sprintf("kubeware=%s", name)
}

func labelSelectorFromURL(kubewareURL string) (string, string, string) {
	localKubePath, err := webservice.FetchKubeFromURL(kubewareURL)
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	utils.CheckError(err)

	normalizedName, err := utils.NormalizeName(md.Name)
	utils.CheckError(err)
	labelSelector := fmt.Sprintf("kubeware=%s,kubeware-version=%s", normalizedName, md.Version)

	return labelSelector, md.Name, md.Version
}
