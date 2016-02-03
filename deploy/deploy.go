package deploy

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

func CmdDeploy(c *cli.Context) {
	// utils.CheckRequiredFlags(c, []string{"attribute", "kubeware", "namespace"})
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	defaults, err := md.AttributeDefaults()
	utils.CheckError(err)
	// build attributes merging "role list" to defaults
	log.Debugf("Building attributes")
	attributes := buildAttributes(c.String("attribute"), defaults)
	// get list of services and parse each one
	log.Debugf("Parsing services")
	servicesSpecs, err := md.ParseServices(attributes)
	utils.CheckError(err)
	// get list of replica controllers and parse each one
	log.Debugf("Parsing controllers")
	controllersSpecs, err := md.ParseControllers(attributes)
	utils.CheckError(err)
	// creates Kubernetes client
	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)
	// create each of the services
	log.Debugf("Creating services")
	err = kubernetes.CreateServices(servicesSpecs)
	utils.CheckError(err)
	// create each of the controllers
	log.Debugf("Creating controllers")
	err = kubernetes.CreateReplicaControllers(controllersSpecs)
	utils.CheckError(err)

	fmt.Printf("Kubeware %s from %s has been deployed", md.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}

func buildAttributes(filePath string, defaults digger.Digger) digger.Digger {
	roleList, err := ioutil.ReadFile(filePath)
	utils.CheckError(err)

	roleListDigger, err := digger.NewJSONDigger([]byte(roleList))
	utils.CheckError(err)

	attributes, err := digger.NewMultiDigger(
		roleListDigger,
		defaults,
	)
	utils.CheckError(err)

	return attributes
}
