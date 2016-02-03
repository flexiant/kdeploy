package delete

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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

	labelSelector := map[string]string{
		"kubeware": md.Name,
	}

	// get services which are currently deployed as part of the kube
	serviceList, err := kubernetes.GetServices(labelSelector)

	utils.CheckError(err)
	log.Debugf("Services: %v", serviceList)

	// get controllers which are currently deployed as part of the kube
	controllerList, err := kubernetes.GetControllers(labelSelector)
	utils.CheckError(err)
	log.Debugf("Controllers: %v", controllerList)

	// delete them
	err = kubernetes.DeleteServices(os.Getenv("KDEPLOY_NAMESPACE"), serviceList)
	utils.CheckError(err)

	// delete them
	err = kubernetes.DeleteControllers(os.Getenv("KDEPLOY_NAMESPACE"), controllerList)
	utils.CheckError(err)

	fmt.Printf("Kubeware %s from %s has been deleted", md.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}
