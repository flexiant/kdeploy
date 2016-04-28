package show

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/fetchers"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	gyml "github.com/ghodss/yaml"
)

// CmdShow implements the 'show' command
func CmdShow(c *cli.Context) {
	utils.CheckRequiredFlags(c, []string{"kubeware"})

	kubeware := os.Getenv("KDEPLOY_KUBEWARE")
	localKubePath, err := fetchers.Fetch(kubeware)
	if err != nil {
		log.Fatal(fmt.Errorf("Could not fetch kubeware: '%s' (%v)", kubeware, err))
	}

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	metadata := template.ParseMetadata(localKubePath)
	defaults, err := metadata.AttributeDefaults()
	utils.CheckError(err)

	log.Debugf("Building attributes")
	attributes := template.BuildAttributes(os.Getenv("KDEPLOY_ATTRIBUTE"), defaults)

	log.Debugf("Parsing services")
	servicesSpecs, err := metadata.ParseServices(attributes)
	utils.CheckError(err)

	log.Debugf("Parsing controllers")
	controllersSpecs, err := metadata.ParseControllers(attributes)
	utils.CheckError(err)

	// print resolved resources
	for _, s := range servicesSpecs {
		y, err := gyml.JSONToYAML([]byte(s))
		utils.CheckError(err)
		fmt.Println(string(y))
	}
	for _, c := range controllersSpecs {
		y, err := gyml.JSONToYAML([]byte(c))
		utils.CheckError(err)
		fmt.Println(string(y))
	}
}
