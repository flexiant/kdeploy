package deploy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

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
	if c.IsSet("attribute") {
		os.Setenv("KDEPLOY_ATTRIBUTE", c.String("attribute"))
	}

	if c.IsSet("kubeware") {
		os.Setenv("KDEPLOY_KUBEWARE", c.String("kubeware"))
	}

	if c.Bool("dry-run") {
		os.Setenv("KDEPLOY_DRYRUN", "1")
	}

	return nil
}

func CmdDeploy(c *cli.Context) {
	kubewareUrl, _ := url.Parse(os.Getenv("KDEPLOY_KUBEWARE"))

	if kubewareUrl != nil {
		if kubewareUrl.Host == "github.com" {
			path := strings.Split(kubewareUrl.Path, "/")
			kubewareName := path[2]
			newPath := []string{""}
			newPath = append(newPath, path[1], path[2], "archive", "master.zip")

			kubewareUrl.Path = strings.Join(newPath, "/")

			client, err := webservice.NewSimpleWebClient(kubewareUrl.String())
			utils.CheckError(err)

			tmpDir, err := ioutil.TempDir("", "kdeploy")
			utils.CheckError(err)

			zipFileLocation, err := client.GetFile(kubewareUrl.Path, tmpDir)
			utils.CheckError(err)

			err = utils.Unzip(zipFileLocation, tmpDir)
			utils.CheckError(err)

			os.Setenv("KDEPLOY_KUBEWARE", fmt.Sprintf("%s/%s-master/", tmpDir, kubewareName))

		} else {
			utils.CheckError(errors.New("We currently only support Github urls"))
		}
	}

	log.Debugf("Going to parse kubeware in %s", os.Getenv("KDEPLOY_KUBEWARE"))

	md := template.ParseMetadata(os.Getenv("KDEPLOY_KUBEWARE"))
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
