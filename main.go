package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/kubeclient"
	"github.com/flexiant/kdeploy/metadata"
	"github.com/flexiant/kdeploy/utils"
)

func cmdNotFound(c *cli.Context, command string) {
	log.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
}
func prepareFlags(c *cli.Context) error {

	if c.Bool("debug") {
		os.Setenv("DEBUG", "1")
		log.SetOutput(os.Stderr)
		log.SetLevel(log.DebugLevel)
	}
	err := utils.InitializeConfig(c)
	if err != nil {
		log.Errorf("Error reading Kdeploy configuration: %s", err)
		return err
	}

	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "kdeploy"
	app.Author = "Concerto Contributors"
	app.Email = "https://github.com/flexiant/kdeploy"

	app.Usage = "Deploys Kubeware in kubernetes clusters"
	app.Version = utils.VERSION
	app.CommandNotFound = cmdNotFound
	app.Before = prepareFlags

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},

		cli.StringFlag{
			EnvVar: "KUBERNETES_CA_CERT",
			Name:   "ca-cert",
			Usage:  "CA to verify remote connections",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_CLIENT_CERT",
			Name:   "client-cert",
			Usage:  "Client cert to use for Kubernetes",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_CLIENT_KEY",
			Name:   "client-key",
			Usage:  "Private key used in client Kubernetes auth",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_ENDPOINT",
			Name:   "kubernetes-endpoint",
			Usage:  "Kubernetes Endpoint",
		},
		cli.StringFlag{
			EnvVar: "KDEPLOY_CONFIG",
			Name:   "kdeploy-config",
			Usage:  "Kdeploy Config File",
		},
	}
	kubewareDefaulPath, _ := filepath.Abs("./examples/guestbook/")
	app.Commands = []cli.Command{
		{
			Name:   "deploy",
			Usage:  "Deploys a Kubeware",
			Action: cmdDeploy,
			Flags: []cli.Flag{
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
			},
		},
		{
			Name:  "destroy",
			Usage: "Destroys a Kubeware",
		},
		{
			Name:  "list",
			Usage: "List's Kubewares deployed",
		},
	}

	app.Run(os.Args)
}

func cmdDeploy(c *cli.Context) {

	md := metadata.ParseMetadata(c.String("kubeware"))
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
	kube, _ := kubeclient.NewKubeClient()
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
	kube, err := kubeclient.NewKubeClient()
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
	kube, err := kubeclient.NewKubeClient()
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
