package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/flexiant/kdeploy/deploy"
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

	app.Commands = []cli.Command{
		{
			Name:   "deploy",
			Usage:  "Deploys a Kubeware",
			Before: deploy.PrepareFlags,
			Action: deploy.CmdDeploy,
			Flags:  deploy.Flags(),
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
