package deploy

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/fetchers"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDeploy implements the 'deploy' command
func CmdDeploy(c *cli.Context) {
	utils.CheckRequiredFlags(c, []string{"kubeware"})

	var kubeware string
	var localKubePath string
	var err error

	kubeware = os.Getenv("KDEPLOY_KUBEWARE")
	localKubePath, err = fetchers.Fetch(kubeware)
	if err != nil {
		glog.Fatal(fmt.Errorf("Could not fetch kubeware: '%s' (%v)", kubeware, err))
	}

	glog.V(4).Infof("Going to parse kubeware in %s", localKubePath)

	metadata := template.ParseMetadata(localKubePath)
	defaults, err := metadata.AttributeDefaults()
	utils.CheckError(err)
	// build attributes merging "role list" to defaults
	glog.V(4).Infof("Building attributes")
	attributes := template.BuildAttributes(c.String("attribute"), defaults)
	// get list of services and parse each one
	glog.V(4).Infof("Parsing services")
	servicesSpecs, err := metadata.ParseServices(attributes)
	utils.CheckError(err)
	// get list of replica controllers and parse each one
	glog.V(4).Infof("Parsing controllers")
	controllersSpecs, err := metadata.ParseControllers(attributes)
	utils.CheckError(err)
	// creates Kubernetes client
	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)
	// check if kubeware already exists
	glog.V(4).Infof("Checking if already deployed")
	deployedVersion, err := kubernetes.FindDeployedKubewareVersion(os.Getenv("KDEPLOY_NAMESPACE"), metadata.Name)
	utils.CheckError(err)
	if deployedVersion != "" {
		glog.Errorf("Can not deploy '%s' since version '%s' is already deployed", metadata.Name, deployedVersion)
		return
	}
	// create each of the services
	glog.V(4).Infof("Creating services")
	err = kubernetes.CreateServices(utils.Values(servicesSpecs))
	utils.CheckError(err)
	// create each of the controllers
	glog.V(4).Infof("Creating controllers")
	err = kubernetes.CreateReplicaControllers(utils.Values(controllersSpecs))
	utils.CheckError(err)

	glog.Infof("Kubeware %s from %s has been deployed", metadata.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}
