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

func Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "attribute, a",
			Usage:  "Attribute List",
			Value:  "./examples/attributes.json",
			EnvVar: "KDEPLOY_ATTRIBUTE",
		},
		cli.StringFlag{
			Name:   "kubeware, k",
			Usage:  "Kubeware path",
			Value:  "https://github.com/flexiant/kubeware-guestbook",
			EnvVar: "KDEPLOY_KUBEWARE",
		},
		cli.BoolFlag{
			Name:   "dry-run, d",
			Usage:  "Dry Run of Deploy used for debuging options",
			EnvVar: "KDEPLOY_DRYRUN",
		},
	}
}

func PrepareFlags(c *cli.Context) error {
	if c.String("attribute") != "" {
		os.Setenv("KDEPLOY_ATTRIBUTE", c.String("attribute"))
	}

	if c.String("kubeware") != "" {
		os.Setenv("KDEPLOY_KUBEWARE", c.String("kubeware"))
	}

	if c.Bool("dry-run") {
		os.Setenv("KDEPLOY_DRYRUN", "1")
	}

	return nil
}

func CmdDeploy(c *cli.Context) {

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
	// create each of the services
	log.Debugf("Creating services")
	err = createServices(servicesSpecs)
	utils.CheckError(err)
	// create each of the controllers
	log.Debugf("Creating controllers")
	err = createControllers(controllersSpecs)
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

func createServices(svcSpecs []string) error {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return err
	}
	for _, spec := range svcSpecs {
		_, err = kube.CreateService("default", []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating services: %v", err)
		}
	}
	return nil
}

func createControllers(rcSpecs []string) error {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return fmt.Errorf("error creating kube client: %v", err)
	}
	for _, spec := range rcSpecs {
		_, err = kube.CreateReplicaController("default", []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating replication controllers: %v", err)
		}
	}
	return nil
}
