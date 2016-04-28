package deploy

import (
	"fmt"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/fetchers"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDeploy implements the 'deploy' command
func CmdDeploy(c *cli.Context) {
	utils.CheckRequiredFlags(c, []string{"kubeware"})

	var kubeware string = os.Getenv("KDEPLOY_KUBEWARE")
	var localKubePath string
	var err error

	if !govalidator.IsURL(kubeware) {
		log.Fatal(fmt.Errorf("Not a valid URL: '%s'", kubeware))
	}
	kubewareURL, err := url.Parse(kubeware)
	if err != nil {
		log.Fatal(fmt.Errorf("Error parsing URL: '%s'", kubeware))
	}

	for _, fetcher := range fetchers.AllFetchers {
		if fetcher.CanHandle(kubewareURL) {
			localKubePath, err = fetcher.Fetch(kubewareURL)
			utils.CheckError(err)
		}
	}

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
	// check if kubeware already exists
	log.Debugf("Checking if already deployed")
	deployedVersion, err := kubernetes.FindDeployedKubewareVersion(os.Getenv("KDEPLOY_NAMESPACE"), metadata.Name)
	utils.CheckError(err)
	if deployedVersion != "" {
		log.Errorf("Can not deploy '%s' since version '%s' is already deployed", metadata.Name, deployedVersion)
		return
	}
	// create each of the services
	log.Debugf("Creating services")
	err = kubernetes.CreateServices(utils.Values(servicesSpecs))
	utils.CheckError(err)
	// create each of the controllers
	log.Debugf("Creating controllers")
	err = kubernetes.CreateReplicaControllers(utils.Values(controllersSpecs))
	utils.CheckError(err)

	log.Infof("Kubeware %s from %s has been deployed", metadata.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}

// func isLocalURL(kube string) bool  {
//   kubewareURL, err := url.Parse(kube)
// 	return (err == nil && kubewareURL.Scheme == "file")
// }
//
// func extractAbsolutePath(kube string) bool  {
//   kubewareURL, err := url.Parse(kube)
// 	if err == nil && kubewareURL.Scheme == "file")
// }
