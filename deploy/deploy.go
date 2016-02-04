package deploy

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDeploy implements the 'deploy' command
func CmdDeploy(c *cli.Context) {
	// utils.CheckRequiredFlags(c, []string{"attribute", "kubeware", "namespace"})
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	metadata := template.ParseMetadata(localKubePath)
	defaults, err := metadata.AttributeDefaults()
	utils.CheckError(err)
	// build attributes merging "role list" to defaults
	log.Debugf("Building attributes")
	attributes := template.BuildAttributes(c.String("attribute"), defaults)
	// get list of services and parse each one
	log.Debugf("Parsing services")
	servicesSpecs, err := metadata.ParseServices(attributes)
	utils.CheckError(err)
	// get list of replica controllers and parse each one
	log.Debugf("Parsing controllers")
	controllersSpecs, err := metadata.ParseControllers(attributes)
	utils.CheckError(err)
	// creates Kubernetes client
	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)
	// create each of the services
	log.Debugf("Creating services")
	err = kubernetes.CreateServices(utils.Values(servicesSpecs))
	utils.CheckError(err)
	// create each of the controllers
	log.Debugf("Creating controllers")
	err = kubernetes.CreateReplicaControllers(utils.Values(controllersSpecs))
	utils.CheckError(err)

	fmt.Printf("Kubeware %s from %s has been deployed", metadata.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}
