package deploy

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

func Flags() []cli.Flag {
	kubewareDefaulPath, _ := filepath.Abs("./examples/guestbook/")
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
			Value:  kubewareDefaulPath,
			EnvVar: "KDEPLOY_KUBEWARE",
		},
	}
}

func CmdDeploy(c *cli.Context) {

	md := template.ParseMetadata(c.String("kubeware"))
	defaults, err := md.AttributeDefaults()
	utils.CheckError(err)
	// build attributes merging "role list" to defaults
	attributes := buildAttributes(c.String("attribute"), defaults)
	// get list of services and parse each one
	servicesSpecs, err := md.ParseServices(attributes)
	utils.CheckError(err)
	// get list of replica controllers and parse each one
	controllersSpecs, err := md.ParseControllers(attributes)
	utils.CheckError(err)

	// get services just to check API availability
	// getServices()

	// create each of the services
	err = createServices(servicesSpecs)
	utils.CheckError(err)
	// create each of the controllers
	err = createControllers(controllersSpecs)
	utils.CheckError(err)
}

func getServices() {
	kube, _ := webservice.NewKubeClient()
	services, _ := kube.GetServices()
	fmt.Println("services: ")
	fmt.Println(services)
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
		return fmt.Errorf("error creating kube client: %v", err)
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
